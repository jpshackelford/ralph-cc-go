package csharpminor

import (
	"bytes"
	"strings"
	"testing"

	"github.com/raymyers/ralph-cc/pkg/ctypes"
)

func TestPrintExpr_Constants(t *testing.T) {
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"int", Econst{Const: Ointconst{Value: 42}}, "42"},
		{"negative_int", Econst{Const: Ointconst{Value: -1}}, "-1"},
		{"long", Econst{Const: Olongconst{Value: 100}}, "100L"},
		{"float", Econst{Const: Ofloatconst{Value: 3.14}}, "3.14"},
		{"single", Econst{Const: Osingleconst{Value: 2.5}}, "2.5f"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printExpr(tt.expr)
			got := buf.String()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintExpr_Variables(t *testing.T) {
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"var", Evar{Name: "x"}, "x"},
		{"tempvar", Etempvar{ID: 1}, "$1"},
		{"addrof", Eaddrof{Name: "y"}, "&y"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printExpr(tt.expr)
			got := buf.String()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintExpr_Operations(t *testing.T) {
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{
			"unop_negint",
			Eunop{Op: Onegint, Arg: Etempvar{ID: 1}},
			"negint($1)",
		},
		{
			"binop_add",
			Ebinop{Op: Oadd, Left: Evar{Name: "a"}, Right: Evar{Name: "b"}},
			"add(a, b)",
		},
		{
			"cmp_eq",
			Ecmp{Op: Ocmp, Cmp: Ceq, Left: Etempvar{ID: 1}, Right: Econst{Const: Ointconst{Value: 0}}},
			"cmp == ($1, 0)",
		},
		{
			"load_int32",
			Eload{Chunk: Mint32, Addr: Eaddrof{Name: "x"}},
			"int32[&x]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printExpr(tt.expr)
			got := buf.String()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintStmt_Basic(t *testing.T) {
	tests := []struct {
		name string
		stmt Stmt
		want string
	}{
		{"return_void", Sreturn{Value: nil}, "return;\n"},
		{"return_value", Sreturn{Value: Etempvar{ID: 1}}, "return $1;\n"},
		{"goto", Sgoto{Label: "done"}, "goto done;\n"},
		{"exit", Sexit{N: 2}, "exit 2;\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printStmt(tt.stmt)
			got := buf.String()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintStmt_Store(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.printStmt(Sstore{
		Chunk: Mint32,
		Addr:  Eaddrof{Name: "x"},
		Value: Econst{Const: Ointconst{Value: 42}},
	})
	got := buf.String()
	want := "int32[&x] = 42;\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintStmt_Set(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.printStmt(Sset{
		TempID: 1,
		RHS:    Ebinop{Op: Oadd, Left: Etempvar{ID: 2}, Right: Etempvar{ID: 3}},
	})
	got := buf.String()
	want := "$1 = add($2, $3);\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrintStmt_Block(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.printStmt(Sblock{
		Body: Sexit{N: 1},
	})
	got := buf.String()
	if !strings.Contains(got, "block {") {
		t.Errorf("expected 'block {' in output: %q", got)
	}
	if !strings.Contains(got, "exit 1;") {
		t.Errorf("expected 'exit 1;' in output: %q", got)
	}
}

func TestPrintStmt_Loop(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.printStmt(Sloop{
		Body: Sreturn{Value: nil},
	})
	got := buf.String()
	if !strings.Contains(got, "loop {") {
		t.Errorf("expected 'loop {' in output: %q", got)
	}
}

func TestPrintFunction(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := Function{
		Name:   "add",
		Sig:    Sig{Return: ctypes.Int(), Args: []ctypes.Type{ctypes.Int(), ctypes.Int()}},
		Params: []string{"a", "b"},
		Temps:  []ctypes.Type{ctypes.Int()},
		Body:   Sreturn{Value: Ebinop{Op: Oadd, Left: Evar{Name: "a"}, Right: Evar{Name: "b"}}},
	}
	p.printFunction(&fn)
	got := buf.String()

	if !strings.Contains(got, "int add(a, b)") {
		t.Errorf("expected function signature in output: %q", got)
	}
	if !strings.Contains(got, "int $1;") {
		t.Errorf("expected temp declaration in output: %q", got)
	}
	if !strings.Contains(got, "return add(a, b);") {
		t.Errorf("expected return statement in output: %q", got)
	}
}

func TestPrintProgram(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	prog := Program{
		Globals: []VarDecl{{Name: "g", Size: 4}},
		Functions: []Function{
			{
				Name:   "main",
				Sig:    Sig{Return: ctypes.Int()},
				Params: nil,
				Body:   Sreturn{Value: Econst{Const: Ointconst{Value: 0}}},
			},
		},
	}
	p.PrintProgram(&prog)
	got := buf.String()

	if !strings.Contains(got, "var g[4];") {
		t.Errorf("expected global variable in output: %q", got)
	}
	if !strings.Contains(got, "int main()") {
		t.Errorf("expected function in output: %q", got)
	}
}
