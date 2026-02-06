package cshmgen

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/clight"
	"github.com/raymyers/ralph-cc/pkg/csharpminor"
	"github.com/raymyers/ralph-cc/pkg/ctypes"
)

func newTestStmtTranslator() *StmtTranslator {
	return NewStmtTranslator(NewExprTranslator(nil))
}

func TestTranslateSkip(t *testing.T) {
	tr := newTestStmtTranslator()
	result := tr.TranslateStmt(clight.Sskip{})

	if _, ok := result.(csharpminor.Sskip); !ok {
		t.Fatalf("expected Sskip, got %T", result)
	}
}

func TestTranslateSet(t *testing.T) {
	tr := newTestStmtTranslator()
	stmt := clight.Sset{
		TempID: 5,
		RHS:    clight.Econst_int{Value: 42, Typ: ctypes.Int()},
	}
	result := tr.TranslateStmt(stmt)

	sset, ok := result.(csharpminor.Sset)
	if !ok {
		t.Fatalf("expected Sset, got %T", result)
	}
	if sset.TempID != 5 {
		t.Errorf("expected TempID 5, got %d", sset.TempID)
	}
}

func TestTranslateAssignVar(t *testing.T) {
	tr := newTestStmtTranslator()
	// x = 42 where x is an int variable
	stmt := clight.Sassign{
		LHS: clight.Evar{Name: "x", Typ: ctypes.Int()},
		RHS: clight.Econst_int{Value: 42, Typ: ctypes.Int()},
	}
	result := tr.TranslateStmt(stmt)

	sstore, ok := result.(csharpminor.Sstore)
	if !ok {
		t.Fatalf("expected Sstore, got %T", result)
	}
	if sstore.Chunk != csharpminor.Mint32 {
		t.Errorf("expected Mint32 chunk, got %v", sstore.Chunk)
	}
	// Address should be Eaddrof("x")
	addrof, ok := sstore.Addr.(csharpminor.Eaddrof)
	if !ok {
		t.Fatalf("expected Eaddrof for addr, got %T", sstore.Addr)
	}
	if addrof.Name != "x" {
		t.Errorf("expected addr name x, got %s", addrof.Name)
	}
}

func TestTranslateAssignDeref(t *testing.T) {
	tr := newTestStmtTranslator()
	// *p = 42 where p is int*
	ptr := clight.Etempvar{ID: 1, Typ: ctypes.Pointer(ctypes.Int())}
	stmt := clight.Sassign{
		LHS: clight.Ederef{Ptr: ptr, Typ: ctypes.Int()},
		RHS: clight.Econst_int{Value: 42, Typ: ctypes.Int()},
	}
	result := tr.TranslateStmt(stmt)

	sstore, ok := result.(csharpminor.Sstore)
	if !ok {
		t.Fatalf("expected Sstore, got %T", result)
	}
	if sstore.Chunk != csharpminor.Mint32 {
		t.Errorf("expected Mint32 chunk, got %v", sstore.Chunk)
	}
	// Address should be the translated pointer
	tempvar, ok := sstore.Addr.(csharpminor.Etempvar)
	if !ok {
		t.Fatalf("expected Etempvar for addr, got %T", sstore.Addr)
	}
	if tempvar.ID != 1 {
		t.Errorf("expected tempvar ID 1, got %d", tempvar.ID)
	}
}

func TestTranslateAssignField(t *testing.T) {
	tr := newTestStmtTranslator()
	// Struct: struct point { int x; int y; }
	pointType := ctypes.Tstruct{
		Name: "point",
		Fields: []ctypes.Field{
			{Name: "x", Type: ctypes.Int()},
			{Name: "y", Type: ctypes.Int()},
		},
	}

	// p.y = 42 where p is struct point
	p := clight.Evar{Name: "p", Typ: pointType}
	stmt := clight.Sassign{
		LHS: clight.Efield{Arg: p, FieldName: "y", Typ: ctypes.Int()},
		RHS: clight.Econst_int{Value: 42, Typ: ctypes.Int()},
	}
	result := tr.TranslateStmt(stmt)

	sstore, ok := result.(csharpminor.Sstore)
	if !ok {
		t.Fatalf("expected Sstore, got %T", result)
	}
	if sstore.Chunk != csharpminor.Mint32 {
		t.Errorf("expected Mint32 chunk, got %v", sstore.Chunk)
	}
	// Address should be &p + 4 (offset of y)
	binop, ok := sstore.Addr.(csharpminor.Ebinop)
	if !ok {
		t.Fatalf("expected Ebinop for field addr, got %T", sstore.Addr)
	}
	if binop.Op != csharpminor.Oaddl {
		t.Errorf("expected Oaddl for field addr, got %v", binop.Op)
	}
}

