package regalloc

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/ltl"
	"github.com/raymyers/ralph-cc/pkg/rtl"
)

func TestTransformSimpleFunction(t *testing.T) {
	// Simple function: return 42
	rtlFn := &rtl.Function{
		Name: "return42",
		Sig:  rtl.Sig{},
		Code: map[rtl.Node]rtl.Instruction{
			1: rtl.Iop{Op: rtl.Ointconst{Value: 42}, Args: nil, Dest: 1, Succ: 2},
			2: rtl.Ireturn{Arg: ptr(rtl.Reg(1))},
		},
		Entrypoint: 1,
	}

	ltlFn := TransformFunction(rtlFn)

	if ltlFn.Name != "return42" {
		t.Errorf("name = %q, want %q", ltlFn.Name, "return42")
	}
	if ltlFn.Entrypoint != 1 {
		t.Errorf("entrypoint = %d, want 1", ltlFn.Entrypoint)
	}
	if len(ltlFn.Code) != 2 {
		t.Errorf("code has %d blocks, want 2", len(ltlFn.Code))
	}

	// Check block 1 has Lop + Lbranch
	block1 := ltlFn.Code[1]
	if block1 == nil {
		t.Fatal("block 1 should exist")
	}
	if len(block1.Body) != 2 {
		t.Errorf("block 1 has %d instructions, want 2", len(block1.Body))
	}
	if _, ok := block1.Body[0].(ltl.Lop); !ok {
		t.Errorf("block 1 first instr should be Lop, got %T", block1.Body[0])
	}
	if br, ok := block1.Body[1].(ltl.Lbranch); !ok || br.Succ != 2 {
		t.Error("block 1 should end with branch to 2")
	}

	// Check block 2 has Lreturn
	block2 := ltlFn.Code[2]
	if block2 == nil {
		t.Fatal("block 2 should exist")
	}
	if _, ok := block2.Body[0].(ltl.Lreturn); !ok {
		t.Errorf("block 2 should have Lreturn, got %T", block2.Body[0])
	}
}

func TestTransformFunctionWithAdd(t *testing.T) {
	// Function: add two values
	rtlFn := &rtl.Function{
		Name:   "add",
		Sig:    rtl.Sig{},
		Params: []rtl.Reg{1, 2},
		Code: map[rtl.Node]rtl.Instruction{
			1: rtl.Iop{Op: rtl.Oadd{}, Args: []rtl.Reg{1, 2}, Dest: 3, Succ: 2},
			2: rtl.Ireturn{Arg: ptr(rtl.Reg(3))},
		},
		Entrypoint: 1,
	}

	ltlFn := TransformFunction(rtlFn)

	// Check that the Lop has proper locations
	block1 := ltlFn.Code[1]
	if block1 == nil {
		t.Fatal("block 1 should exist")
	}

	lop, ok := block1.Body[0].(ltl.Lop)
	if !ok {
		t.Fatalf("expected Lop, got %T", block1.Body[0])
	}

	// Should have 2 args
	if len(lop.Args) != 2 {
		t.Errorf("Lop has %d args, want 2", len(lop.Args))
	}

	// Dest should be a location
	if lop.Dest == nil {
		t.Error("Lop dest should not be nil")
	}
}

func TestTransformFunctionWithConditional(t *testing.T) {
	rtlFn := &rtl.Function{
		Name: "cond",
		Code: map[rtl.Node]rtl.Instruction{
			1: rtl.Iop{Op: rtl.Ointconst{Value: 1}, Args: nil, Dest: 1, Succ: 2},
			2: rtl.Icond{
				Cond:  rtl.Ccompimm{Cond: rtl.Ceq, N: 0},
				Args:  []rtl.Reg{1},
				IfSo:  3,
				IfNot: 4,
			},
			3: rtl.Ireturn{Arg: ptr(rtl.Reg(1))},
			4: rtl.Ireturn{Arg: ptr(rtl.Reg(1))},
		},
		Entrypoint: 1,
	}

	ltlFn := TransformFunction(rtlFn)

	// Check block 2 has Lcond
	block2 := ltlFn.Code[2]
	if block2 == nil {
		t.Fatal("block 2 should exist")
	}

	lcond, ok := block2.Body[0].(ltl.Lcond)
	if !ok {
		t.Fatalf("expected Lcond, got %T", block2.Body[0])
	}

	if lcond.IfSo != 3 || lcond.IfNot != 4 {
		t.Errorf("Lcond targets = %d/%d, want 3/4", lcond.IfSo, lcond.IfNot)
	}
}

func TestTransformFunctionWithLoad(t *testing.T) {
	rtlFn := &rtl.Function{
		Name: "load",
		Code: map[rtl.Node]rtl.Instruction{
			1: rtl.Iop{Op: rtl.Ointconst{Value: 0}, Args: nil, Dest: 1, Succ: 2},
			2: rtl.Iload{
				Chunk: rtl.Mint64,
				Addr:  rtl.Aindexed{Offset: 0},
				Args:  []rtl.Reg{1},
				Dest:  2,
				Succ:  3,
			},
			3: rtl.Ireturn{Arg: ptr(rtl.Reg(2))},
		},
		Entrypoint: 1,
	}

	ltlFn := TransformFunction(rtlFn)

	block2 := ltlFn.Code[2]
	if block2 == nil {
		t.Fatal("block 2 should exist")
	}

	lload, ok := block2.Body[0].(ltl.Lload)
	if !ok {
		t.Fatalf("expected Lload, got %T", block2.Body[0])
	}

	if lload.Chunk != ltl.Mint64 {
		t.Errorf("Lload chunk = %v, want Mint64", lload.Chunk)
	}
}

