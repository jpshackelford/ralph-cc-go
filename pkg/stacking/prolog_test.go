package stacking

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/linear"
	"github.com/raymyers/ralph-cc/pkg/ltl"
	"github.com/raymyers/ralph-cc/pkg/mach"
)

func TestGeneratePrologueEmpty(t *testing.T) {
	fn := linear.NewFunction("empty", linear.Sig{})
	layout := ComputeLayout(fn, 0)
	calleeSave := &CalleeSaveInfo{}

	prologue := GeneratePrologue(layout, calleeSave)

	// Should have at least frame allocation, FP/LR save, FP setup
	if len(prologue) < 3 {
		t.Errorf("expected at least 3 prologue instructions, got %d", len(prologue))
	}

	// First instruction should be frame allocation (Oaddlimm with negative value)
	if op, ok := prologue[0].(mach.Mop); ok {
		if addl, isAddlimm := op.Op.(mach.Oaddlimm); !isAddlimm {
			t.Errorf("first prologue inst should be Oaddlimm, got %T", op.Op)
		} else if addl.N >= 0 {
			t.Errorf("frame allocation should have negative immediate, got %d", addl.N)
		}
	} else {
		t.Errorf("first prologue inst should be Mop, got %T", prologue[0])
	}
}

func TestGeneratePrologueWithCalleeSave(t *testing.T) {
	fn := linear.NewFunction("withCalleeSave", linear.Sig{})
	layout := ComputeLayout(fn, 2)
	calleeSave := &CalleeSaveInfo{
		Regs:        []ltl.MReg{ltl.X19, ltl.X20},
		SaveOffsets: []int64{-16, -24},
	}

	prologue := GeneratePrologue(layout, calleeSave)

	// Count Msetstack instructions (FP, LR, plus 2 callee-save)
	setstackCount := 0
	for _, inst := range prologue {
		if _, ok := inst.(mach.Msetstack); ok {
			setstackCount++
		}
	}

	// Should have 4 Msetstack: FP, LR, X19, X20
	if setstackCount != 4 {
		t.Errorf("expected 4 Msetstack instructions, got %d", setstackCount)
	}
}

func TestGenerateEpilogueEmpty(t *testing.T) {
	fn := linear.NewFunction("empty", linear.Sig{})
	layout := ComputeLayout(fn, 0)
	calleeSave := &CalleeSaveInfo{}

	epilogue := GenerateEpilogue(layout, calleeSave)

	// Should end with Mreturn
	if len(epilogue) == 0 {
		t.Fatal("epilogue should not be empty")
	}
	if _, ok := epilogue[len(epilogue)-1].(mach.Mreturn); !ok {
		t.Errorf("epilogue should end with Mreturn, got %T", epilogue[len(epilogue)-1])
	}
}

func TestGenerateEpilogueWithCalleeSave(t *testing.T) {
	fn := linear.NewFunction("withCalleeSave", linear.Sig{})
	layout := ComputeLayout(fn, 2)
	calleeSave := &CalleeSaveInfo{
		Regs:        []ltl.MReg{ltl.X19, ltl.X20},
		SaveOffsets: []int64{-16, -24},
	}

	epilogue := GenerateEpilogue(layout, calleeSave)

	// Count Mgetstack instructions (FP, LR, plus 2 callee-save)
	getstackCount := 0
	for _, inst := range epilogue {
		if _, ok := inst.(mach.Mgetstack); ok {
			getstackCount++
		}
	}

	// Should have 4 Mgetstack: X19, X20, FP, LR
	if getstackCount != 4 {
		t.Errorf("expected 4 Mgetstack instructions, got %d", getstackCount)
	}

	// Should end with Mreturn
	if _, ok := epilogue[len(epilogue)-1].(mach.Mreturn); !ok {
		t.Errorf("epilogue should end with Mreturn")
	}
}

func TestGenerateTailEpilogue(t *testing.T) {
	fn := linear.NewFunction("tail", linear.Sig{})
	layout := ComputeLayout(fn, 0)
	calleeSave := &CalleeSaveInfo{}

	epilogue := GenerateTailEpilogue(layout, calleeSave)

	// Should NOT end with Mreturn
	if len(epilogue) > 0 {
		if _, ok := epilogue[len(epilogue)-1].(mach.Mreturn); ok {
			t.Error("tail epilogue should not end with Mreturn")
		}
	}
}

func TestIsLeafFunctionTrue(t *testing.T) {
	code := []mach.Instruction{
		mach.Mop{},
		mach.Mgetstack{},
		mach.Mreturn{},
	}

	if !IsLeafFunction(code) {
		t.Error("expected leaf function")
	}
}

func TestIsLeafFunctionFalseCall(t *testing.T) {
	code := []mach.Instruction{
		mach.Mop{},
		mach.Mcall{Fn: mach.FunSymbol{Name: "foo"}},
		mach.Mreturn{},
	}

	if IsLeafFunction(code) {
		t.Error("expected non-leaf function (has Mcall)")
	}
}

func TestIsLeafFunctionFalseTailcall(t *testing.T) {
	code := []mach.Instruction{
		mach.Mop{},
		mach.Mtailcall{Fn: mach.FunSymbol{Name: "foo"}},
	}

	if IsLeafFunction(code) {
		t.Error("expected non-leaf function (has Mtailcall)")
	}
}

func TestPrologueEpilogueSymmetry(t *testing.T) {
	fn := linear.NewFunction("sym", linear.Sig{})
	layout := ComputeLayout(fn, 2)
	calleeSave := &CalleeSaveInfo{
		Regs:        []ltl.MReg{ltl.X19, ltl.X20},
		SaveOffsets: []int64{-16, -24},
	}

	prologue := GeneratePrologue(layout, calleeSave)
	epilogue := GenerateEpilogue(layout, calleeSave)

	// Count saves and restores
	saveCount := 0
	restoreCount := 0
	for _, inst := range prologue {
		if _, ok := inst.(mach.Msetstack); ok {
			saveCount++
		}
	}
	for _, inst := range epilogue {
		if _, ok := inst.(mach.Mgetstack); ok {
			restoreCount++
		}
	}

	if saveCount != restoreCount {
		t.Errorf("save count (%d) != restore count (%d)", saveCount, restoreCount)
	}
}
