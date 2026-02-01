// Package selection - Operator selection for instruction selection pass.
// This file maps Cminor high-level operators to machine-level operators.
// The selection process includes recognizing combined operation patterns
// (like shift+add on ARM64) for better code generation.
package selection

import (
	"github.com/raymyers/ralph-cc/pkg/cminor"
	"github.com/raymyers/ralph-cc/pkg/cminorsel"
)

// OpResult represents the result of operator selection.
// It can be a simple operator mapping or a combined operation.
type OpResult struct {
	// For simple operations
	UnaryOp  *cminorsel.MachUnaryOp
	BinaryOp *cminorsel.MachBinaryOp

	// For combined operations (ARM64 shift+op patterns)
	Combined   bool
	CombinedOp cminorsel.MachBinaryOp
	Shift      int // shift amount for combined ops
}

// SelectUnaryOp maps a Cminor unary operator to a machine unary operator.
func SelectUnaryOp(op cminor.UnaryOp) cminorsel.MachUnaryOp {
	switch op {
	// Negation
	case cminor.Onegint:
		return cminorsel.MOnegint
	case cminor.Onegf:
		return cminorsel.MOnegf
	case cminor.Onegl:
		return cminorsel.MOnegl
	case cminor.Onegs:
		return cminorsel.MOnegs

	// Bitwise not
	case cminor.Onotint:
		return cminorsel.MOnotint
	case cminor.Onotl:
		return cminorsel.MOnotl

	// Cast operations
	case cminor.Ocast8signed:
		return cminorsel.MOcast8signed
	case cminor.Ocast8unsigned:
		return cminorsel.MOcast8unsigned
	case cminor.Ocast16signed:
		return cminorsel.MOcast16signed
	case cminor.Ocast16unsigned:
		return cminorsel.MOcast16unsigned

	// Float conversions
	case cminor.Osingleoffloat:
		return cminorsel.MOsingleoffloat
	case cminor.Ofloatofsingle:
		return cminorsel.MOfloatofsingle

	// Int/float conversions
	case cminor.Ointoffloat:
		return cminorsel.MOintoffloat
	case cminor.Ointuoffloat:
		return cminorsel.MOintuoffloat
	case cminor.Ofloatofint:
		return cminorsel.MOfloatofint
	case cminor.Ofloatofintu:
		return cminorsel.MOfloatofintu

	// Long/float conversions
	case cminor.Olongoffloat:
		return cminorsel.MOlongoffloat
	case cminor.Olonguoffloat:
		return cminorsel.MOlonguoffloat
	case cminor.Ofloatoflong:
		return cminorsel.MOfloatoflong
	case cminor.Ofloatoflongu:
		return cminorsel.MOfloatoflongu

	// Long/single conversions
	case cminor.Olongofsingle:
		return cminorsel.MOlongofsingle
	case cminor.Olonguofsingle:
		return cminorsel.MOlonguofsingle
	case cminor.Osingleoflong:
		return cminorsel.MOsingleoflong
	case cminor.Osingleoflongu:
		return cminorsel.MOsingleoflongu

	// Int/long conversions
	case cminor.Ointoflong:
		return cminorsel.MOintoflong
	case cminor.Olongofint:
		return cminorsel.MOlongofint
	case cminor.Olongofintu:
		return cminorsel.MOlongofintu

	default:
		// Onotbool maps to negint (boolean is 0/1 integer)
		return cminorsel.MOnegint
	}
}

