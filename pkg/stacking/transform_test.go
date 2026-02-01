package stacking

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/linear"
	"github.com/raymyers/ralph-cc/pkg/ltl"
	"github.com/raymyers/ralph-cc/pkg/mach"
	"github.com/raymyers/ralph-cc/pkg/rtl"
)

func TestTransformEmpty(t *testing.T) {
	fn := linear.NewFunction("empty", linear.Sig{})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	if machFn.Name != "empty" {
		t.Errorf("Name = %q, want 'empty'", machFn.Name)
	}
	// Should have prologue + epilogue
	if len(machFn.Code) < 4 {
		t.Errorf("expected at least 4 instructions (prologue+epilogue), got %d", len(machFn.Code))
	}
	// Should end with Mreturn
	if _, ok := machFn.Code[len(machFn.Code)-1].(mach.Mreturn); !ok {
		t.Errorf("expected last instruction to be Mreturn, got %T", machFn.Code[len(machFn.Code)-1])
	}
}

func TestTransformWithOp(t *testing.T) {
	fn := linear.NewFunction("add", linear.Sig{})
	fn.Append(linear.Lop{
		Op:   rtl.Oadd{},
		Args: []linear.Loc{linear.R{Reg: ltl.X0}, linear.R{Reg: ltl.X1}},
		Dest: linear.R{Reg: ltl.X2},
	})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Find the Mop instruction (skip prologue)
	found := false
	for _, inst := range machFn.Code {
		if op, ok := inst.(mach.Mop); ok {
			if _, isAdd := op.Op.(rtl.Oadd); isAdd {
				found = true
				if op.Dest != ltl.X2 {
					t.Errorf("Mop Dest = %v, want X2", op.Dest)
				}
				if len(op.Args) != 2 || op.Args[0] != ltl.X0 || op.Args[1] != ltl.X1 {
					t.Errorf("Mop Args = %v, want [X0, X1]", op.Args)
				}
				break
			}
		}
	}
	if !found {
		t.Error("expected to find Mop with Oadd")
	}
}

