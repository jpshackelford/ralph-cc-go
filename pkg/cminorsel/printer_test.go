package cminorsel

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintExpr_Var(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintExpr(Evar{Name: "x"})
	if buf.String() != "x" {
		t.Errorf("expected 'x', got %q", buf.String())
	}
}

func TestPrintExpr_Constants(t *testing.T) {
	tests := []struct {
		name     string
		expr     Expr
		expected string
	}{
		{"int", Econst{Const: Ointconst{Value: 42}}, "42"},
		{"long", Econst{Const: Olongconst{Value: 100}}, "100L"},
		{"float", Econst{Const: Ofloatconst{Value: 3.14}}, "3.14"},
		{"single", Econst{Const: Osingleconst{Value: 2.5}}, "2.5f"},
		{"symbol", Econst{Const: Oaddrsymbol{Symbol: "foo", Offset: 0}}, "&foo"},
		{"symbol_offset", Econst{Const: Oaddrsymbol{Symbol: "arr", Offset: 8}}, "&arr+8"},
		{"stack", Econst{Const: Oaddrstack{Offset: 16}}, "[sp+16]"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.PrintExpr(tc.expr)
			if buf.String() != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, buf.String())
			}
		})
	}
}

func TestPrintExpr_Unop(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintExpr(Eunop{Op: Onegint, Arg: Evar{Name: "x"}})
	expected := "negint(x)"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestPrintExpr_Binop(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintExpr(Ebinop{
		Op:    Oadd,
		Left:  Evar{Name: "x"},
		Right: Econst{Const: Ointconst{Value: 1}},
	})
	expected := "(x add 1)"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestPrintExpr_Load(t *testing.T) {
	tests := []struct {
		name     string
		expr     Eload
		expected string
	}{
		{
			"aindexed",
			Eload{
				Chunk: Mint32,
				Mode:  Aindexed{Offset: 8},
				Args:  []Expr{Evar{Name: "ptr"}},
			},
			"int32[ptr+8]",
		},
		{
			"aindexed_zero",
			Eload{
				Chunk: Mint64,
				Mode:  Aindexed{Offset: 0},
				Args:  []Expr{Evar{Name: "ptr"}},
			},
			"int64[ptr]",
		},
		{
			"aindexed2",
			Eload{
				Chunk: Mint32,
				Mode:  Aindexed2{},
				Args:  []Expr{Evar{Name: "base"}, Evar{Name: "idx"}},
			},
			"int32[base+idx]",
		},
		{
			"aindexed2shift",
			Eload{
				Chunk: Mint64,
				Mode:  Aindexed2shift{Shift: 3},
				Args:  []Expr{Evar{Name: "base"}, Evar{Name: "idx"}},
			},
			"int64[base+idx<<3]",
		},
		{
			"aglobal",
			Eload{
				Chunk: Mint32,
				Mode:  Aglobal{Symbol: "arr", Offset: 4},
				Args:  nil,
			},
			"int32[&arr+4]",
		},
		{
			"ainstack",
			Eload{
				Chunk: Mint32,
				Mode:  Ainstack{Offset: 16},
				Args:  nil,
			},
			"int32[[sp+16]]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.PrintExpr(tc.expr)
			if buf.String() != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, buf.String())
			}
		})
	}
}