// SelectBinaryOp maps a Cminor binary operator to a machine binary operator.
func SelectBinaryOp(op cminor.BinaryOp) cminorsel.MachBinaryOp {
	switch op {
	// Integer arithmetic
	case cminor.Oadd:
		return cminorsel.MOadd
	case cminor.Osub:
		return cminorsel.MOsub
	case cminor.Omul:
		return cminorsel.MOmul
	case cminor.Odiv:
		return cminorsel.MOdiv
	case cminor.Odivu:
		return cminorsel.MOdivu
	case cminor.Omod:
		return cminorsel.MOmod
	case cminor.Omodu:
		return cminorsel.MOmodu

	// Float64 arithmetic
	case cminor.Oaddf:
		return cminorsel.MOaddf
	case cminor.Osubf:
		return cminorsel.MOsubf
	case cminor.Omulf:
		return cminorsel.MOmulf
	case cminor.Odivf:
		return cminorsel.MOdivf

	// Float32 arithmetic
	case cminor.Oadds:
		return cminorsel.MOadds
	case cminor.Osubs:
		return cminorsel.MOsubs
	case cminor.Omuls:
		return cminorsel.MOmuls
	case cminor.Odivs:
		return cminorsel.MOdivs

	// Long arithmetic
	case cminor.Oaddl:
		return cminorsel.MOaddl
	case cminor.Osubl:
		return cminorsel.MOsubl
	case cminor.Omull:
		return cminorsel.MOmull
	case cminor.Odivl:
		return cminorsel.MOdivl
	case cminor.Odivlu:
		return cminorsel.MOdivlu
	case cminor.Omodl:
		return cminorsel.MOmodl
	case cminor.Omodlu:
		return cminorsel.MOmodlu

	// Integer bitwise
	case cminor.Oand:
		return cminorsel.MOand
	case cminor.Oor:
		return cminorsel.MOor
	case cminor.Oxor:
		return cminorsel.MOxor
	case cminor.Oshl:
		return cminorsel.MOshl
	case cminor.Oshr:
		return cminorsel.MOshr
	case cminor.Oshru:
		return cminorsel.MOshru

	// Long bitwise
	case cminor.Oandl:
		return cminorsel.MOandl
	case cminor.Oorl:
		return cminorsel.MOorl
	case cminor.Oxorl:
		return cminorsel.MOxorl
	case cminor.Oshll:
		return cminorsel.MOshll
	case cminor.Oshrl:
		return cminorsel.MOshrl
	case cminor.Oshrlu:
		return cminorsel.MOshrlu

	// Comparisons
	case cminor.Ocmp:
		return cminorsel.MOcmp
	case cminor.Ocmpu:
		return cminorsel.MOcmpu
	case cminor.Ocmpf:
		return cminorsel.MOcmpf
	case cminor.Ocmps:
		return cminorsel.MOcmps
	case cminor.Ocmpl:
		return cminorsel.MOcmpl
	case cminor.Ocmplu:
		return cminorsel.MOcmplu

	default:
		return cminorsel.MOadd
	}
}

// CombinedOpResult holds the result of combined operation detection.
type CombinedOpResult struct {
	IsCombined bool
	Op         cminorsel.MachBinaryOp
	Shift      int
	Base       cminor.Expr // non-shifted operand
	Index      cminor.Expr // shifted operand (before shift)
}

// TrySelectCombinedOp attempts to recognize combined shift+arithmetic patterns.
// ARM64 supports: add/sub/and/or/xor with shifted second operand.
// Pattern: op(base, shift(index, amount))
func TrySelectCombinedOp(op cminor.BinaryOp, left, right cminor.Expr) CombinedOpResult {
	// Check if right operand is a shift operation
	if shift, index, amount, ok := extractShift(right); ok && isValidShiftAmount(amount) {
		switch op {
		case cminor.Oadd:
			if shift == cminor.Oshl {
				return CombinedOpResult{true, cminorsel.MOaddshift, amount, left, index}
			}
		case cminor.Osub:
			if shift == cminor.Oshl {
				return CombinedOpResult{true, cminorsel.MOsubshift, amount, left, index}
			}
		case cminor.Oand:
			if shift == cminor.Oshl {
				return CombinedOpResult{true, cminorsel.MOandshift, amount, left, index}
			}
		case cminor.Oor:
			if shift == cminor.Oshl {
				return CombinedOpResult{true, cminorsel.MOorshift, amount, left, index}
			}
		case cminor.Oxor:
			if shift == cminor.Oshl {
				return CombinedOpResult{true, cminorsel.MOxorshift, amount, left, index}
			}

		// Long versions
		case cminor.Oaddl:
			if shift == cminor.Oshll {
				return CombinedOpResult{true, cminorsel.MOaddlshift, amount, left, index}
			}
		case cminor.Osubl:
			if shift == cminor.Oshll {
				return CombinedOpResult{true, cminorsel.MOsublshift, amount, left, index}
			}
		case cminor.Oandl:
			if shift == cminor.Oshll {
				return CombinedOpResult{true, cminorsel.MOandlshift, amount, left, index}
			}
		case cminor.Oorl:
			if shift == cminor.Oshll {
				return CombinedOpResult{true, cminorsel.MOorlshift, amount, left, index}
			}
		case cminor.Oxorl:
			if shift == cminor.Oshll {
				return CombinedOpResult{true, cminorsel.MOxorlshift, amount, left, index}
			}
		}
	}

	// For commutative operations, also check if left operand is shifted
	if isCommutative(op) {
		if shift, index, amount, ok := extractShift(left); ok && isValidShiftAmount(amount) {
			switch op {
			case cminor.Oadd:
				if shift == cminor.Oshl {
					return CombinedOpResult{true, cminorsel.MOaddshift, amount, right, index}
				}
			case cminor.Oand:
				if shift == cminor.Oshl {
					return CombinedOpResult{true, cminorsel.MOandshift, amount, right, index}
				}
			case cminor.Oor:
				if shift == cminor.Oshl {
					return CombinedOpResult{true, cminorsel.MOorshift, amount, right, index}
				}
			case cminor.Oxor:
				if shift == cminor.Oshl {
					return CombinedOpResult{true, cminorsel.MOxorshift, amount, right, index}
				}

			// Long versions
			case cminor.Oaddl:
				if shift == cminor.Oshll {
					return CombinedOpResult{true, cminorsel.MOaddlshift, amount, right, index}
				}
			case cminor.Oandl:
				if shift == cminor.Oshll {
					return CombinedOpResult{true, cminorsel.MOandlshift, amount, right, index}
				}
			case cminor.Oorl:
				if shift == cminor.Oshll {
					return CombinedOpResult{true, cminorsel.MOorlshift, amount, right, index}
				}
			case cminor.Oxorl:
				if shift == cminor.Oshll {
					return CombinedOpResult{true, cminorsel.MOxorlshift, amount, right, index}
				}
			}
		}
	}

	return CombinedOpResult{IsCombined: false}
}