func TestTransformWithLabel(t *testing.T) {
	fn := linear.NewFunction("label", linear.Sig{})
	fn.Append(linear.Llabel{Lbl: 1})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Find the Mlabel instruction
	found := false
	for _, inst := range machFn.Code {
		if lbl, ok := inst.(mach.Mlabel); ok {
			if lbl.Lbl == 1 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected to find Mlabel with Lbl=1")
	}
}

func TestTransformWithGoto(t *testing.T) {
	fn := linear.NewFunction("goto", linear.Sig{})
	fn.Append(linear.Llabel{Lbl: 1})
	fn.Append(linear.Lgoto{Target: 1})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Find the Mgoto instruction
	found := false
	for _, inst := range machFn.Code {
		if gt, ok := inst.(mach.Mgoto); ok {
			if gt.Target == 1 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected to find Mgoto with Target=1")
	}
}

func TestTransformWithCond(t *testing.T) {
	fn := linear.NewFunction("cond", linear.Sig{})
	fn.Append(linear.Llabel{Lbl: 1})
	fn.Append(linear.Lcond{
		Cond: rtl.Ccomp{Cond: rtl.Ceq},
		Args: []linear.Loc{linear.R{Reg: ltl.X0}},
		IfSo: 1,
	})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Find the Mcond instruction
	found := false
	for _, inst := range machFn.Code {
		if cond, ok := inst.(mach.Mcond); ok {
			if cond.IfSo == 1 {
				found = true
				if len(cond.Args) != 1 || cond.Args[0] != ltl.X0 {
					t.Errorf("Mcond Args = %v, want [X0]", cond.Args)
				}
				break
			}
		}
	}
	if !found {
		t.Error("expected to find Mcond with IfSo=1")
	}
}

func TestTransformWithCall(t *testing.T) {
	fn := linear.NewFunction("caller", linear.Sig{})
	fn.Append(linear.Lcall{
		Sig: linear.Sig{},
		Fn:  linear.FunSymbol{Name: "callee"},
	})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Find the Mcall instruction
	found := false
	for _, inst := range machFn.Code {
		if call, ok := inst.(mach.Mcall); ok {
			if sym, isSym := call.Fn.(mach.FunSymbol); isSym && sym.Name == "callee" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected to find Mcall to 'callee'")
	}
}

func TestTransformWithTailcall(t *testing.T) {
	fn := linear.NewFunction("tailer", linear.Sig{})
	fn.Append(linear.Ltailcall{
		Sig: linear.Sig{},
		Fn:  linear.FunSymbol{Name: "target"},
	})

	machFn := Transform(fn)

	// Find the Mtailcall instruction
	found := false
	for _, inst := range machFn.Code {
		if tc, ok := inst.(mach.Mtailcall); ok {
			if sym, isSym := tc.Fn.(mach.FunSymbol); isSym && sym.Name == "target" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected to find Mtailcall to 'target'")
	}

	// Should NOT have Mreturn (tail call doesn't return)
	for _, inst := range machFn.Code {
		if _, ok := inst.(mach.Mreturn); ok {
			t.Error("tail call function should not have Mreturn")
			break
		}
	}
}

func TestTransformWithGetstack(t *testing.T) {
	fn := linear.NewFunction("getstack", linear.Sig{})
	fn.Append(linear.Lgetstack{
		Slot: linear.SlotLocal,
		Ofs:  0,
		Ty:   linear.Tlong,
		Dest: ltl.X0,
	})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Find a Mgetstack instruction (excluding prologue ones)
	getstackCount := 0
	for _, inst := range machFn.Code {
		if gs, ok := inst.(mach.Mgetstack); ok {
			if gs.Dest == ltl.X0 {
				getstackCount++
			}
		}
	}
	if getstackCount != 1 {
		t.Errorf("expected 1 Mgetstack to X0, got %d", getstackCount)
	}
}

func TestTransformWithSetstack(t *testing.T) {
	fn := linear.NewFunction("setstack", linear.Sig{})
	fn.Append(linear.Lsetstack{
		Src:  ltl.X0,
		Slot: linear.SlotLocal,
		Ofs:  0,
		Ty:   linear.Tlong,
	})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Find Msetstack instructions with X0 as source
	setstackCount := 0
	for _, inst := range machFn.Code {
		if ss, ok := inst.(mach.Msetstack); ok {
			if ss.Src == ltl.X0 {
				setstackCount++
			}
		}
	}
	if setstackCount != 1 {
		t.Errorf("expected 1 Msetstack from X0, got %d", setstackCount)
	}
}

func TestTransformWithLoad(t *testing.T) {
	fn := linear.NewFunction("load", linear.Sig{})
	fn.Append(linear.Lload{
		Chunk: linear.Mint64,
		Addr:  rtl.Aindexed{Offset: 0},
		Args:  []linear.Loc{linear.R{Reg: ltl.X1}},
		Dest:  linear.R{Reg: ltl.X0},
	})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Find Mload instruction
	found := false
	for _, inst := range machFn.Code {
		if ld, ok := inst.(mach.Mload); ok {
			if ld.Dest == ltl.X0 && ld.Chunk == linear.Mint64 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected to find Mload")
	}
}

func TestTransformWithStore(t *testing.T) {
	fn := linear.NewFunction("store", linear.Sig{})
	fn.Append(linear.Lstore{
		Chunk: linear.Mint32,
		Addr:  rtl.Aindexed{Offset: 0},
		Args:  []linear.Loc{linear.R{Reg: ltl.X1}},
		Src:   linear.R{Reg: ltl.X0},
	})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Find Mstore instruction
	found := false
	for _, inst := range machFn.Code {
		if st, ok := inst.(mach.Mstore); ok {
			if st.Src == ltl.X0 && st.Chunk == linear.Mint32 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected to find Mstore")
	}
}

func TestTransformWithJumptable(t *testing.T) {
	fn := linear.NewFunction("jumptable", linear.Sig{})
	fn.Append(linear.Llabel{Lbl: 1})
	fn.Append(linear.Llabel{Lbl: 2})
	fn.Append(linear.Ljumptable{
		Arg:     linear.R{Reg: ltl.X0},
		Targets: []linear.Label{1, 2},
	})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Find Mjumptable instruction
	found := false
	for _, inst := range machFn.Code {
		if jt, ok := inst.(mach.Mjumptable); ok {
			if jt.Arg == ltl.X0 && len(jt.Targets) == 2 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected to find Mjumptable")
	}
}

func TestTransformProgram(t *testing.T) {
	prog := &linear.Program{
		Globals: []linear.GlobVar{
			{Name: "x", Size: 8, Init: nil},
		},
		Functions: []linear.Function{
			{
				Name: "main",
				Sig:  linear.Sig{},
				Code: []linear.Instruction{linear.Lreturn{}},
			},
		},
	}

	machProg := TransformProgram(prog)

	if len(machProg.Globals) != 1 {
		t.Errorf("expected 1 global, got %d", len(machProg.Globals))
	}
	if machProg.Globals[0].Name != "x" {
		t.Errorf("global name = %q, want 'x'", machProg.Globals[0].Name)
	}
	if len(machProg.Functions) != 1 {
		t.Errorf("expected 1 function, got %d", len(machProg.Functions))
	}
	if machProg.Functions[0].Name != "main" {
		t.Errorf("function name = %q, want 'main'", machProg.Functions[0].Name)
	}
}

func TestTransformStacksize(t *testing.T) {
	fn := linear.NewFunction("withLocals", linear.Sig{})
	fn.Append(linear.Lgetstack{
		Slot: linear.SlotLocal,
		Ofs:  0,
		Ty:   linear.Tlong,
		Dest: ltl.X0,
	})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Stacksize should be set (at least 16 for FP/LR)
	if machFn.Stacksize < 16 {
		t.Errorf("Stacksize = %d, want >= 16", machFn.Stacksize)
	}
	// Stacksize should be 16-byte aligned
	if machFn.Stacksize%16 != 0 {
		t.Errorf("Stacksize %d is not 16-byte aligned", machFn.Stacksize)
	}
}

func TestTransformCalleeSaveRegs(t *testing.T) {
	fn := linear.NewFunction("usesCalleeSave", linear.Sig{})
	fn.Append(linear.Lop{
		Op:   rtl.Omove{},
		Args: []linear.Loc{linear.R{Reg: ltl.X19}},
		Dest: linear.R{Reg: ltl.X0},
	})
	fn.Append(linear.Lreturn{})

	machFn := Transform(fn)

	// Should have X19 in callee-save list (padded to even)
	if len(machFn.CalleeSaveRegs) == 0 {
		t.Error("expected callee-save regs to be populated")
	}
	foundX19 := false
	for _, reg := range machFn.CalleeSaveRegs {
		if reg == ltl.X19 {
			foundX19 = true
			break
		}
	}
	if !foundX19 {
		t.Errorf("expected X19 in CalleeSaveRegs, got %v", machFn.CalleeSaveRegs)
	}
}
