package selection

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/cminor"
	"github.com/raymyers/ralph-cc/pkg/cminorsel"
)

func TestSelectUnaryOp(t *testing.T) {
	tests := []struct {
		name   string
		input  cminor.UnaryOp
		expect cminorsel.MachUnaryOp
	}{
		// Negation
		{"negint", cminor.Onegint, cminorsel.MOnegint},
		{"negf", cminor.Onegf, cminorsel.MOnegf},
		{"negl", cminor.Onegl, cminorsel.MOnegl},
		{"negs", cminor.Onegs, cminorsel.MOnegs},

		// Bitwise not
		{"notint", cminor.Onotint, cminorsel.MOnotint},
		{"notl", cminor.Onotl, cminorsel.MOnotl},

		// Cast operations
		{"cast8signed", cminor.Ocast8signed, cminorsel.MOcast8signed},
		{"cast8unsigned", cminor.Ocast8unsigned, cminorsel.MOcast8unsigned},
		{"cast16signed", cminor.Ocast16signed, cminorsel.MOcast16signed},
		{"cast16unsigned", cminor.Ocast16unsigned, cminorsel.MOcast16unsigned},

		// Float conversions
		{"singleoffloat", cminor.Osingleoffloat, cminorsel.MOsingleoffloat},
		{"floatofsingle", cminor.Ofloatofsingle, cminorsel.MOfloatofsingle},

		// Int/float conversions
		{"intoffloat", cminor.Ointoffloat, cminorsel.MOintoffloat},
		{"intuoffloat", cminor.Ointuoffloat, cminorsel.MOintuoffloat},
		{"floatofint", cminor.Ofloatofint, cminorsel.MOfloatofint},
		{"floatofintu", cminor.Ofloatofintu, cminorsel.MOfloatofintu},

		// Long/float conversions
		{"longoffloat", cminor.Olongoffloat, cminorsel.MOlongoffloat},
		{"longuoffloat", cminor.Olonguoffloat, cminorsel.MOlonguoffloat},
		{"floatoflong", cminor.Ofloatoflong, cminorsel.MOfloatoflong},
		{"floatoflongu", cminor.Ofloatoflongu, cminorsel.MOfloatoflongu},

		// Int/long conversions
		{"intoflong", cminor.Ointoflong, cminorsel.MOintoflong},
		{"longofint", cminor.Olongofint, cminorsel.MOlongofint},
		{"longofintu", cminor.Olongofintu, cminorsel.MOlongofintu},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SelectUnaryOp(tc.input)
			if got != tc.expect {
				t.Errorf("SelectUnaryOp(%v) = %v, want %v", tc.input, got, tc.expect)
			}
		})
	}
}

func TestSelectBinaryOp(t *testing.T) {
	tests := []struct {
		name   string
		input  cminor.BinaryOp
		expect cminorsel.MachBinaryOp
	}{
		// Integer arithmetic
		{"add", cminor.Oadd, cminorsel.MOadd},
		{"sub", cminor.Osub, cminorsel.MOsub},
		{"mul", cminor.Omul, cminorsel.MOmul},
		{"div", cminor.Odiv, cminorsel.MOdiv},
		{"divu", cminor.Odivu, cminorsel.MOdivu},
		{"mod", cminor.Omod, cminorsel.MOmod},
		{"modu", cminor.Omodu, cminorsel.MOmodu},

		// Float64 arithmetic
		{"addf", cminor.Oaddf, cminorsel.MOaddf},
		{"subf", cminor.Osubf, cminorsel.MOsubf},
		{"mulf", cminor.Omulf, cminorsel.MOmulf},
		{"divf", cminor.Odivf, cminorsel.MOdivf},

		// Float32 arithmetic
		{"adds", cminor.Oadds, cminorsel.MOadds},
		{"subs", cminor.Osubs, cminorsel.MOsubs},
		{"muls", cminor.Omuls, cminorsel.MOmuls},
		{"divs", cminor.Odivs, cminorsel.MOdivs},

		// Long arithmetic
		{"addl", cminor.Oaddl, cminorsel.MOaddl},
		{"subl", cminor.Osubl, cminorsel.MOsubl},
		{"mull", cminor.Omull, cminorsel.MOmull},
		{"divl", cminor.Odivl, cminorsel.MOdivl},
		{"divlu", cminor.Odivlu, cminorsel.MOdivlu},
		{"modl", cminor.Omodl, cminorsel.MOmodl},
		{"modlu", cminor.Omodlu, cminorsel.MOmodlu},

		// Integer bitwise
		{"and", cminor.Oand, cminorsel.MOand},
		{"or", cminor.Oor, cminorsel.MOor},
		{"xor", cminor.Oxor, cminorsel.MOxor},
		{"shl", cminor.Oshl, cminorsel.MOshl},
		{"shr", cminor.Oshr, cminorsel.MOshr},
		{"shru", cminor.Oshru, cminorsel.MOshru},

		// Long bitwise
		{"andl", cminor.Oandl, cminorsel.MOandl},
		{"orl", cminor.Oorl, cminorsel.MOorl},
		{"xorl", cminor.Oxorl, cminorsel.MOxorl},
		{"shll", cminor.Oshll, cminorsel.MOshll},
		{"shrl", cminor.Oshrl, cminorsel.MOshrl},
		{"shrlu", cminor.Oshrlu, cminorsel.MOshrlu},

		// Comparisons
		{"cmp", cminor.Ocmp, cminorsel.MOcmp},
		{"cmpu", cminor.Ocmpu, cminorsel.MOcmpu},
		{"cmpf", cminor.Ocmpf, cminorsel.MOcmpf},
		{"cmps", cminor.Ocmps, cminorsel.MOcmps},
		{"cmpl", cminor.Ocmpl, cminorsel.MOcmpl},
		{"cmplu", cminor.Ocmplu, cminorsel.MOcmplu},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SelectBinaryOp(tc.input)
			if got != tc.expect {
				t.Errorf("SelectBinaryOp(%v) = %v, want %v", tc.input, got, tc.expect)
			}
		})
	}
}

