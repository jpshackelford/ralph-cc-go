package mach

import (
	"bytes"
	"strings"
	"testing"

	"github.com/raymyers/ralph-cc/pkg/ltl"
	"github.com/raymyers/ralph-cc/pkg/rtl"
)

func TestPrintFunctionEmpty(t *testing.T) {
	fn := NewFunction("empty", Sig{})
	fn.Append(Mreturn{})

	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "empty:") {
		t.Error("expected function name in output")
	}
	if !strings.Contains(out, "return") {
		t.Error("expected return in output")
	}
}

func TestPrintFunctionWithStacksize(t *testing.T) {
	fn := NewFunction("withStack", Sig{})
	fn.Stacksize = 32
	fn.Append(Mreturn{})

	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "32 bytes") {
		t.Errorf("expected stack frame size in output, got: %s", out)
	}
}

func TestPrintFunctionWithCalleeSave(t *testing.T) {
	fn := NewFunction("withCalleeSave", Sig{})
	fn.CalleeSaveRegs = []MReg{ltl.X19, ltl.X20}
	fn.Append(Mreturn{})

	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "X19") || !strings.Contains(out, "X20") {
		t.Errorf("expected callee-save regs in output, got: %s", out)
	}
}

func TestPrintMgetstack(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mgetstack{Ofs: -16, Ty: Tlong, Dest: ltl.X0})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "X0 = stack(-16)") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMsetstack(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Msetstack{Src: ltl.X0, Ofs: -8, Ty: Tint})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "stack(-8) = X0") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMgetparam(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mgetparam{Ofs: 16, Ty: Tlong, Dest: ltl.X1})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "X1 = param(16)") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMop(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mop{
		Op:   rtl.Oadd{},
		Args: []MReg{ltl.X0, ltl.X1},
		Dest: ltl.X2,
	})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "X2 = add X0, X1") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMopUnary(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mop{
		Op:   rtl.Oneg{},
		Args: []MReg{ltl.X0},
		Dest: ltl.X1,
	})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "X1 = neg X0") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMcall(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mcall{Sig: Sig{}, Fn: FunSymbol{Name: "callee"}})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "call callee") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMtailcall(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mtailcall{Sig: Sig{}, Fn: FunSymbol{Name: "target"}})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "tailcall target") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMlabel(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mlabel{Lbl: 42})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "42:") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMgoto(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mgoto{Target: 10})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "goto 10") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMcond(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mcond{
		Cond: rtl.Ccomp{Cond: rtl.Ceq},
		Args: []MReg{ltl.X0, ltl.X1},
		IfSo: 5,
	})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "if") && !strings.Contains(out, "goto 5") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMjumptable(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mjumptable{Arg: ltl.X0, Targets: []Label{1, 2, 3}})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "jumptable") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMload(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mload{
		Chunk: Mint64,
		Addr:  ltl.Aindexed{Offset: 0},
		Args:  []MReg{ltl.X1},
		Dest:  ltl.X0,
	})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "X0 =") && !strings.Contains(out, "int64") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintMstore(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf)
	fn := NewFunction("test", Sig{})
	fn.Append(Mstore{
		Chunk: Mint32,
		Addr:  ltl.Aindexed{Offset: 8},
		Args:  []MReg{ltl.X1},
		Src:   ltl.X0,
	})
	p.PrintFunction(fn)

	out := buf.String()
	if !strings.Contains(out, "= X0") && !strings.Contains(out, "int32") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintProgram(t *testing.T) {
	prog := &Program{
		Globals: []GlobVar{
			{Name: "x", Size: 8},
		},
		Functions: []Function{
			{Name: "main", Code: []Instruction{Mreturn{}}},
		},
	}

	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintProgram(prog)

	out := buf.String()
	if !strings.Contains(out, "\"x\"") {
		t.Errorf("expected global in output, got: %s", out)
	}
	if !strings.Contains(out, "main:") {
		t.Errorf("expected function in output, got: %s", out)
	}
}
