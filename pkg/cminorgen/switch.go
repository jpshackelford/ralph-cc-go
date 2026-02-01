// Package cminorgen implements the Cminorgen pass: Csharpminor → Cminor
// This file handles switch statement simplification.
package cminorgen

import (
	"sort"

	"github.com/raymyers/ralph-cc/pkg/cminor"
)

// SwitchStrategy represents the strategy for implementing a switch
type SwitchStrategy int

const (
	StrategyLinear    SwitchStrategy = iota // Linear if-cascade (small switches)
	StrategyBinary                          // Binary search (sparse cases)
	StrategyJumpTable                       // Jump table (dense cases)
)

// SwitchCase represents a case for transformation (sorted by value)
type SwitchCase struct {
	Value int64
	Exit  int // Exit depth for this case
}

// SwitchAnalysis holds the analysis of a switch statement
type SwitchAnalysis struct {
	Cases    []SwitchCase   // Cases sorted by value
	Default  int            // Default exit depth
	IsLong   bool           // True for long switch
	Strategy SwitchStrategy // Selected implementation strategy
	MinVal   int64          // Minimum case value
	MaxVal   int64          // Maximum case value
	Density  float64        // Case density (numCases / range)
}

// AnalyzeSwitch analyzes a switch statement and determines the best strategy.
func AnalyzeSwitch(sw cminor.Sswitch, defaultExit int) *SwitchAnalysis {
	analysis := &SwitchAnalysis{
		Cases:   make([]SwitchCase, len(sw.Cases)),
		Default: defaultExit,
		IsLong:  sw.IsLong,
	}

	// Extract and sort cases by value
	for i, c := range sw.Cases {
		// The body of each case is typically a Sexit or statement leading to exit
		exitDepth := extractExitDepth(c.Body)
		analysis.Cases[i] = SwitchCase{
			Value: c.Value,
			Exit:  exitDepth,
		}
	}
	sort.Slice(analysis.Cases, func(i, j int) bool {
		return analysis.Cases[i].Value < analysis.Cases[j].Value
	})

	// Compute statistics
	if len(analysis.Cases) > 0 {
		analysis.MinVal = analysis.Cases[0].Value
		analysis.MaxVal = analysis.Cases[len(analysis.Cases)-1].Value
		rangeSize := analysis.MaxVal - analysis.MinVal + 1
		if rangeSize > 0 {
			analysis.Density = float64(len(analysis.Cases)) / float64(rangeSize)
		}
	}

	// Select strategy based on case characteristics
	analysis.Strategy = selectStrategy(len(analysis.Cases), analysis.Density)

	return analysis
}

// selectStrategy chooses the implementation strategy based on case count and density.
// Strategy selection follows CompCert's approach in Cminorgen.v:
// - Small switches (<=4 cases): linear if-cascade
// - Dense switches (density >= 0.5): jump table
// - Otherwise: binary search
func selectStrategy(numCases int, density float64) SwitchStrategy {
	const (
		linearThreshold = 4
		denseThreshold  = 0.5
	)

	if numCases <= linearThreshold {
		return StrategyLinear
	}
	if density >= denseThreshold {
		return StrategyJumpTable
	}
	return StrategyBinary
}

// extractExitDepth extracts the exit depth from a case body.
// Returns 0 if no Sexit found (shouldn't happen for well-formed switches).
func extractExitDepth(body cminor.Stmt) int {
	switch s := body.(type) {
	case cminor.Sexit:
		return s.N
	case cminor.Sseq:
		// Check last statement in sequence
		return extractExitDepth(s.Second)
	default:
		return 0 // Default if no explicit exit
	}
}

// TransformSwitchLinear transforms a switch using linear if-cascade.
// Each case becomes: if (expr == value) exit(n);
func TransformSwitchLinear(expr cminor.Expr, analysis *SwitchAnalysis) cminor.Stmt {
	if len(analysis.Cases) == 0 {
		return cminor.Sexit{N: analysis.Default}
	}

	// Build from last case backwards to create proper nesting
	var result cminor.Stmt = cminor.Sexit{N: analysis.Default}

	for i := len(analysis.Cases) - 1; i >= 0; i-- {
		c := analysis.Cases[i]
		cond := makeComparison(expr, c.Value, analysis.IsLong)
		result = cminor.Sifthenelse{
			Cond: cond,
			Then: cminor.Sexit{N: c.Exit},
			Else: result,
		}
	}

	return result
}

// TransformSwitchBinary transforms a switch using binary search.
// Recursively splits the case space and generates comparisons.
func TransformSwitchBinary(expr cminor.Expr, analysis *SwitchAnalysis) cminor.Stmt {
	if len(analysis.Cases) == 0 {
		return cminor.Sexit{N: analysis.Default}
	}

	return transformBinaryRecursive(expr, analysis.Cases, analysis.Default, analysis.IsLong)
}

