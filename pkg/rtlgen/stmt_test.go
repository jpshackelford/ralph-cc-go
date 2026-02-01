package rtlgen

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/cminorsel"
	"github.com/raymyers/ralph-cc/pkg/rtl"
)

func TestTranslateStmt_Skip(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	succ := cfg.AllocNode()
	entry := trans.TranslateStmt(cminorsel.Sskip{}, succ)
	
	if entry != succ {
		t.Errorf("Sskip: entry = %d, want succ=%d", entry, succ)
	}
}

func TestTranslateStmt_Assign(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	succ := cfg.AllocNode()
	
	// x = 42
	entry := trans.TranslateStmt(cminorsel.Sassign{
		Name: "x",
		RHS:  cminorsel.Econst{Const: cminorsel.Ointconst{Value: 42}},
	}, succ)
	
	// Should emit const load to x's register
	code := cfg.GetCode()
	if len(code) < 1 {
		t.Fatalf("expected at least 1 instruction, got %d", len(code))
	}
	
	// Entry should be const load
	instr := code[entry]
	iop, ok := instr.(rtl.Iop)
	if !ok {
		t.Fatalf("expected Iop, got %T", instr)
	}
	if _, ok := iop.Op.(rtl.Ointconst); !ok {
		t.Errorf("expected Ointconst, got %T", iop.Op)
	}
	
	// x should be mapped
	xReg, ok := regs.LookupVar("x")
	if !ok {
		t.Error("x should be mapped")
	}
	if iop.Dest != xReg {
		t.Errorf("dest = %d, want x=%d", iop.Dest, xReg)
	}
}

func TestTranslateStmt_Store(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	regs.MapVar("ptr")
	succ := cfg.AllocNode()
	
	// *ptr = 5
	entry := trans.TranslateStmt(cminorsel.Sstore{
		Chunk: cminorsel.Mint32,
		Mode:  cminorsel.Aindexed{Offset: 0},
		Args:  []cminorsel.Expr{cminorsel.Evar{Name: "ptr"}},
		Value: cminorsel.Econst{Const: cminorsel.Ointconst{Value: 5}},
	}, succ)
	
	// Should generate store instruction
	code := cfg.GetCode()
	foundStore := false
	for _, instr := range code {
		if _, ok := instr.(rtl.Istore); ok {
			foundStore = true
			break
		}
	}
	if !foundStore {
		t.Error("expected Istore instruction")
	}
	_ = entry
}

func TestTranslateStmt_Call(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	succ := cfg.AllocNode()
	resultName := "result"
	sig := &cminorsel.Sig{Args: []string{"int"}, Return: "int"}
	
	// result = foo(1)
	entry := trans.TranslateStmt(cminorsel.Scall{
		Result: &resultName,
		Sig:    sig,
		Func:   cminorsel.Econst{Const: cminorsel.Oaddrsymbol{Symbol: "foo", Offset: 0}},
		Args:   []cminorsel.Expr{cminorsel.Econst{Const: cminorsel.Ointconst{Value: 1}}},
	}, succ)
	
	// Should generate call instruction
	code := cfg.GetCode()
	foundCall := false
	for _, instr := range code {
		if icall, ok := instr.(rtl.Icall); ok {
			foundCall = true
			if fsym, ok := icall.Fn.(rtl.FunSymbol); ok {
				if fsym.Name != "foo" {
					t.Errorf("call target = %q, want %q", fsym.Name, "foo")
				}
			}
		}
	}
	if !foundCall {
		t.Error("expected Icall instruction")
	}
	_ = entry
}

func TestTranslateStmt_Return(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	// return 42
	entry := trans.TranslateStmt(cminorsel.Sreturn{
		Value: cminorsel.Econst{Const: cminorsel.Ointconst{Value: 42}},
	}, 0)
	
	// Should generate return instruction
	code := cfg.GetCode()
	foundReturn := false
	for _, instr := range code {
		if iret, ok := instr.(rtl.Ireturn); ok {
			foundReturn = true
			if iret.Arg == nil {
				t.Error("expected non-nil return arg")
			}
		}
	}
	if !foundReturn {
		t.Error("expected Ireturn instruction")
	}
	_ = entry
}

func TestTranslateStmt_ReturnVoid(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	// return
	entry := trans.TranslateStmt(cminorsel.Sreturn{Value: nil}, 0)
	
	// Should generate void return
	code := cfg.GetCode()
	instr := code[entry]
	iret, ok := instr.(rtl.Ireturn)
	if !ok {
		t.Fatalf("expected Ireturn, got %T", instr)
	}
	if iret.Arg != nil {
		t.Errorf("expected nil return arg, got %v", iret.Arg)
	}
}

