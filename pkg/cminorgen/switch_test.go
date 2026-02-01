package cminorgen

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/cminor"
)

func TestSelectStrategy(t *testing.T) {
	tests := []struct {
		name     string
		numCases int
		density  float64
		want     SwitchStrategy
	}{
		{"empty", 0, 0.0, StrategyLinear},
		{"single case", 1, 1.0, StrategyLinear},
		{"small switch", 4, 0.2, StrategyLinear},
		{"dense switch", 10, 0.8, StrategyJumpTable},
		{"sparse switch", 10, 0.1, StrategyBinary},
		{"borderline dense", 10, 0.5, StrategyJumpTable},
		{"borderline sparse", 10, 0.49, StrategyBinary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectStrategy(tt.numCases, tt.density)
			if got != tt.want {
				t.Errorf("selectStrategy(%d, %f) = %v, want %v",
					tt.numCases, tt.density, got, tt.want)
			}
		})
	}
}

func TestAnalyzeSwitch(t *testing.T) {
	tests := []struct {
		name       string
		sw         cminor.Sswitch
		defaultExt int
		wantMin    int64
		wantMax    int64
		wantStrat  SwitchStrategy
	}{
		{
			name: "empty switch",
			sw: cminor.Sswitch{
				Cases:   nil,
				Default: cminor.Sexit{N: 0},
			},
			defaultExt: 0,
			wantStrat:  StrategyLinear,
		},
		{
			name: "small int switch",
			sw: cminor.Sswitch{
				Expr:   cminor.Evar{Name: "x"},
				IsLong: false,
				Cases: []cminor.SwitchCase{
					{Value: 1, Body: cminor.Sexit{N: 1}},
					{Value: 2, Body: cminor.Sexit{N: 2}},
				},
				Default: cminor.Sexit{N: 0},
			},
			defaultExt: 0,
			wantMin:    1,
			wantMax:    2,
			wantStrat:  StrategyLinear, // only 2 cases
		},
		{
			name: "dense switch",
			sw: cminor.Sswitch{
				Expr:   cminor.Evar{Name: "x"},
				IsLong: false,
				Cases: []cminor.SwitchCase{
					{Value: 0, Body: cminor.Sexit{N: 1}},
					{Value: 1, Body: cminor.Sexit{N: 2}},
					{Value: 2, Body: cminor.Sexit{N: 3}},
					{Value: 3, Body: cminor.Sexit{N: 4}},
					{Value: 4, Body: cminor.Sexit{N: 5}},
					{Value: 5, Body: cminor.Sexit{N: 6}},
				},
				Default: cminor.Sexit{N: 0},
			},
			defaultExt: 0,
			wantMin:    0,
			wantMax:    5,
			wantStrat:  StrategyJumpTable, // density = 1.0
		},
		{
			name: "sparse switch",
			sw: cminor.Sswitch{
				Expr:   cminor.Evar{Name: "x"},
				IsLong: true,
				Cases: []cminor.SwitchCase{
					{Value: 0, Body: cminor.Sexit{N: 1}},
					{Value: 100, Body: cminor.Sexit{N: 2}},
					{Value: 200, Body: cminor.Sexit{N: 3}},
					{Value: 300, Body: cminor.Sexit{N: 4}},
					{Value: 400, Body: cminor.Sexit{N: 5}},
				},
				Default: cminor.Sexit{N: 0},
			},
			defaultExt: 0,
			wantMin:    0,
			wantMax:    400,
			wantStrat:  StrategyBinary, // density ~= 0.0125
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := AnalyzeSwitch(tt.sw, tt.defaultExt)

			if analysis.Strategy != tt.wantStrat {
				t.Errorf("strategy = %v, want %v", analysis.Strategy, tt.wantStrat)
			}

			if len(tt.sw.Cases) > 0 {
				if analysis.MinVal != tt.wantMin {
					t.Errorf("MinVal = %d, want %d", analysis.MinVal, tt.wantMin)
				}
				if analysis.MaxVal != tt.wantMax {
					t.Errorf("MaxVal = %d, want %d", analysis.MaxVal, tt.wantMax)
				}
			}
		})
	}
}

