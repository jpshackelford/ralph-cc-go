package clight

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
		{"int", Econst_int{Value: 42, Typ: ctypes.Int()}, "42"},
		{"negative int", Econst_int{Value: -5, Typ: ctypes.Int()}, "-5"},
		{"long", Econst_long{Value: 100, Typ: ctypes.Long()}, "100L"},
		{"float", Econst_float{Value: 3.14, Typ: ctypes.Double()}, "3.14"},
		{"single", Econst_single{Value: 1.5, Typ: ctypes.Float()}, "1.5f"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printExpr(tt.expr)
			got := buf.String()
			if got != tt.want {
				t.Errorf("printExpr() = %q, want %q", got, tt.want)
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
		{"var", Evar{Name: "x", Typ: ctypes.Int()}, "x"},
		{"tempvar", Etempvar{ID: 1, Typ: ctypes.Int()}, "$1"},
		{"tempvar 10", Etempvar{ID: 10, Typ: ctypes.Int()}, "$10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printExpr(tt.expr)
			got := buf.String()
			if got != tt.want {
				t.Errorf("printExpr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintExpr_UnaryOps(t *testing.T) {
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"deref", Ederef{Ptr: Evar{Name: "p", Typ: ctypes.Pointer(ctypes.Int())}, Typ: ctypes.Int()}, "*p"},
		{"addrof", Eaddrof{Arg: Evar{Name: "x", Typ: ctypes.Int()}, Typ: ctypes.Pointer(ctypes.Int())}, "&x"},
		{"neg", Eunop{Op: Oneg, Arg: Evar{Name: "x", Typ: ctypes.Int()}, Typ: ctypes.Int()}, "-x"},
		{"not", Eunop{Op: Onotbool, Arg: Evar{Name: "x", Typ: ctypes.Int()}, Typ: ctypes.Int()}, "!x"},
		{"bitnot", Eunop{Op: Onotint, Arg: Evar{Name: "x", Typ: ctypes.Int()}, Typ: ctypes.Int()}, "~x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printExpr(tt.expr)
			got := buf.String()
			if got != tt.want {
				t.Errorf("printExpr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintExpr_BinaryOps(t *testing.T) {
	x := Evar{Name: "x", Typ: ctypes.Int()}
	y := Evar{Name: "y", Typ: ctypes.Int()}

	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"add", Ebinop{Op: Oadd, Left: x, Right: y, Typ: ctypes.Int()}, "x + y"},
		{"sub", Ebinop{Op: Osub, Left: x, Right: y, Typ: ctypes.Int()}, "x - y"},
		{"mul", Ebinop{Op: Omul, Left: x, Right: y, Typ: ctypes.Int()}, "x * y"},
		{"div", Ebinop{Op: Odiv, Left: x, Right: y, Typ: ctypes.Int()}, "x / y"},
		{"eq", Ebinop{Op: Oeq, Left: x, Right: y, Typ: ctypes.Int()}, "x == y"},
		{"lt", Ebinop{Op: Olt, Left: x, Right: y, Typ: ctypes.Int()}, "x < y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printExpr(tt.expr)
			got := buf.String()
			if got != tt.want {
				t.Errorf("printExpr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintExpr_Special(t *testing.T) {
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"sizeof", Esizeof{ArgType: ctypes.Int(), Typ: ctypes.UInt()}, "sizeof(int)"},
		{"alignof", Ealignof{ArgType: ctypes.Int(), Typ: ctypes.UInt()}, "_Alignof(int)"},
		{"cast", Ecast{Arg: Evar{Name: "x", Typ: ctypes.Int()}, Typ: ctypes.Long()}, "(long)x"},
		{"field", Efield{Arg: Evar{Name: "s", Typ: ctypes.Tstruct{Name: "S"}}, FieldName: "x", Typ: ctypes.Int()}, "s.x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printExpr(tt.expr)
			got := buf.String()
			if got != tt.want {
				t.Errorf("printExpr() = %q, want %q", got, tt.want)
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
		{"break", Sbreak{}, "  break;\n"},
		{"continue", Scontinue{}, "  continue;\n"},
		{"return void", Sreturn{Value: nil}, "  return;\n"},
		{"return value", Sreturn{Value: Econst_int{Value: 0, Typ: ctypes.Int()}}, "  return 0;\n"},
		{"goto", Sgoto{Label: "done"}, "  goto done;\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.indent = 1
			p.printStmt(tt.stmt)
			got := buf.String()
			if got != tt.want {
				t.Errorf("printStmt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintStmt_Assign(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.indent = 1

	p.printStmt(Sassign{
		LHS: Evar{Name: "x", Typ: ctypes.Int()},
		RHS: Econst_int{Value: 5, Typ: ctypes.Int()},
	})

	want := "  x = 5;\n"
	got := buf.String()
	if got != want {
		t.Errorf("printStmt() = %q, want %q", got, want)
	}
}

func TestPrintStmt_Set(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.indent = 1

	p.printStmt(Sset{
		TempID: 1,
		RHS:    Econst_int{Value: 10, Typ: ctypes.Int()},
	})

	want := "  $1 = 10;\n"
	got := buf.String()
	if got != want {
		t.Errorf("printStmt() = %q, want %q", got, want)
	}
}

func TestPrintStmt_Call(t *testing.T) {
	tests := []struct {
		name string
		stmt Scall
		want string
	}{
		{
			"void call",
			Scall{
				Result: nil,
				Func:   Evar{Name: "f", Typ: ctypes.Int()},
				Args:   nil,
			},
			"  f();\n",
		},
		{
			"call with result",
			Scall{
				Result: intPtr(1),
				Func:   Evar{Name: "f", Typ: ctypes.Int()},
				Args:   nil,
			},
			"  $1 = f();\n",
		},
		{
			"call with args",
			Scall{
				Result: nil,
				Func:   Evar{Name: "g", Typ: ctypes.Int()},
				Args: []Expr{
					Econst_int{Value: 1, Typ: ctypes.Int()},
					Evar{Name: "x", Typ: ctypes.Int()},
				},
			},
			"  g(1, x);\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.indent = 1
			p.printStmt(tt.stmt)
			got := buf.String()
			if got != tt.want {
				t.Errorf("printStmt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintStmt_If(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.indent = 1

	p.printStmt(Sifthenelse{
		Cond: Evar{Name: "x", Typ: ctypes.Int()},
		Then: Sreturn{Value: Econst_int{Value: 1, Typ: ctypes.Int()}},
		Else: Sreturn{Value: Econst_int{Value: 0, Typ: ctypes.Int()}},
	})

	got := buf.String()
	if !strings.Contains(got, "if (x)") {
		t.Errorf("expected 'if (x)' in output: %s", got)
	}
	if !strings.Contains(got, "return 1") {
		t.Errorf("expected 'return 1' in output: %s", got)
	}
	if !strings.Contains(got, "else") {
		t.Errorf("expected 'else' in output: %s", got)
	}
	if !strings.Contains(got, "return 0") {
		t.Errorf("expected 'return 0' in output: %s", got)
	}
}

func TestPrintStmt_Sequence(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.indent = 1

	p.printStmt(Ssequence{
		First: Sset{TempID: 1, RHS: Econst_int{Value: 1, Typ: ctypes.Int()}},
		Second: Sreturn{Value: Etempvar{ID: 1, Typ: ctypes.Int()}},
	})

	got := buf.String()
	if !strings.Contains(got, "$1 = 1") {
		t.Errorf("expected '$1 = 1' in output: %s", got)
	}
	if !strings.Contains(got, "return $1") {
		t.Errorf("expected 'return $1' in output: %s", got)
	}
}

func TestPrintFunction(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)

	fn := Function{
		Name:   "add",
		Return: ctypes.Int(),
		Params: []VarDecl{
			{Name: "a", Type: ctypes.Int()},
			{Name: "b", Type: ctypes.Int()},
		},
		Locals: nil,
		Temps:  []ctypes.Type{ctypes.Int()},
		Body: Seq(
			Sset{TempID: 1, RHS: Ebinop{
				Op:    Oadd,
				Left:  Evar{Name: "a", Typ: ctypes.Int()},
				Right: Evar{Name: "b", Typ: ctypes.Int()},
				Typ:   ctypes.Int(),
			}},
			Sreturn{Value: Etempvar{ID: 1, Typ: ctypes.Int()}},
		),
	}

	p.printFunction(&fn)
	got := buf.String()

	// Check function signature
	if !strings.Contains(got, "int add(int a, int b)") {
		t.Errorf("expected function signature in output: %s", got)
	}

	// Check temp declaration
	if !strings.Contains(got, "int $1;") {
		t.Errorf("expected temp declaration in output: %s", got)
	}

	// Check body
	if !strings.Contains(got, "$1 = a + b") {
		t.Errorf("expected temp assignment in output: %s", got)
	}
	if !strings.Contains(got, "return $1") {
		t.Errorf("expected return statement in output: %s", got)
	}
}

func TestPrintProgram(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)

	prog := Program{
		Structs: []ctypes.Tstruct{
			{Name: "Point", Fields: []ctypes.Field{
				{Name: "x", Type: ctypes.Int()},
				{Name: "y", Type: ctypes.Int()},
			}},
		},
		Globals: []VarDecl{
			{Name: "count", Type: ctypes.Int()},
		},
		Functions: []Function{
			{
				Name:   "main",
				Return: ctypes.Int(),
				Body:   Sreturn{Value: Econst_int{Value: 0, Typ: ctypes.Int()}},
			},
		},
	}

	p.PrintProgram(&prog)
	got := buf.String()

	// Check struct definition
	if !strings.Contains(got, "struct Point") {
		t.Errorf("expected struct definition in output: %s", got)
	}

	// Check global
	if !strings.Contains(got, "int count;") {
		t.Errorf("expected global variable in output: %s", got)
	}

	// Check function
	if !strings.Contains(got, "int main()") {
		t.Errorf("expected main function in output: %s", got)
	}
}

func TestPrintStmt_Loop(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.indent = 1

	p.printStmt(Sloop{
		Body:     Sbreak{},
		Continue: Sskip{},
	})

	got := buf.String()
	if !strings.Contains(got, "loop {") {
		t.Errorf("expected 'loop {' in output: %s", got)
	}
	if !strings.Contains(got, "break;") {
		t.Errorf("expected 'break;' in output: %s", got)
	}
}

func TestPrintStmt_Switch(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.indent = 1

	p.printStmt(Sswitch{
		Expr: Evar{Name: "x", Typ: ctypes.Int()},
		Cases: []SwitchCase{
			{Value: 0, Body: Sreturn{Value: Econst_int{Value: 1, Typ: ctypes.Int()}}},
			{Value: 1, Body: Sreturn{Value: Econst_int{Value: 2, Typ: ctypes.Int()}}},
		},
		Default: Sreturn{Value: Econst_int{Value: 0, Typ: ctypes.Int()}},
	})

	got := buf.String()
	if !strings.Contains(got, "switch (x)") {
		t.Errorf("expected 'switch (x)' in output: %s", got)
	}
	if !strings.Contains(got, "case 0:") {
		t.Errorf("expected 'case 0:' in output: %s", got)
	}
	if !strings.Contains(got, "default:") {
		t.Errorf("expected 'default:' in output: %s", got)
	}
}

func TestPrintStmt_Label(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.indent = 1

	p.printStmt(Slabel{
		Label: "done",
		Stmt:  Sreturn{Value: Econst_int{Value: 0, Typ: ctypes.Int()}},
	})

	got := buf.String()
	if !strings.Contains(got, "done:") {
		t.Errorf("expected 'done:' in output: %s", got)
	}
	if !strings.Contains(got, "return 0") {
		t.Errorf("expected 'return 0' in output: %s", got)
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}