func TestPrintExpr_Addshift(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintExpr(Eaddshift{
		Op:    Slsl,
		Shift: 2,
		Left:  Evar{Name: "x"},
		Right: Evar{Name: "y"},
	})
	expected := "(x + (lsl y 2))"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestPrintExpr_Condition(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintExpr(Econdition{
		Cond: CondCmp{Cmp: Cgt, Left: Evar{Name: "x"}, Right: Econst{Const: Ointconst{Value: 0}}},
		Then: Evar{Name: "a"},
		Else: Evar{Name: "b"},
	})
	expected := "(x > 0 ? a : b)"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestPrintStmt_Assign(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintStmt(Sassign{Name: "x", RHS: Econst{Const: Ointconst{Value: 42}}})
	expected := "x = 42;\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestPrintStmt_Store(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintStmt(Sstore{
		Chunk: Mint32,
		Mode:  Aindexed{Offset: 8},
		Args:  []Expr{Evar{Name: "ptr"}},
		Value: Econst{Const: Ointconst{Value: 100}},
	})
	expected := "int32[ptr+8] = 100;\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestPrintStmt_Call(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	result := "r"
	p.PrintStmt(Scall{
		Result: &result,
		Func:   Evar{Name: "foo"},
		Args:   []Expr{Evar{Name: "x"}, Econst{Const: Ointconst{Value: 1}}},
	})
	expected := "r = foo(x, 1);\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestPrintStmt_Ifthenelse(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintStmt(Sifthenelse{
		Cond: CondCmp{Cmp: Ceq, Left: Evar{Name: "x"}, Right: Econst{Const: Ointconst{Value: 0}}},
		Then: Sreturn{Value: Econst{Const: Ointconst{Value: 1}}},
		Else: Sreturn{Value: Econst{Const: Ointconst{Value: 0}}},
	})
	if !strings.Contains(buf.String(), "if (") {
		t.Errorf("expected 'if (' in output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "} else {") {
		t.Errorf("expected '} else {' in output: %q", buf.String())
	}
}

func TestPrintStmt_Loop(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintStmt(Sloop{
		Body: Sskip{},
	})
	if !strings.Contains(buf.String(), "loop {") {
		t.Errorf("expected 'loop {' in output: %q", buf.String())
	}
}

func TestPrintStmt_Switch(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintStmt(Sswitch{
		IsLong: false,
		Expr:   Evar{Name: "x"},
		Cases: []SwitchCase{
			{Value: 1, Body: Sexit{N: 1}},
		},
		Default: Sexit{N: 0},
	})
	if !strings.Contains(buf.String(), "switch (") {
		t.Errorf("expected 'switch (' in output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "case 1:") {
		t.Errorf("expected 'case 1:' in output: %q", buf.String())
	}
}

func TestPrintCond(t *testing.T) {
	tests := []struct {
		name     string
		cond     Condition
		expected string
	}{
		{"true", CondTrue{}, "true"},
		{"false", CondFalse{}, "false"},
		{"not", CondNot{Cond: CondTrue{}}, "!(true)"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printCond(tc.cond)
			if buf.String() != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, buf.String())
			}
		})
	}
}

func TestPrintFunction(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.Print(Program{
		Functions: []Function{
			{
				Name:       "add",
				Sig:        Sig{Args: []string{"i", "i"}, Return: "i"},
				Params:     []string{"a", "b"},
				Vars:       []string{"result"},
				Stackspace: 8,
				Body: Sseq{
					First: Sassign{
						Name: "result",
						RHS: Ebinop{
							Op:    Oadd,
							Left:  Evar{Name: "a"},
							Right: Evar{Name: "b"},
						},
					},
					Second: Sreturn{Value: Evar{Name: "result"}},
				},
			},
		},
	})

	output := buf.String()
	if !strings.Contains(output, "i add(i a, i b)") {
		t.Errorf("expected function signature in output: %q", output)
	}
	if !strings.Contains(output, "var result;") {
		t.Errorf("expected var declaration in output: %q", output)
	}
	if !strings.Contains(output, "stackspace 8;") {
		t.Errorf("expected stackspace in output: %q", output)
	}
}

func TestPrintGlobal(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.Print(Program{
		Globals: []GlobVar{
			{Name: "counter", Size: 4, Init: nil},
			{Name: "data", Size: 8, Init: []byte{1, 2, 3}},
		},
	})

	output := buf.String()
	if !strings.Contains(output, "var counter[4];") {
		t.Errorf("expected 'var counter[4];' in output: %q", output)
	}
	if !strings.Contains(output, "var data[8] = {...};") {
		t.Errorf("expected 'var data[8] = {...};' in output: %q", output)
	}
}
