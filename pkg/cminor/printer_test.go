package cminor

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintProgram(t *testing.T) {
	prog := &Program{
		Globals: []GlobVar{
			{Name: "g", Size: 4},
		},
		Functions: []Function{
			{
				Name:       "f",
				Sig:        Sig{Args: []string{"int"}, Return: "int"},
				Params:     []string{"x"},
				Vars:       []string{"y"},
				Stackspace: 8,
				Body: Sseq{
					First: Sassign{
						Name: "y",
						RHS: Ebinop{
							Op:    Oadd,
							Left:  Evar{Name: "x"},
							Right: Econst{Const: Ointconst{Value: 1}},
						},
					},
					Second: Sreturn{Value: Evar{Name: "y"}},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	// Check expected content
	expectations := []string{
		`var "g"[4];`,
		`"f"(x: int): int`,
		`stack 8;`,
		`var y;`,
		`y = add("x", 1);`,
		`return "y";`,
	}

	for _, exp := range expectations {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q\nGot:\n%s", exp, output)
		}
	}
}

func TestPrintSimpleFunction(t *testing.T) {
	prog := &Program{
		Functions: []Function{
			{
				Name:       "main",
				Sig:        Sig{Return: "int"},
				Params:     nil,
				Vars:       nil,
				Stackspace: 0,
				Body:       Sreturn{Value: Econst{Const: Ointconst{Value: 0}}},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	expectations := []string{
		`"main"(): int`,
		`return 0;`,
	}

	for _, exp := range expectations {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q\nGot:\n%s", exp, output)
		}
	}

	// Should NOT contain stack or var declarations
	if strings.Contains(output, "stack") {
		t.Errorf("expected output NOT to contain 'stack'\nGot:\n%s", output)
	}
}

func TestPrintIfThenElse(t *testing.T) {
	prog := &Program{
		Functions: []Function{
			{
				Name:   "abs",
				Sig:    Sig{Args: []string{"int"}, Return: "int"},
				Params: []string{"x"},
				Body: Sifthenelse{
					Cond: Ecmp{
						Op:    Ocmp,
						Cmp:   Clt,
						Left:  Evar{Name: "x"},
						Right: Econst{Const: Ointconst{Value: 0}},
					},
					Then: Sreturn{Value: Eunop{Op: Onegint, Arg: Evar{Name: "x"}}},
					Else: Sreturn{Value: Evar{Name: "x"}},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	expectations := []string{
		`"abs"(x: int): int`,
		`if (`,
		`} else {`,
		`return`,
	}

	for _, exp := range expectations {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q\nGot:\n%s", exp, output)
		}
	}
}

func TestPrintLoop(t *testing.T) {
	prog := &Program{
		Functions: []Function{
			{
				Name:   "forever",
				Sig:    Sig{Return: "void"},
				Params: nil,
				Body: Sloop{
					Body: Sskip{},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	if !strings.Contains(output, "loop {") {
		t.Errorf("expected output to contain 'loop {'\nGot:\n%s", output)
	}
}

func TestPrintBlock(t *testing.T) {
	prog := &Program{
		Functions: []Function{
			{
				Name:   "f",
				Sig:    Sig{Return: "void"},
				Params: nil,
				Body: Sblock{
					Body: Sexit{N: 1},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	expectations := []string{
		"block {",
		"exit 1;",
	}

	for _, exp := range expectations {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q\nGot:\n%s", exp, output)
		}
	}
}

func TestPrintSwitch(t *testing.T) {
	prog := &Program{
		Functions: []Function{
			{
				Name:   "classify",
				Sig:    Sig{Args: []string{"int"}, Return: "int"},
				Params: []string{"x"},
				Body: Sswitch{
					IsLong: false,
					Expr:   Evar{Name: "x"},
					Cases: []SwitchCase{
						{Value: 1, Body: Sreturn{Value: Econst{Const: Ointconst{Value: 10}}}},
						{Value: 2, Body: Sreturn{Value: Econst{Const: Ointconst{Value: 20}}}},
					},
					Default: Sreturn{Value: Econst{Const: Ointconst{Value: 0}}},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	expectations := []string{
		`switch ("x") {`,
		"case 1:",
		"return 10;",
		"case 2:",
		"return 20;",
		"default:",
		"return 0;",
	}

	for _, exp := range expectations {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q\nGot:\n%s", exp, output)
		}
	}
}

func TestPrintLongSwitch(t *testing.T) {
	prog := &Program{
		Functions: []Function{
			{
				Name:   "f",
				Sig:    Sig{Args: []string{"long"}, Return: "long"},
				Params: []string{"x"},
				Body: Sswitch{
					IsLong:  true,
					Expr:    Evar{Name: "x"},
					Cases:   nil,
					Default: Sreturn{Value: Econst{Const: Olongconst{Value: 0}}},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	if !strings.Contains(output, "switchl") {
		t.Errorf("expected output to contain 'switchl'\nGot:\n%s", output)
	}
}

func TestPrintCall(t *testing.T) {
	result := "_t0"
	prog := &Program{
		Functions: []Function{
			{
				Name:   "caller",
				Sig:    Sig{Return: "int"},
				Params: nil,
				Vars:   []string{"_t0"},
				Body: Sseq{
					First: Scall{
						Result: &result,
						Func:   Evar{Name: "callee"},
						Args: []Expr{
							Econst{Const: Ointconst{Value: 1}},
							Econst{Const: Ointconst{Value: 2}},
						},
					},
					Second: Sreturn{Value: Evar{Name: "_t0"}},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	expectations := []string{
		`_t0 = "callee"(1, 2);`,
		`return "_t0";`,
	}

	for _, exp := range expectations {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q\nGot:\n%s", exp, output)
		}
	}
}

func TestPrintTailcall(t *testing.T) {
	prog := &Program{
		Functions: []Function{
			{
				Name:   "f",
				Sig:    Sig{Args: []string{"int"}, Return: "int"},
				Params: []string{"x"},
				Body: Stailcall{
					Func: Evar{Name: "g"},
					Args: []Expr{Evar{Name: "x"}},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	if !strings.Contains(output, `tailcall "g"("x");`) {
		t.Errorf("expected output to contain tailcall\nGot:\n%s", output)
	}
}

func TestPrintStore(t *testing.T) {
	prog := &Program{
		Functions: []Function{
			{
				Name:   "f",
				Sig:    Sig{Args: []string{"int*"}, Return: "void"},
				Params: []string{"p"},
				Body: Sstore{
					Chunk: Mint32,
					Addr:  Evar{Name: "p"},
					Value: Econst{Const: Ointconst{Value: 42}},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	if !strings.Contains(output, `int32["p"] = 42;`) {
		t.Errorf("expected output to contain store\nGot:\n%s", output)
	}
}

func TestPrintLoad(t *testing.T) {
	prog := &Program{
		Functions: []Function{
			{
				Name:   "f",
				Sig:    Sig{Args: []string{"int*"}, Return: "int"},
				Params: []string{"p"},
				Body: Sreturn{
					Value: Eload{
						Chunk: Mint32,
						Addr:  Evar{Name: "p"},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	if !strings.Contains(output, `int32["p"]`) {
		t.Errorf("expected output to contain load\nGot:\n%s", output)
	}
}

func TestPrintGotoLabel(t *testing.T) {
	prog := &Program{
		Functions: []Function{
			{
				Name:   "f",
				Sig:    Sig{Return: "void"},
				Params: nil,
				Body: Slabel{
					Label: "loop",
					Body:  Sgoto{Label: "loop"},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	expectations := []string{
		"loop:",
		"goto loop;",
	}

	for _, exp := range expectations {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q\nGot:\n%s", exp, output)
		}
	}
}

func TestPrintBuiltin(t *testing.T) {
	result := "_t0"
	prog := &Program{
		Functions: []Function{
			{
				Name:   "f",
				Sig:    Sig{Return: "int"},
				Params: nil,
				Vars:   []string{"_t0"},
				Body: Sseq{
					First: Sbuiltin{
						Result:  &result,
						Builtin: "memcpy",
						Args:    []Expr{Econst{Const: Ointconst{Value: 0}}},
					},
					Second: Sreturn{Value: Evar{Name: "_t0"}},
				},
			},
		},
	}

	var buf bytes.Buffer
	printer := NewPrinter(&buf)
	printer.PrintProgram(prog)

	output := buf.String()

	if !strings.Contains(output, "__builtin_memcpy(0);") {
		t.Errorf("expected output to contain builtin\nGot:\n%s", output)
	}
}

func TestPrintConsts(t *testing.T) {
	tests := []struct {
		name   string
		expr   Expr
		expect string
	}{
		{"int", Econst{Const: Ointconst{Value: 42}}, "42"},
		{"long", Econst{Const: Olongconst{Value: 123}}, "123L"},
		{"float64", Econst{Const: Ofloatconst{Value: 3.14}}, "3.14"},
		{"float32", Econst{Const: Osingleconst{Value: 2.5}}, "2.5f"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			printer := NewPrinter(&buf)
			printer.printExpr(tc.expr)
			if !strings.Contains(buf.String(), tc.expect) {
				t.Errorf("expected %q, got %q", tc.expect, buf.String())
			}
		})
	}
}
