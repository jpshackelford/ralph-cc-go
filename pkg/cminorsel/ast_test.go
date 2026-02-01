package cminorsel

import "testing"

func TestAddressingModeInterface(t *testing.T) {
	// Test that all addressing modes implement the interface
	modes := []AddressingMode{
		Aindexed{Offset: 8},
		Aindexed2{},
		Aindexed2shift{Shift: 2},
		Aindexed2ext{Extend: Xsgn32, Shift: 2},
		Aglobal{Symbol: "foo", Offset: 0},
		Ainstack{Offset: 16},
	}
	for i, m := range modes {
		// Verify interface implementation
		m.implAddressingMode()
		_ = i
	}
	if len(modes) != 6 {
		t.Errorf("expected 6 addressing modes, got %d", len(modes))
	}
}

func TestConditionInterface(t *testing.T) {
	x := Evar{Name: "x"}
	y := Evar{Name: "y"}

	// Test that all conditions implement the interface
	conditions := []Condition{
		CondTrue{},
		CondFalse{},
		CondCmp{Cmp: Ceq, Left: x, Right: y},
		CondCmpu{Cmp: Clt, Left: x, Right: y},
		CondCmpf{Cmp: Cle, Left: x, Right: y},
		CondCmps{Cmp: Cgt, Left: x, Right: y},
		CondCmpl{Cmp: Cge, Left: x, Right: y},
		CondCmplu{Cmp: Cne, Left: x, Right: y},
		CondNot{Cond: CondTrue{}},
		CondAnd{Left: CondTrue{}, Right: CondFalse{}},
		CondOr{Left: CondTrue{}, Right: CondFalse{}},
	}
	for i, c := range conditions {
		c.implCondition()
		_ = i
	}
	if len(conditions) != 11 {
		t.Errorf("expected 11 conditions, got %d", len(conditions))
	}
}

func TestConstantInterface(t *testing.T) {
	constants := []Constant{
		Ointconst{Value: 42},
		Ofloatconst{Value: 3.14},
		Olongconst{Value: 1000000000000},
		Osingleconst{Value: 2.71},
		Oaddrsymbol{Symbol: "main", Offset: 0},
		Oaddrstack{Offset: 16},
	}
	for _, c := range constants {
		c.implCminorSelConst()
		c.implCminorSelNode()
	}
	if len(constants) != 6 {
		t.Errorf("expected 6 constant types, got %d", len(constants))
	}
}

func TestExprInterface(t *testing.T) {
	x := Evar{Name: "x"}
	y := Evar{Name: "y"}

	exprs := []Expr{
		Evar{Name: "foo"},
		Econst{Const: Ointconst{Value: 5}},
		Eunop{Op: Onegint, Arg: x},
		Ebinop{Op: Oadd, Left: x, Right: y},
		Eload{Chunk: Mint32, Mode: Aindexed{Offset: 0}, Args: []Expr{x}},
		Econdition{Cond: CondTrue{}, Then: x, Else: y},
		Elet{Bind: x, Body: Eletvar{Index: 0}},
		Eletvar{Index: 0},
		Eaddshift{Op: Slsl, Shift: 2, Left: x, Right: y},
		Esubshift{Op: Slsr, Shift: 3, Left: x, Right: y},
	}
	for _, e := range exprs {
		e.implCminorSelExpr()
		e.implCminorSelNode()
	}
	if len(exprs) != 10 {
		t.Errorf("expected 10 expression types, got %d", len(exprs))
	}
}

func TestStmtInterface(t *testing.T) {
	x := Evar{Name: "x"}

	stmts := []Stmt{
		Sskip{},
		Sassign{Name: "x", RHS: Econst{Const: Ointconst{Value: 0}}},
		Sstore{Chunk: Mint32, Mode: Aindexed{Offset: 0}, Args: []Expr{x}, Value: x},
		Scall{Result: nil, Func: x, Args: nil},
		Stailcall{Func: x, Args: nil},
		Sbuiltin{Builtin: "__builtin_trap", Args: nil},
		Sseq{First: Sskip{}, Second: Sskip{}},
		Sifthenelse{Cond: CondTrue{}, Then: Sskip{}, Else: Sskip{}},
		Sloop{Body: Sskip{}},
		Sblock{Body: Sskip{}},
		Sexit{N: 1},
		Sswitch{IsLong: false, Expr: x, Cases: nil, Default: Sskip{}},
		Sreturn{Value: x},
		Slabel{Label: "L1", Body: Sskip{}},
		Sgoto{Label: "L1"},
	}
	for _, s := range stmts {
		s.implCminorSelStmt()
		s.implCminorSelNode()
	}
	if len(stmts) != 15 {
		t.Errorf("expected 15 statement types, got %d", len(stmts))
	}
}