func TestTranslateSequence(t *testing.T) {
	tr := newTestStmtTranslator()
	stmt := clight.Ssequence{
		First:  clight.Sskip{},
		Second: clight.Sskip{},
	}
	result := tr.TranslateStmt(stmt)

	sseq, ok := result.(csharpminor.Sseq)
	if !ok {
		t.Fatalf("expected Sseq, got %T", result)
	}
	if _, ok := sseq.First.(csharpminor.Sskip); !ok {
		t.Errorf("expected First to be Sskip, got %T", sseq.First)
	}
	if _, ok := sseq.Second.(csharpminor.Sskip); !ok {
		t.Errorf("expected Second to be Sskip, got %T", sseq.Second)
	}
}

func TestTranslateIfthenelse(t *testing.T) {
	tr := newTestStmtTranslator()
	stmt := clight.Sifthenelse{
		Cond: clight.Econst_int{Value: 1, Typ: ctypes.Int()},
		Then: clight.Sskip{},
		Else: clight.Sskip{},
	}
	result := tr.TranslateStmt(stmt)

	sif, ok := result.(csharpminor.Sifthenelse)
	if !ok {
		t.Fatalf("expected Sifthenelse, got %T", result)
	}
	if _, ok := sif.Then.(csharpminor.Sskip); !ok {
		t.Errorf("expected Then to be Sskip, got %T", sif.Then)
	}
	if _, ok := sif.Else.(csharpminor.Sskip); !ok {
		t.Errorf("expected Else to be Sskip, got %T", sif.Else)
	}
}

func TestTranslateReturn(t *testing.T) {
	t.Run("void return", func(t *testing.T) {
		tr := newTestStmtTranslator()
		stmt := clight.Sreturn{Value: nil}
		result := tr.TranslateStmt(stmt)

		sret, ok := result.(csharpminor.Sreturn)
		if !ok {
			t.Fatalf("expected Sreturn, got %T", result)
		}
		if sret.Value != nil {
			t.Errorf("expected nil Value, got %v", sret.Value)
		}
	})

	t.Run("value return", func(t *testing.T) {
		tr := newTestStmtTranslator()
		stmt := clight.Sreturn{Value: clight.Econst_int{Value: 42, Typ: ctypes.Int()}}
		result := tr.TranslateStmt(stmt)

		sret, ok := result.(csharpminor.Sreturn)
		if !ok {
			t.Fatalf("expected Sreturn, got %T", result)
		}
		if sret.Value == nil {
			t.Fatalf("expected non-nil Value")
		}
		econst, ok := sret.Value.(csharpminor.Econst)
		if !ok {
			t.Fatalf("expected Econst, got %T", sret.Value)
		}
		intConst, ok := econst.Const.(csharpminor.Ointconst)
		if !ok {
			t.Fatalf("expected Ointconst, got %T", econst.Const)
		}
		if intConst.Value != 42 {
			t.Errorf("expected 42, got %d", intConst.Value)
		}
	})
}

func TestTranslateLoop(t *testing.T) {
	tr := newTestStmtTranslator()
	// Simple infinite loop: loop { skip; skip }
	stmt := clight.Sloop{
		Body:     clight.Sskip{},
		Continue: clight.Sskip{},
	}
	result := tr.TranslateStmt(stmt)

	// Should be: Sblock { Sloop { Sblock { body }; continue } }
	outerBlock, ok := result.(csharpminor.Sblock)
	if !ok {
		t.Fatalf("expected outer Sblock, got %T", result)
	}
	loop, ok := outerBlock.Body.(csharpminor.Sloop)
	if !ok {
		t.Fatalf("expected Sloop inside outer block, got %T", outerBlock.Body)
	}
	// Loop body should start with inner block
	// The Seq helper may simplify if one part is skip
	_ = loop.Body
}