func TestExtractExitDepth(t *testing.T) {
	tests := []struct {
		name string
		body cminor.Stmt
		want int
	}{
		{
			name: "simple exit",
			body: cminor.Sexit{N: 5},
			want: 5,
		},
		{
			name: "exit in sequence",
			body: cminor.Sseq{
				First:  cminor.Sskip{},
				Second: cminor.Sexit{N: 3},
			},
			want: 3,
		},
		{
			name: "no exit",
			body: cminor.Sskip{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractExitDepth(tt.body)
			if got != tt.want {
				t.Errorf("extractExitDepth() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTransformSwitchLinear(t *testing.T) {
	expr := cminor.Evar{Name: "x"}
	analysis := &SwitchAnalysis{
		Cases: []SwitchCase{
			{Value: 1, Exit: 1},
			{Value: 2, Exit: 2},
		},
		Default: 0,
		IsLong:  false,
	}

	result := TransformSwitchLinear(expr, analysis)

	// Should be: if (x == 1) exit(1) else if (x == 2) exit(2) else exit(0)
	switch s := result.(type) {
	case cminor.Sifthenelse:
		// First condition: x == 1
		if _, ok := s.Then.(cminor.Sexit); !ok {
			t.Error("expected Then to be Sexit")
		}
		// Else should be another if-then-else
		if _, ok := s.Else.(cminor.Sifthenelse); !ok {
			t.Errorf("expected Else to be Sifthenelse, got %T", s.Else)
		}
	default:
		t.Errorf("expected Sifthenelse, got %T", result)
	}
}

func TestTransformSwitchBinary(t *testing.T) {
	expr := cminor.Evar{Name: "x"}
	analysis := &SwitchAnalysis{
		Cases: []SwitchCase{
			{Value: 1, Exit: 1},
			{Value: 5, Exit: 2},
			{Value: 10, Exit: 3},
			{Value: 15, Exit: 4},
			{Value: 20, Exit: 5},
		},
		Default: 0,
		IsLong:  false,
	}

	result := TransformSwitchBinary(expr, analysis)

	// Should generate a binary search tree
	if _, ok := result.(cminor.Sifthenelse); !ok {
		t.Errorf("expected Sifthenelse, got %T", result)
	}
}

func TestTransformSwitchJumpTable(t *testing.T) {
	expr := cminor.Evar{Name: "x"}
	analysis := &SwitchAnalysis{
		Cases: []SwitchCase{
			{Value: 0, Exit: 1},
			{Value: 1, Exit: 2},
			{Value: 2, Exit: 3},
		},
		Default: 0,
		IsLong:  false,
		MinVal:  0,
		MaxVal:  2,
	}

	result := TransformSwitchJumpTable(expr, analysis)

	// Should return a normalized switch statement
	switch s := result.(type) {
	case cminor.Sswitch:
		if len(s.Cases) != 3 {
			t.Errorf("expected 3 cases, got %d", len(s.Cases))
		}
		// Check zero-based indices
		for i, c := range s.Cases {
			if c.Value != int64(i) {
				t.Errorf("case %d has value %d, want %d", i, c.Value, i)
			}
		}
	default:
		t.Errorf("expected Sswitch, got %T", result)
	}
}

func TestTransformSwitchJumpTableWithOffset(t *testing.T) {
	expr := cminor.Evar{Name: "x"}
	analysis := &SwitchAnalysis{
		Cases: []SwitchCase{
			{Value: 10, Exit: 1},
			{Value: 11, Exit: 2},
			{Value: 12, Exit: 3},
		},
		Default: 0,
		IsLong:  false,
		MinVal:  10,
		MaxVal:  12,
	}

	result := TransformSwitchJumpTable(expr, analysis)

	switch s := result.(type) {
	case cminor.Sswitch:
		// Expr should be x - 10
		if binop, ok := s.Expr.(cminor.Ebinop); ok {
			if binop.Op != cminor.Osub {
				t.Errorf("expected Osub, got %v", binop.Op)
			}
		} else {
			t.Errorf("expected Ebinop for normalized expression, got %T", s.Expr)
		}

		// Cases should be 0, 1, 2
		for i, c := range s.Cases {
			if c.Value != int64(i) {
				t.Errorf("case %d has value %d, want %d", i, c.Value, i)
			}
		}
	default:
		t.Errorf("expected Sswitch, got %T", result)
	}
}

func TestTransformSwitchEmpty(t *testing.T) {
	sw := cminor.Sswitch{
		Expr:    cminor.Evar{Name: "x"},
		Cases:   nil,
		Default: cminor.Sexit{N: 5},
	}

	result := TransformSwitch(sw, 5)

	if exit, ok := result.(cminor.Sexit); ok {
		if exit.N != 5 {
			t.Errorf("expected exit depth 5, got %d", exit.N)
		}
	} else {
		t.Errorf("expected Sexit for empty switch, got %T", result)
	}
}

func TestMakeComparison(t *testing.T) {
	tests := []struct {
		name   string
		value  int64
		isLong bool
		wantOp cminor.BinaryOp
	}{
		{"int comparison", 42, false, cminor.Ocmp},
		{"long comparison", 42, true, cminor.Ocmpl},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := cminor.Evar{Name: "x"}
			result := makeComparison(expr, tt.value, tt.isLong)

			if cmp, ok := result.(cminor.Ecmp); ok {
				if cmp.Op != tt.wantOp {
					t.Errorf("comparison op = %v, want %v", cmp.Op, tt.wantOp)
				}
				if cmp.Cmp != cminor.Ceq {
					t.Errorf("comparison type = %v, want Ceq", cmp.Cmp)
				}
			} else {
				t.Errorf("expected Ecmp, got %T", result)
			}
		})
	}
}

func TestMakeLessThan(t *testing.T) {
	tests := []struct {
		name   string
		value  int64
		isLong bool
		wantOp cminor.BinaryOp
	}{
		{"int less than", 10, false, cminor.Ocmp},
		{"long less than", 10, true, cminor.Ocmpl},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := cminor.Evar{Name: "x"}
			result := makeLessThan(expr, tt.value, tt.isLong)

			if cmp, ok := result.(cminor.Ecmp); ok {
				if cmp.Op != tt.wantOp {
					t.Errorf("comparison op = %v, want %v", cmp.Op, tt.wantOp)
				}
				if cmp.Cmp != cminor.Clt {
					t.Errorf("comparison type = %v, want Clt", cmp.Cmp)
				}
			} else {
				t.Errorf("expected Ecmp, got %T", result)
			}
		})
	}
}

func TestMakeConstant(t *testing.T) {
	tests := []struct {
		name   string
		value  int64
		isLong bool
	}{
		{"int constant", 42, false},
		{"long constant", 42, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeConstant(tt.value, tt.isLong)

			if c, ok := result.(cminor.Econst); ok {
				if tt.isLong {
					if lc, ok := c.Const.(cminor.Olongconst); ok {
						if lc.Value != tt.value {
							t.Errorf("long value = %d, want %d", lc.Value, tt.value)
						}
					} else {
						t.Errorf("expected Olongconst, got %T", c.Const)
					}
				} else {
					if ic, ok := c.Const.(cminor.Ointconst); ok {
						if int64(ic.Value) != tt.value {
							t.Errorf("int value = %d, want %d", ic.Value, tt.value)
						}
					} else {
						t.Errorf("expected Ointconst, got %T", c.Const)
					}
				}
			} else {
				t.Errorf("expected Econst, got %T", result)
			}
		})
	}
}