func TestSeq(t *testing.T) {
	s1 := Sassign{Name: "x", RHS: Econst{Const: Ointconst{Value: 1}}}
	s2 := Sassign{Name: "y", RHS: Econst{Const: Ointconst{Value: 2}}}
	s3 := Sassign{Name: "z", RHS: Econst{Const: Ointconst{Value: 3}}}

	// Empty sequence
	if _, ok := Seq().(Sskip); !ok {
		t.Error("Seq() should return Sskip")
	}

	// Single statement
	if s := Seq(s1); s != s1 {
		t.Error("Seq(s1) should return s1")
	}

	// Skip is flattened
	if s := Seq(Sskip{}, s1); s != s1 {
		t.Error("Seq(skip, s1) should return s1")
	}
	if s := Seq(s1, Sskip{}); s != s1 {
		t.Error("Seq(s1, skip) should return s1")
	}

	// Multiple statements
	seq := Seq(s1, s2, s3)
	outer, ok := seq.(Sseq)
	if !ok {
		t.Fatal("Seq(s1,s2,s3) should be Sseq")
	}
	inner, ok := outer.First.(Sseq)
	if !ok {
		t.Fatal("first should be Sseq")
	}
	if inner.First != s1 {
		t.Error("innermost first should be s1")
	}
	if inner.Second != s2 {
		t.Error("innermost second should be s2")
	}
	if outer.Second != s3 {
		t.Error("outer second should be s3")
	}
}

func TestExtendOpString(t *testing.T) {
	tests := []struct {
		op   ExtendOp
		want string
	}{
		{Xsgn32, "sxtw"},
		{Xuns32, "uxtw"},
		{ExtendOp(99), "?"},
	}
	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("ExtendOp(%d).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestShiftOpString(t *testing.T) {
	tests := []struct {
		op   ShiftOp
		want string
	}{
		{Slsl, "lsl"},
		{Slsr, "lsr"},
		{Sasr, "asr"},
		{ShiftOp(99), "?"},
	}
	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("ShiftOp(%d).String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestChunkReexport(t *testing.T) {
	// Verify that chunk constants are properly re-exported
	chunks := []Chunk{
		Mint8signed, Mint8unsigned,
		Mint16signed, Mint16unsigned,
		Mint32, Mint64,
		Mfloat32, Mfloat64,
		Many32, Many64,
	}
	if len(chunks) != 10 {
		t.Errorf("expected 10 chunk types, got %d", len(chunks))
	}
}

func TestComparisonReexport(t *testing.T) {
	// Verify that comparison constants are properly re-exported
	comparisons := []Comparison{Ceq, Cne, Clt, Cle, Cgt, Cge}
	if len(comparisons) != 6 {
		t.Errorf("expected 6 comparison types, got %d", len(comparisons))
	}
}

func TestEloadWithAddressingModes(t *testing.T) {
	base := Evar{Name: "base"}
	idx := Evar{Name: "idx"}

	// Test various addressing modes in Eload
	tests := []struct {
		name string
		load Eload
	}{
		{
			name: "indexed",
			load: Eload{Chunk: Mint32, Mode: Aindexed{Offset: 8}, Args: []Expr{base}},
		},
		{
			name: "indexed2",
			load: Eload{Chunk: Mint64, Mode: Aindexed2{}, Args: []Expr{base, idx}},
		},
		{
			name: "indexed2shift",
			load: Eload{Chunk: Mint32, Mode: Aindexed2shift{Shift: 2}, Args: []Expr{base, idx}},
		},
		{
			name: "global",
			load: Eload{Chunk: Mint64, Mode: Aglobal{Symbol: "data", Offset: 0}, Args: nil},
		},
		{
			name: "stack",
			load: Eload{Chunk: Mint32, Mode: Ainstack{Offset: 16}, Args: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all loads are valid expressions
			var _ Expr = tt.load
			tt.load.implCminorSelExpr()
		})
	}
}

func TestSstoreWithAddressingModes(t *testing.T) {
	base := Evar{Name: "base"}
	val := Econst{Const: Ointconst{Value: 42}}

	store := Sstore{
		Chunk: Mint32,
		Mode:  Aindexed{Offset: 0},
		Args:  []Expr{base},
		Value: val,
	}

	var _ Stmt = store
	store.implCminorSelStmt()
}