func TestTranslateStmt_Seq(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	succ := cfg.AllocNode()
	
	// x = 1; y = 2
	entry := trans.TranslateStmt(cminorsel.Sseq{
		First: cminorsel.Sassign{
			Name: "x",
			RHS:  cminorsel.Econst{Const: cminorsel.Ointconst{Value: 1}},
		},
		Second: cminorsel.Sassign{
			Name: "y",
			RHS:  cminorsel.Econst{Const: cminorsel.Ointconst{Value: 2}},
		},
	}, succ)
	
	// Both x and y should be mapped
	if _, ok := regs.LookupVar("x"); !ok {
		t.Error("x should be mapped")
	}
	if _, ok := regs.LookupVar("y"); !ok {
		t.Error("y should be mapped")
	}
	
	// Should have at least 2 instructions
	code := cfg.GetCode()
	if len(code) < 2 {
		t.Errorf("expected at least 2 instructions, got %d", len(code))
	}
	_ = entry
}

func TestTranslateStmt_If(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	regs.MapVar("x")
	regs.MapVar("y")
	succ := cfg.AllocNode()
	
	// if (x < y) x = 1 else x = 2
	entry := trans.TranslateStmt(cminorsel.Sifthenelse{
		Cond: cminorsel.CondCmp{
			Cmp:   cminorsel.Clt,
			Left:  cminorsel.Evar{Name: "x"},
			Right: cminorsel.Evar{Name: "y"},
		},
		Then: cminorsel.Sassign{
			Name: "x",
			RHS:  cminorsel.Econst{Const: cminorsel.Ointconst{Value: 1}},
		},
		Else: cminorsel.Sassign{
			Name: "x",
			RHS:  cminorsel.Econst{Const: cminorsel.Ointconst{Value: 2}},
		},
	}, succ)
	
	// Should generate conditional branch
	code := cfg.GetCode()
	foundCond := false
	for _, instr := range code {
		if _, ok := instr.(rtl.Icond); ok {
			foundCond = true
		}
	}
	if !foundCond {
		t.Error("expected Icond instruction")
	}
	_ = entry
}

func TestTranslateStmt_Loop(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	succ := cfg.AllocNode()
	
	// loop { x = 1 }
	entry := trans.TranslateStmt(cminorsel.Sloop{
		Body: cminorsel.Sassign{
			Name: "x",
			RHS:  cminorsel.Econst{Const: cminorsel.Ointconst{Value: 1}},
		},
	}, succ)
	
	// Entry should be the loop header
	code := cfg.GetCode()
	headerInstr, ok := code[entry]
	if !ok {
		t.Fatal("header node should have instruction")
	}
	
	// Header should be nop pointing to body
	inop, ok := headerInstr.(rtl.Inop)
	if !ok {
		t.Fatalf("header expected Inop, got %T", headerInstr)
	}
	
	// Body should eventually loop back to header
	// (We'd need to follow the chain to verify)
	_ = inop
}

func TestTranslateStmt_Block(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	succ := cfg.AllocNode()
	
	// { x = 1 }
	entry := trans.TranslateStmt(cminorsel.Sblock{
		Body: cminorsel.Sassign{
			Name: "x",
			RHS:  cminorsel.Econst{Const: cminorsel.Ointconst{Value: 1}},
		},
	}, succ)
	
	// Should just translate the body
	code := cfg.GetCode()
	if len(code) < 1 {
		t.Error("expected at least 1 instruction")
	}
	_ = entry
}

func TestTranslateStmt_Exit(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	succ := cfg.AllocNode()
	
	// { exit(0) } - exit from block
	entry := trans.TranslateStmt(cminorsel.Sblock{
		Body: cminorsel.Sexit{N: 0},
	}, succ)
	
	// The exit should jump to succ
	code := cfg.GetCode()
	foundNop := false
	for _, instr := range code {
		if inop, ok := instr.(rtl.Inop); ok {
			if inop.Succ == succ {
				foundNop = true
			}
		}
	}
	if !foundNop {
		t.Error("expected nop jumping to succ (exit target)")
	}
	_ = entry
}