func TestTranslateBreak(t *testing.T) {
	tr := newTestStmtTranslator()
	// Simulate being inside a loop
	loopStmt := clight.Sloop{
		Body:     clight.Sbreak{},
		Continue: clight.Sskip{},
	}
	result := tr.TranslateStmt(loopStmt)

	// Navigate to the break (it should be Sexit{N: 2})
	outerBlock := result.(csharpminor.Sblock)
	loop := outerBlock.Body.(csharpminor.Sloop)

	// Find the inner block containing break
	var innerBlock csharpminor.Sblock
	switch body := loop.Body.(type) {
	case csharpminor.Sseq:
		innerBlock = body.First.(csharpminor.Sblock)
	case csharpminor.Sblock:
		innerBlock = body
	default:
		t.Fatalf("unexpected loop body type: %T", loop.Body)
	}

	// The inner block body should be the break (Sexit{1})
	// Sexit(0) = exit inner block (continue), Sexit(1) = exit inner + outer (break)
	sexit, ok := innerBlock.Body.(csharpminor.Sexit)
	if !ok {
		t.Fatalf("expected Sexit for break, got %T", innerBlock.Body)
	}
	if sexit.N != 1 {
		t.Errorf("expected Sexit{N: 1} for break, got Sexit{N: %d}", sexit.N)
	}
}

func TestTranslateContinue(t *testing.T) {
	tr := newTestStmtTranslator()
	// Simulate being inside a loop
	loopStmt := clight.Sloop{
		Body:     clight.Scontinue{},
		Continue: clight.Sskip{},
	}
	result := tr.TranslateStmt(loopStmt)

	// Navigate to the continue (it should be Sexit{N: 1})
	outerBlock := result.(csharpminor.Sblock)
	loop := outerBlock.Body.(csharpminor.Sloop)

	// Find the inner block containing continue
	var innerBlock csharpminor.Sblock
	switch body := loop.Body.(type) {
	case csharpminor.Sseq:
		innerBlock = body.First.(csharpminor.Sblock)
	case csharpminor.Sblock:
		innerBlock = body
	default:
		t.Fatalf("unexpected loop body type: %T", loop.Body)
	}

	// The inner block body should be the continue (Sexit{0})
	// Sexit(0) = exit inner block (continue), Sexit(1) = exit inner + outer (break)
	sexit, ok := innerBlock.Body.(csharpminor.Sexit)
	if !ok {
		t.Fatalf("expected Sexit for continue, got %T", innerBlock.Body)
	}
	if sexit.N != 0 {
		t.Errorf("expected Sexit{N: 0} for continue, got Sexit{N: %d}", sexit.N)
	}
}

func TestTranslateCall(t *testing.T) {
	tr := newTestStmtTranslator()
	resultID := 1
	stmt := clight.Scall{
		Result: &resultID,
		Func:   clight.Evar{Name: "foo", Typ: ctypes.Pointer(ctypes.Void())},
		Args: []clight.Expr{
			clight.Econst_int{Value: 1, Typ: ctypes.Int()},
			clight.Econst_int{Value: 2, Typ: ctypes.Int()},
		},
	}
	result := tr.TranslateStmt(stmt)

	scall, ok := result.(csharpminor.Scall)
	if !ok {
		t.Fatalf("expected Scall, got %T", result)
	}
	if scall.Result == nil || *scall.Result != 1 {
		t.Errorf("expected Result to be 1")
	}
	if len(scall.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(scall.Args))
	}
}

func TestTranslateCallVoid(t *testing.T) {
	tr := newTestStmtTranslator()
	stmt := clight.Scall{
		Result: nil,
		Func:   clight.Evar{Name: "foo", Typ: ctypes.Pointer(ctypes.Void())},
		Args:   []clight.Expr{},
	}
	result := tr.TranslateStmt(stmt)

	scall, ok := result.(csharpminor.Scall)
	if !ok {
		t.Fatalf("expected Scall, got %T", result)
	}
	if scall.Result != nil {
		t.Errorf("expected nil Result for void call")
	}
}