func TestTrySelectCombinedOp(t *testing.T) {
	// Helper to create shift expression: x << n
	makeShift := func(op cminor.BinaryOp, arg cminor.Expr, amount int32) cminor.Expr {
		return cminor.Ebinop{
			Op:    op,
			Left:  arg,
			Right: cminor.Econst{Const: cminor.Ointconst{Value: amount}},
		}
	}

	x := cminor.Evar{Name: "x"}
	y := cminor.Evar{Name: "y"}

	tests := []struct {
		name     string
		op       cminor.BinaryOp
		left     cminor.Expr
		right    cminor.Expr
		combined bool
		expectOp cminorsel.MachBinaryOp
		shift    int
	}{
		// add with shifted operand: x + (y << 2)
		{
			name:     "add_shift_right",
			op:       cminor.Oadd,
			left:     x,
			right:    makeShift(cminor.Oshl, y, 2),
			combined: true,
			expectOp: cminorsel.MOaddshift,
			shift:    2,
		},
		// sub with shifted operand: x - (y << 3)
		{
			name:     "sub_shift_right",
			op:       cminor.Osub,
			left:     x,
			right:    makeShift(cminor.Oshl, y, 3),
			combined: true,
			expectOp: cminorsel.MOsubshift,
			shift:    3,
		},
		// and with shifted operand: x & (y << 4)
		{
			name:     "and_shift_right",
			op:       cminor.Oand,
			left:     x,
			right:    makeShift(cminor.Oshl, y, 4),
			combined: true,
			expectOp: cminorsel.MOandshift,
			shift:    4,
		},
		// or with shifted operand: x | (y << 5)
		{
			name:     "or_shift_right",
			op:       cminor.Oor,
			left:     x,
			right:    makeShift(cminor.Oshl, y, 5),
			combined: true,
			expectOp: cminorsel.MOorshift,
			shift:    5,
		},
		// xor with shifted operand: x ^ (y << 6)
		{
			name:     "xor_shift_right",
			op:       cminor.Oxor,
			left:     x,
			right:    makeShift(cminor.Oshl, y, 6),
			combined: true,
			expectOp: cminorsel.MOxorshift,
			shift:    6,
		},
		// Commutative: (y << 2) + x => x + (y << 2)
		{
			name:     "add_shift_left_commutative",
			op:       cminor.Oadd,
			left:     makeShift(cminor.Oshl, y, 2),
			right:    x,
			combined: true,
			expectOp: cminorsel.MOaddshift,
			shift:    2,
		},
		// Long versions
		{
			name:     "addl_shift",
			op:       cminor.Oaddl,
			left:     x,
			right:    makeShift(cminor.Oshll, y, 3),
			combined: true,
			expectOp: cminorsel.MOaddlshift,
			shift:    3,
		},
		{
			name:     "subl_shift",
			op:       cminor.Osubl,
			left:     x,
			right:    makeShift(cminor.Oshll, y, 4),
			combined: true,
			expectOp: cminorsel.MOsublshift,
			shift:    4,
		},
		// Non-combined cases
		{
			name:     "add_no_shift",
			op:       cminor.Oadd,
			left:     x,
			right:    y,
			combined: false,
		},
		{
			name:     "sub_non_commutative_shift_left",
			op:       cminor.Osub,
			left:     makeShift(cminor.Oshl, y, 2),
			right:    x,
			combined: false, // sub is not commutative
		},
		// Shift right doesn't combine with add on ARM64
		{
			name:     "add_shift_right_not_combined",
			op:       cminor.Oadd,
			left:     x,
			right:    makeShift(cminor.Oshr, y, 2),
			combined: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := TrySelectCombinedOp(tc.op, tc.left, tc.right)
			if result.IsCombined != tc.combined {
				t.Errorf("IsCombined = %v, want %v", result.IsCombined, tc.combined)
			}
			if tc.combined {
				if result.Op != tc.expectOp {
					t.Errorf("Op = %v, want %v", result.Op, tc.expectOp)
				}
				if result.Shift != tc.shift {
					t.Errorf("Shift = %v, want %v", result.Shift, tc.shift)
				}
			}
		})
	}
}