func TestTranslateStmt_Switch(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	succ := cfg.AllocNode()
	
	// switch(x) { case 1: y=1; default: y=0 }
	entry := trans.TranslateStmt(cminorsel.Sswitch{
		IsLong: false,
		Expr:   cminorsel.Evar{Name: "x"},
		Cases: []cminorsel.SwitchCase{
			{Value: 1, Body: cminorsel.Sassign{Name: "y", RHS: cminorsel.Econst{Const: cminorsel.Ointconst{Value: 1}}}},
		},
		Default: cminorsel.Sassign{Name: "y", RHS: cminorsel.Econst{Const: cminorsel.Ointconst{Value: 0}}},
	}, succ)
	
	// Should generate conditional branches
	code := cfg.GetCode()
	foundCond := false
	for _, instr := range code {
		if _, ok := instr.(rtl.Icond); ok {
			foundCond = true
		}
	}
	if !foundCond {
		t.Error("expected Icond instruction for switch")
	}
	_ = entry
}

func TestTranslateStmt_Label(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	succ := cfg.AllocNode()
	
	// L1: x = 1
	entry := trans.TranslateStmt(cminorsel.Slabel{
		Label: "L1",
		Body: cminorsel.Sassign{
			Name: "x",
			RHS:  cminorsel.Econst{Const: cminorsel.Ointconst{Value: 1}},
		},
	}, succ)
	
	// Label should be created
	labelNode, ok := cfg.GetLabel("L1")
	if !ok {
		t.Error("L1 label should be created")
	}
	if entry != labelNode {
		t.Errorf("entry = %d, want labelNode=%d", entry, labelNode)
	}
}

func TestTranslateStmt_Goto(t *testing.T) {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	trans := NewStmtTranslator(cfg, regs)
	
	// First create label
	labelNode := cfg.GetOrCreateLabel("L1")
	
	// goto L1
	entry := trans.TranslateStmt(cminorsel.Sgoto{Label: "L1"}, 0)
	
	// Should emit nop to label
	code := cfg.GetCode()
	instr := code[entry]
	inop, ok := instr.(rtl.Inop)
	if !ok {
		t.Fatalf("expected Inop, got %T", instr)
	}
	if inop.Succ != labelNode {
		t.Errorf("goto succ = %d, want labelNode=%d", inop.Succ, labelNode)
	}
}

func TestTranslateFunction(t *testing.T) {
	fn := cminorsel.Function{
		Name:       "foo",
		Sig:        cminorsel.Sig{Args: []string{"int"}, Return: "int"},
		Params:     []string{"x"},
		Vars:       []string{"y"},
		Stackspace: 16,
		Body: cminorsel.Sseq{
			First: cminorsel.Sassign{
				Name: "y",
				RHS:  cminorsel.Evar{Name: "x"},
			},
			Second: cminorsel.Sreturn{
				Value: cminorsel.Evar{Name: "y"},
			},
		},
	}
	
	rtlFn := TranslateFunction(fn)
	
	if rtlFn.Name != "foo" {
		t.Errorf("name = %q, want %q", rtlFn.Name, "foo")
	}
	if len(rtlFn.Params) != 1 {
		t.Errorf("len(params) = %d, want 1", len(rtlFn.Params))
	}
	if rtlFn.Stacksize != 16 {
		t.Errorf("stacksize = %d, want 16", rtlFn.Stacksize)
	}
	if len(rtlFn.Code) == 0 {
		t.Error("expected non-empty code")
	}
	if rtlFn.Entrypoint == 0 {
		t.Error("expected non-zero entrypoint")
	}
}

func TestTranslateProgram(t *testing.T) {
	prog := cminorsel.Program{
		Globals: []cminorsel.GlobVar{
			{Name: "g", Size: 4, Init: nil},
		},
		Functions: []cminorsel.Function{
			{
				Name:       "main",
				Sig:        cminorsel.Sig{Return: "int"},
				Params:     nil,
				Vars:       nil,
				Stackspace: 0,
				Body: cminorsel.Sreturn{
					Value: cminorsel.Econst{Const: cminorsel.Ointconst{Value: 0}},
				},
			},
		},
	}
	
	rtlProg := TranslateProgram(prog)
	
	if len(rtlProg.Globals) != 1 {
		t.Errorf("len(globals) = %d, want 1", len(rtlProg.Globals))
	}
	if rtlProg.Globals[0].Name != "g" {
		t.Errorf("global name = %q, want %q", rtlProg.Globals[0].Name, "g")
	}
	
	if len(rtlProg.Functions) != 1 {
		t.Errorf("len(functions) = %d, want 1", len(rtlProg.Functions))
	}
	if rtlProg.Functions[0].Name != "main" {
		t.Errorf("function name = %q, want %q", rtlProg.Functions[0].Name, "main")
	}
}