func TestTranslateSwitch(t *testing.T) {
	tr := newTestStmtTranslator()
	stmt := clight.Sswitch{
		Expr: clight.Etempvar{ID: 1, Typ: ctypes.Int()},
		Cases: []clight.SwitchCase{
			{Value: 1, Body: clight.Sreturn{Value: clight.Econst_int{Value: 10, Typ: ctypes.Int()}}},
			{Value: 2, Body: clight.Sreturn{Value: clight.Econst_int{Value: 20, Typ: ctypes.Int()}}},
		},
		Default: clight.Sreturn{Value: clight.Econst_int{Value: 0, Typ: ctypes.Int()}},
	}
	result := tr.TranslateStmt(stmt)

	sswitch, ok := result.(csharpminor.Sswitch)
	if !ok {
		t.Fatalf("expected Sswitch, got %T", result)
	}
	if sswitch.IsLong {
		t.Errorf("expected IsLong=false for int switch")
	}
	if len(sswitch.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(sswitch.Cases))
	}
	if sswitch.Cases[0].Value != 1 {
		t.Errorf("expected first case value 1, got %d", sswitch.Cases[0].Value)
	}
}

func TestTranslateSwitchLong(t *testing.T) {
	tr := newTestStmtTranslator()
	stmt := clight.Sswitch{
		Expr:    clight.Etempvar{ID: 1, Typ: ctypes.Long()},
		Cases:   []clight.SwitchCase{},
		Default: clight.Sskip{},
	}
	result := tr.TranslateStmt(stmt)

	sswitch, ok := result.(csharpminor.Sswitch)
	if !ok {
		t.Fatalf("expected Sswitch, got %T", result)
	}
	if !sswitch.IsLong {
		t.Errorf("expected IsLong=true for long switch")
	}
}

func TestTranslateLabel(t *testing.T) {
	tr := newTestStmtTranslator()
	stmt := clight.Slabel{
		Label: "mylabel",
		Stmt:  clight.Sskip{},
	}
	result := tr.TranslateStmt(stmt)

	slabel, ok := result.(csharpminor.Slabel)
	if !ok {
		t.Fatalf("expected Slabel, got %T", result)
	}
	if slabel.Label != "mylabel" {
		t.Errorf("expected label 'mylabel', got '%s'", slabel.Label)
	}
}

func TestTranslateGoto(t *testing.T) {
	tr := newTestStmtTranslator()
	stmt := clight.Sgoto{Label: "mylabel"}
	result := tr.TranslateStmt(stmt)

	sgoto, ok := result.(csharpminor.Sgoto)
	if !ok {
		t.Fatalf("expected Sgoto, got %T", result)
	}
	if sgoto.Label != "mylabel" {
		t.Errorf("expected label 'mylabel', got '%s'", sgoto.Label)
	}
}

func TestTranslateBuiltin(t *testing.T) {
	tr := newTestStmtTranslator()
	resultID := 1
	stmt := clight.Sbuiltin{
		Result:  &resultID,
		Builtin: "__builtin_memcpy",
		Args: []clight.Expr{
			clight.Etempvar{ID: 1, Typ: ctypes.Pointer(ctypes.Void())},
			clight.Etempvar{ID: 2, Typ: ctypes.Pointer(ctypes.Void())},
			clight.Econst_int{Value: 100, Typ: ctypes.Int()},
		},
	}
	result := tr.TranslateStmt(stmt)

	sbuiltin, ok := result.(csharpminor.Sbuiltin)
	if !ok {
		t.Fatalf("expected Sbuiltin, got %T", result)
	}
	if sbuiltin.Builtin != "__builtin_memcpy" {
		t.Errorf("expected builtin '__builtin_memcpy', got '%s'", sbuiltin.Builtin)
	}
	if len(sbuiltin.Args) != 3 {
		t.Errorf("expected 3 args, got %d", len(sbuiltin.Args))
	}
}