// transformBinaryRecursive implements binary search for switch cases.
func transformBinaryRecursive(expr cminor.Expr, cases []SwitchCase, defaultExit int, isLong bool) cminor.Stmt {
	n := len(cases)
	if n == 0 {
		return cminor.Sexit{N: defaultExit}
	}
	if n == 1 {
		// Single case: if (expr == value) exit(n) else default
		c := cases[0]
		return cminor.Sifthenelse{
			Cond: makeComparison(expr, c.Value, isLong),
			Then: cminor.Sexit{N: c.Exit},
			Else: cminor.Sexit{N: defaultExit},
		}
	}

	// Split at midpoint
	mid := n / 2
	pivot := cases[mid].Value

	// Left branch: cases < pivot
	// Right branch: cases >= pivot
	leftCases := cases[:mid]
	rightCases := cases[mid:]

	// Generate: if (expr < pivot) left else right
	cond := makeLessThan(expr, pivot, isLong)
	return cminor.Sifthenelse{
		Cond: cond,
		Then: transformBinaryRecursive(expr, leftCases, defaultExit, isLong),
		Else: transformBinaryRecursive(expr, rightCases, defaultExit, isLong),
	}
}

// TransformSwitchJumpTable transforms a switch using a jump table.
// Generates: switch(expr - minval) { case 0: ..., case 1: ..., ..., default: }
// Gaps in the table fall through to default.
func TransformSwitchJumpTable(expr cminor.Expr, analysis *SwitchAnalysis) cminor.Stmt {
	if len(analysis.Cases) == 0 {
		return cminor.Sexit{N: analysis.Default}
	}

	// Normalize: subtract minimum to get zero-based index
	// new_expr = expr - minval
	var normalizedExpr cminor.Expr
	if analysis.MinVal == 0 {
		normalizedExpr = expr
	} else {
		normalizedExpr = cminor.Ebinop{
			Op:    subtractOp(analysis.IsLong),
			Left:  expr,
			Right: makeConstant(analysis.MinVal, analysis.IsLong),
		}
	}

	// Build case list with gaps filled by default
	rangeSize := analysis.MaxVal - analysis.MinVal + 1
	caseMap := make(map[int64]int)
	for _, c := range analysis.Cases {
		caseMap[c.Value] = c.Exit
	}

	newCases := make([]cminor.SwitchCase, 0, rangeSize)
	for i := int64(0); i < rangeSize; i++ {
		actualValue := analysis.MinVal + i
		if exit, ok := caseMap[actualValue]; ok {
			newCases = append(newCases, cminor.SwitchCase{
				Value: i, // Zero-based index
				Body:  cminor.Sexit{N: exit},
			})
		}
		// Gaps go to default, which is handled by the default case
	}

	return cminor.Sswitch{
		IsLong:  analysis.IsLong,
		Expr:    normalizedExpr,
		Cases:   newCases,
		Default: cminor.Sexit{N: analysis.Default},
	}
}

// TransformSwitch transforms a switch statement using the best strategy.
func TransformSwitch(sw cminor.Sswitch, defaultExit int) cminor.Stmt {
	analysis := AnalyzeSwitch(sw, defaultExit)

	switch analysis.Strategy {
	case StrategyLinear:
		return TransformSwitchLinear(sw.Expr, analysis)
	case StrategyBinary:
		return TransformSwitchBinary(sw.Expr, analysis)
	case StrategyJumpTable:
		return TransformSwitchJumpTable(sw.Expr, analysis)
	default:
		// Fallback to linear
		return TransformSwitchLinear(sw.Expr, analysis)
	}
}

// makeComparison generates an equality comparison: expr == value
func makeComparison(expr cminor.Expr, value int64, isLong bool) cminor.Expr {
	op := cminor.Ocmp
	if isLong {
		op = cminor.Ocmpl
	}
	return cminor.Ecmp{
		Op:    op,
		Cmp:   cminor.Ceq,
		Left:  expr,
		Right: makeConstant(value, isLong),
	}
}

// makeLessThan generates a less-than comparison: expr < value
func makeLessThan(expr cminor.Expr, value int64, isLong bool) cminor.Expr {
	op := cminor.Ocmp
	if isLong {
		op = cminor.Ocmpl
	}
	return cminor.Ecmp{
		Op:    op,
		Cmp:   cminor.Clt,
		Left:  expr,
		Right: makeConstant(value, isLong),
	}
}

// makeConstant creates an integer or long constant
func makeConstant(value int64, isLong bool) cminor.Expr {
	if isLong {
		return cminor.Econst{Const: cminor.Olongconst{Value: value}}
	}
	return cminor.Econst{Const: cminor.Ointconst{Value: int32(value)}}
}

// subtractOp returns the appropriate subtraction operator
func subtractOp(isLong bool) cminor.BinaryOp {
	if isLong {
		return cminor.Osubl
	}
	return cminor.Osub
}