func TestTransformFunctionWithStore(t *testing.T) {
	rtlFn := &rtl.Function{
		Name: "store",
		Code: map[rtl.Node]rtl.Instruction{
			1: rtl.Iop{Op: rtl.Ointconst{Value: 0}, Args: nil, Dest: 1, Succ: 2},
			2: rtl.Iop{Op: rtl.Ointconst{Value: 42}, Args: nil, Dest: 2, Succ: 3},
			3: rtl.Istore{
				Chunk: rtl.Mint64,
				Addr:  rtl.Aindexed{Offset: 0},
				Args:  []rtl.Reg{1},
				Src:   2,
				Succ:  4,
			},
			4: rtl.Ireturn{Arg: nil},
		},
		Entrypoint: 1,
	}

	ltlFn := TransformFunction(rtlFn)

	block3 := ltlFn.Code[3]
	if block3 == nil {
		t.Fatal("block 3 should exist")
	}

	lstore, ok := block3.Body[0].(ltl.Lstore)
	if !ok {
		t.Fatalf("expected Lstore, got %T", block3.Body[0])
	}

	if lstore.Chunk != ltl.Mint64 {
		t.Errorf("Lstore chunk = %v, want Mint64", lstore.Chunk)
	}
}

func TestTransformFunctionWithCall(t *testing.T) {
	rtlFn := &rtl.Function{
		Name: "caller",
		Code: map[rtl.Node]rtl.Instruction{
			1: rtl.Iop{Op: rtl.Ointconst{Value: 1}, Args: nil, Dest: 1, Succ: 2},
			2: rtl.Icall{
				Sig:  rtl.Sig{},
				Fn:   rtl.FunSymbol{Name: "callee"},
				Args: []rtl.Reg{1},
				Dest: 2,
				Succ: 3,
			},
			3: rtl.Ireturn{Arg: ptr(rtl.Reg(2))},
		},
		Entrypoint: 1,
	}

	ltlFn := TransformFunction(rtlFn)

	block2 := ltlFn.Code[2]
	if block2 == nil {
		t.Fatal("block 2 should exist")
	}

	lcall, ok := block2.Body[0].(ltl.Lcall)
	if !ok {
		t.Fatalf("expected Lcall, got %T", block2.Body[0])
	}

	if sym, ok := lcall.Fn.(ltl.FunSymbol); !ok || sym.Name != "callee" {
		t.Error("Lcall should call 'callee'")
	}
}

func TestTransformFunctionWithJumptable(t *testing.T) {
	rtlFn := &rtl.Function{
		Name: "switch",
		Code: map[rtl.Node]rtl.Instruction{
			1: rtl.Iop{Op: rtl.Ointconst{Value: 0}, Args: nil, Dest: 1, Succ: 2},
			2: rtl.Ijumptable{
				Arg:     1,
				Targets: []rtl.Node{3, 4, 5},
			},
			3: rtl.Ireturn{Arg: nil},
			4: rtl.Ireturn{Arg: nil},
			5: rtl.Ireturn{Arg: nil},
		},
		Entrypoint: 1,
	}

	ltlFn := TransformFunction(rtlFn)

	block2 := ltlFn.Code[2]
	if block2 == nil {
		t.Fatal("block 2 should exist")
	}

	ljump, ok := block2.Body[0].(ltl.Ljumptable)
	if !ok {
		t.Fatalf("expected Ljumptable, got %T", block2.Body[0])
	}

	if len(ljump.Targets) != 3 {
		t.Errorf("jumptable has %d targets, want 3", len(ljump.Targets))
	}
}

func TestTransformProgram(t *testing.T) {
	rtlProg := &rtl.Program{
		Globals: []rtl.GlobVar{
			{Name: "g", Size: 8},
		},
		Functions: []rtl.Function{
			{
				Name: "main",
				Code: map[rtl.Node]rtl.Instruction{
					1: rtl.Ireturn{Arg: nil},
				},
				Entrypoint: 1,
			},
		},
	}

	ltlProg := TransformProgram(rtlProg)

	if len(ltlProg.Globals) != 1 {
		t.Errorf("program has %d globals, want 1", len(ltlProg.Globals))
	}
	if ltlProg.Globals[0].Name != "g" {
		t.Error("global should be named 'g'")
	}

	if len(ltlProg.Functions) != 1 {
		t.Errorf("program has %d functions, want 1", len(ltlProg.Functions))
	}
	if ltlProg.Functions[0].Name != "main" {
		t.Error("function should be named 'main'")
	}
}

func TestTransformNop(t *testing.T) {
	rtlFn := &rtl.Function{
		Name: "nop",
		Code: map[rtl.Node]rtl.Instruction{
			1: rtl.Inop{Succ: 2},
			2: rtl.Ireturn{Arg: nil},
		},
		Entrypoint: 1,
	}

	ltlFn := TransformFunction(rtlFn)

	block1 := ltlFn.Code[1]
	if block1 == nil {
		t.Fatal("block 1 should exist")
	}

	// Nop becomes just a branch
	if br, ok := block1.Body[0].(ltl.Lbranch); !ok || br.Succ != 2 {
		t.Error("nop should become branch to successor")
	}
}
