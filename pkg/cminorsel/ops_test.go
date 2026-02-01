package cminorsel

import "testing"

func TestMachUnaryOpString(t *testing.T) {
	tests := []struct {
		op   MachUnaryOp
		want string
	}{
		{MOnegint, "negint"},
		{MOnotint, "notint"},
		{MOnegf, "negf"},
		{MOnegs, "negs"},
		{MOabsf, "absf"},
		{MOabss, "abss"},
		{MOnegl, "negl"},
		{MOnotl, "notl"},
		{MOcast8signed, "cast8s"},
		{MOcast8unsigned, "cast8u"},
		{MOcast16signed, "cast16s"},
		{MOcast16unsigned, "cast16u"},
		{MOsingleoffloat, "singleoffloat"},
		{MOfloatofsingle, "floatofsingle"},
		{MOintoffloat, "intoffloat"},
		{MOintuoffloat, "intuoffloat"},
		{MOfloatofint, "floatofint"},
		{MOfloatofintu, "floatofintu"},
		{MOlongoffloat, "longoffloat"},
		{MOlonguoffloat, "longuoffloat"},
		{MOfloatoflong, "floatoflong"},
		{MOfloatoflongu, "floatoflongu"},
		{MOlongofsingle, "longofsingle"},
		{MOlonguofsingle, "longuofsingle"},
		{MOsingleoflong, "singleoflong"},
		{MOsingleoflongu, "singleoflongu"},
		{MOintoflong, "intoflong"},
		{MOlongofint, "longofint"},
		{MOlongofintu, "longofintu"},
		{MOrbit, "rbit"},
		{MOclz, "clz"},
		{MOcls, "cls"},
		{MOrev, "rev"},
		{MOrev16, "rev16"},
		{MOsqrtf, "sqrtf"},
		{MOsqrts, "sqrts"},
		{MachUnaryOp(999), "?"},
	}
	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("MachUnaryOp(%d).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestMachBinaryOpString(t *testing.T) {
	tests := []struct {
		op   MachBinaryOp
		want string
	}{
		{MOadd, "add"},
		{MOsub, "sub"},
		{MOmul, "mul"},
		{MOmulhs, "mulhs"},
		{MOmulhu, "mulhu"},
		{MOdiv, "div"},
		{MOdivu, "divu"},
		{MOmod, "mod"},
		{MOmodu, "modu"},
		{MOaddf, "addf"},
		{MOsubf, "subf"},
		{MOmulf, "mulf"},
		{MOdivf, "divf"},
		{MOadds, "adds"},
		{MOsubs, "subs"},
		{MOmuls, "muls"},
		{MOdivs, "divs"},
		{MOaddl, "addl"},
		{MOsubl, "subl"},
		{MOmull, "mull"},
		{MOmullhs, "mullhs"},
		{MOmullhu, "mullhu"},
		{MOdivl, "divl"},
		{MOdivlu, "divlu"},
		{MOmodl, "modl"},
		{MOmodlu, "modlu"},
		{MOand, "and"},
		{MOor, "or"},
		{MOxor, "xor"},
		{MOshl, "shl"},
		{MOshr, "shr"},
		{MOshru, "shru"},
		{MOandl, "andl"},
		{MOorl, "orl"},
		{MOxorl, "xorl"},
		{MOshll, "shll"},
		{MOshrl, "shrl"},
		{MOshrlu, "shrlu"},
		{MOcmp, "cmp"},
		{MOcmpu, "cmpu"},
		{MOcmpf, "cmpf"},
		{MOcmps, "cmps"},
		{MOcmpl, "cmpl"},
		{MOcmplu, "cmplu"},
		{MOaddshift, "addshift"},
		{MOsubshift, "subshift"},
		{MOandshift, "andshift"},
		{MOorshift, "orshift"},
		{MOxorshift, "xorshift"},
		{MOaddlshift, "addlshift"},
		{MOsublshift, "sublshift"},
		{MOandlshift, "andlshift"},
		{MOorlshift, "orlshift"},
		{MOxorlshift, "xorlshift"},
		{MOmadd, "madd"},
		{MOmsub, "msub"},
		{MOmaddl, "maddl"},
		{MOmsubl, "msubl"},
		{MOmaddf, "maddf"},
		{MOmsubf, "msubf"},
		{MOmadds, "madds"},
		{MOmsubs, "msubs"},
		{MOnmaddf, "nmaddf"},
		{MOnmsubf, "nmsubf"},
		{MOnmadds, "nmadds"},
		{MOnmsubs, "nmsubs"},
		{MObic, "bic"},
		{MOorn, "orn"},
		{MOeon, "eon"},
		{MObicl, "bicl"},
		{MOornl, "ornl"},
		{MOeonl, "eonl"},
		{MOcsel, "csel"},
		{MOcsell, "csell"},
		{MachBinaryOp(999), "?"},
	}
	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("MachBinaryOp(%d).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestMachTernaryOpString(t *testing.T) {
	tests := []struct {
		op   MachTernaryOp
		want string
	}{
		{MOTmadd, "madd"},
		{MOTmsub, "msub"},
		{MOTmaddl, "maddl"},
		{MOTmsubl, "msubl"},
		{MOTmaddf, "maddf"},
		{MOTmsubf, "msubf"},
		{MOTmadds, "madds"},
		{MOTmsubs, "msubs"},
		{MachTernaryOp(999), "?"},
	}
	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("MachTernaryOp(%d).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestMachBinaryOpIsCommutative(t *testing.T) {
	commutative := []MachBinaryOp{
		MOadd, MOmul, MOaddf, MOmulf, MOadds, MOmuls,
		MOaddl, MOmull, MOand, MOor, MOxor, MOandl, MOorl, MOxorl,
	}
	nonCommutative := []MachBinaryOp{
		MOsub, MOdiv, MOdivu, MOmod, MOmodu,
		MOsubf, MOdivf, MOsubs, MOdivs,
		MOsubl, MOdivl, MOdivlu, MOmodl, MOmodlu,
		MOshl, MOshr, MOshru, MOshll, MOshrl, MOshrlu,
		MOcmp, MOcmpu, MOcmpf, MOcmps, MOcmpl, MOcmplu,
	}

	for _, op := range commutative {
		if !op.IsCommutative() {
			t.Errorf("%s should be commutative", op)
		}
	}
	for _, op := range nonCommutative {
		if op.IsCommutative() {
			t.Errorf("%s should not be commutative", op)
		}
	}
}

func TestMachBinaryOpIsCompare(t *testing.T) {
	compares := []MachBinaryOp{
		MOcmp, MOcmpu, MOcmpf, MOcmps, MOcmpl, MOcmplu,
	}
	nonCompares := []MachBinaryOp{
		MOadd, MOsub, MOmul, MOdiv, MOand, MOor,
	}

	for _, op := range compares {
		if !op.IsCompare() {
			t.Errorf("%s should be a comparison", op)
		}
	}
	for _, op := range nonCompares {
		if op.IsCompare() {
			t.Errorf("%s should not be a comparison", op)
		}
	}
}

func TestMachBinaryOpIsShiftCombined(t *testing.T) {
	shiftCombined := []MachBinaryOp{
		MOaddshift, MOsubshift, MOandshift, MOorshift, MOxorshift,
		MOaddlshift, MOsublshift, MOandlshift, MOorlshift, MOxorlshift,
	}
	nonShiftCombined := []MachBinaryOp{
		MOadd, MOsub, MOshl, MOshr, MOand, MOor,
	}

	for _, op := range shiftCombined {
		if !op.IsShiftCombined() {
			t.Errorf("%s should be a shift-combined op", op)
		}
	}
	for _, op := range nonShiftCombined {
		if op.IsShiftCombined() {
			t.Errorf("%s should not be a shift-combined op", op)
		}
	}
}

func TestMachBinaryOpIsFusedMultiply(t *testing.T) {
	fusedMultiply := []MachBinaryOp{
		MOmadd, MOmsub, MOmaddl, MOmsubl,
		MOmaddf, MOmsubf, MOmadds, MOmsubs,
		MOnmaddf, MOnmsubf, MOnmadds, MOnmsubs,
	}
	nonFusedMultiply := []MachBinaryOp{
		MOadd, MOsub, MOmul, MOdiv, MOand, MOor,
	}

	for _, op := range fusedMultiply {
		if !op.IsFusedMultiply() {
			t.Errorf("%s should be a fused multiply op", op)
		}
	}
	for _, op := range nonFusedMultiply {
		if op.IsFusedMultiply() {
			t.Errorf("%s should not be a fused multiply op", op)
		}
	}
}

func TestAllUnaryOpsHaveNames(t *testing.T) {
	// Ensure all ops have valid names (not "?")
	for op := MachUnaryOp(0); op <= MOabsfs; op++ {
		name := op.String()
		if name == "?" {
			t.Errorf("MachUnaryOp(%d) has no name", op)
		}
	}
}

func TestAllBinaryOpsHaveNames(t *testing.T) {
	// Ensure all ops have valid names (not "?")
	for op := MachBinaryOp(0); op <= MOcsell; op++ {
		name := op.String()
		if name == "?" {
			t.Errorf("MachBinaryOp(%d) has no name", op)
		}
	}
}

func TestAllTernaryOpsHaveNames(t *testing.T) {
	// Ensure all ops have valid names (not "?")
	for op := MachTernaryOp(0); op <= MOTmsubs; op++ {
		name := op.String()
		if name == "?" {
			t.Errorf("MachTernaryOp(%d) has no name", op)
		}
	}
}