// extractShift checks if expr is a shift operation and returns (shiftOp, operand, amount, ok)
func extractShift(e cminor.Expr) (cminor.BinaryOp, cminor.Expr, int, bool) {
	binop, ok := e.(cminor.Ebinop)
	if !ok {
		return 0, nil, 0, false
	}

	// Check for shift operations
	switch binop.Op {
	case cminor.Oshl, cminor.Oshr, cminor.Oshru,
		cminor.Oshll, cminor.Oshrl, cminor.Oshrlu:
		// Get shift amount (must be constant)
		if amt := extractConstantInt(binop.Right); amt != nil {
			return binop.Op, binop.Left, int(*amt), true
		}
	}

	return 0, nil, 0, false
}

// isValidShiftAmount checks if shift amount is valid for ARM64 (0-63 for 64-bit, 0-31 for 32-bit)
func isValidShiftAmount(amount int) bool {
	return amount >= 0 && amount <= 63
}

// isCommutative returns true if the binary operation is commutative
func isCommutative(op cminor.BinaryOp) bool {
	switch op {
	case cminor.Oadd, cminor.Omul, cminor.Oand, cminor.Oor, cminor.Oxor,
		cminor.Oaddl, cminor.Omull, cminor.Oandl, cminor.Oorl, cminor.Oxorl,
		cminor.Oaddf, cminor.Omulf, cminor.Oadds, cminor.Omuls:
		return true
	}
	return false
}

// SelectComparison maps a Cminor comparison to a CminorSel condition.
// This is used for branch conditions.
func SelectComparison(cmpOp cminor.BinaryOp, cmp cminor.Comparison, left, right cminorsel.Expr) cminorsel.Condition {
	switch cmpOp {
	case cminor.Ocmp:
		return cminorsel.CondCmp{Cmp: cminorsel.Comparison(cmp), Left: left, Right: right}
	case cminor.Ocmpu:
		return cminorsel.CondCmpu{Cmp: cminorsel.Comparison(cmp), Left: left, Right: right}
	case cminor.Ocmpf:
		return cminorsel.CondCmpf{Cmp: cminorsel.Comparison(cmp), Left: left, Right: right}
	case cminor.Ocmps:
		return cminorsel.CondCmps{Cmp: cminorsel.Comparison(cmp), Left: left, Right: right}
	case cminor.Ocmpl:
		return cminorsel.CondCmpl{Cmp: cminorsel.Comparison(cmp), Left: left, Right: right}
	case cminor.Ocmplu:
		return cminorsel.CondCmplu{Cmp: cminorsel.Comparison(cmp), Left: left, Right: right}
	default:
		// Default to signed int comparison
		return cminorsel.CondCmp{Cmp: cminorsel.Comparison(cmp), Left: left, Right: right}
	}
}

// NegateComparison returns the negation of a comparison.
func NegateComparison(cmp cminor.Comparison) cminor.Comparison {
	switch cmp {
	case cminor.Ceq:
		return cminor.Cne
	case cminor.Cne:
		return cminor.Ceq
	case cminor.Clt:
		return cminor.Cge
	case cminor.Cle:
		return cminor.Cgt
	case cminor.Cgt:
		return cminor.Cle
	case cminor.Cge:
		return cminor.Clt
	default:
		return cmp
	}
}

// SwapComparison returns the comparison for swapped operands.
// E.g., a < b becomes b > a
func SwapComparison(cmp cminor.Comparison) cminor.Comparison {
	switch cmp {
	case cminor.Clt:
		return cminor.Cgt
	case cminor.Cle:
		return cminor.Cge
	case cminor.Cgt:
		return cminor.Clt
	case cminor.Cge:
		return cminor.Cle
	default:
		return cmp // Ceq and Cne are symmetric
	}
}

// Note: extractConstantInt is defined in addressing.go and reused here.