func TestSelectComparison(t *testing.T) {
	left := cminorsel.Evar{Name: "x"}
	right := cminorsel.Evar{Name: "y"}

	tests := []struct {
		name   string
		cmpOp  cminor.BinaryOp
		cmp    cminor.Comparison
		expect string // type name for verification
	}{
		{"cmp_eq", cminor.Ocmp, cminor.Ceq, "CondCmp"},
		{"cmpu_ne", cminor.Ocmpu, cminor.Cne, "CondCmpu"},
		{"cmpf_lt", cminor.Ocmpf, cminor.Clt, "CondCmpf"},
		{"cmps_le", cminor.Ocmps, cminor.Cle, "CondCmps"},
		{"cmpl_gt", cminor.Ocmpl, cminor.Cgt, "CondCmpl"},
		{"cmplu_ge", cminor.Ocmplu, cminor.Cge, "CondCmplu"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := SelectComparison(tc.cmpOp, tc.cmp, left, right)

			// Verify the correct type was produced
			switch tc.expect {
			case "CondCmp":
				if _, ok := result.(cminorsel.CondCmp); !ok {
					t.Errorf("expected CondCmp, got %T", result)
				}
			case "CondCmpu":
				if _, ok := result.(cminorsel.CondCmpu); !ok {
					t.Errorf("expected CondCmpu, got %T", result)
				}
			case "CondCmpf":
				if _, ok := result.(cminorsel.CondCmpf); !ok {
					t.Errorf("expected CondCmpf, got %T", result)
				}
			case "CondCmps":
				if _, ok := result.(cminorsel.CondCmps); !ok {
					t.Errorf("expected CondCmps, got %T", result)
				}
			case "CondCmpl":
				if _, ok := result.(cminorsel.CondCmpl); !ok {
					t.Errorf("expected CondCmpl, got %T", result)
				}
			case "CondCmplu":
				if _, ok := result.(cminorsel.CondCmplu); !ok {
					t.Errorf("expected CondCmplu, got %T", result)
				}
			}
		})
	}
}

func TestNegateComparison(t *testing.T) {
	tests := []struct {
		input  cminor.Comparison
		expect cminor.Comparison
	}{
		{cminor.Ceq, cminor.Cne},
		{cminor.Cne, cminor.Ceq},
		{cminor.Clt, cminor.Cge},
		{cminor.Cle, cminor.Cgt},
		{cminor.Cgt, cminor.Cle},
		{cminor.Cge, cminor.Clt},
	}

	for _, tc := range tests {
		t.Run(tc.input.String(), func(t *testing.T) {
			got := NegateComparison(tc.input)
			if got != tc.expect {
				t.Errorf("NegateComparison(%v) = %v, want %v", tc.input, got, tc.expect)
			}
		})
	}
}

func TestSwapComparison(t *testing.T) {
	tests := []struct {
		input  cminor.Comparison
		expect cminor.Comparison
	}{
		{cminor.Ceq, cminor.Ceq}, // symmetric
		{cminor.Cne, cminor.Cne}, // symmetric
		{cminor.Clt, cminor.Cgt},
		{cminor.Cle, cminor.Cge},
		{cminor.Cgt, cminor.Clt},
		{cminor.Cge, cminor.Cle},
	}

	for _, tc := range tests {
		t.Run(tc.input.String(), func(t *testing.T) {
			got := SwapComparison(tc.input)
			if got != tc.expect {
				t.Errorf("SwapComparison(%v) = %v, want %v", tc.input, got, tc.expect)
			}
		})
	}
}

func TestIsCommutative(t *testing.T) {
	commutative := []cminor.BinaryOp{
		cminor.Oadd, cminor.Omul, cminor.Oand, cminor.Oor, cminor.Oxor,
		cminor.Oaddl, cminor.Omull, cminor.Oandl, cminor.Oorl, cminor.Oxorl,
		cminor.Oaddf, cminor.Omulf, cminor.Oadds, cminor.Omuls,
	}

	nonCommutative := []cminor.BinaryOp{
		cminor.Osub, cminor.Odiv, cminor.Omod, cminor.Oshl, cminor.Oshr,
		cminor.Osubl, cminor.Odivl, cminor.Omodl,
		cminor.Osubf, cminor.Odivf, cminor.Osubs, cminor.Odivs,
	}

	for _, op := range commutative {
		if !isCommutative(op) {
			t.Errorf("isCommutative(%v) = false, want true", op)
		}
	}

	for _, op := range nonCommutative {
		if isCommutative(op) {
			t.Errorf("isCommutative(%v) = true, want false", op)
		}
	}
}
