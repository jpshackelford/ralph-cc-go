// Statement translation for RTLgen.
// Translates CminorSel statements to RTL CFG.
// Uses backward chaining: statements chain to a successor node.

package rtlgen

import (
	"github.com/raymyers/ralph-cc/pkg/cminorsel"
	"github.com/raymyers/ralph-cc/pkg/rtl"
)

// StmtTranslator translates CminorSel statements to RTL CFG.
type StmtTranslator struct {
	cfg  *CFGBuilder
	regs *RegAllocator
	expr *ExprTranslator
	ctx  *ExitContext
}

// NewStmtTranslator creates a statement translator.
func NewStmtTranslator(cfg *CFGBuilder, regs *RegAllocator) *StmtTranslator {
	return &StmtTranslator{
		cfg:  cfg,
		regs: regs,
		expr: NewExprTranslator(cfg, regs),
		ctx:  NewExitContext(),
	}
}

// TranslateStmt translates a statement to RTL instructions.
// Returns the entry node of the translated code.
// Instructions chain backward to succ.
func (t *StmtTranslator) TranslateStmt(s cminorsel.Stmt, succ rtl.Node) rtl.Node {
	switch stmt := s.(type) {
	case cminorsel.Sskip:
		return succ
	case cminorsel.Sassign:
		return t.translateAssign(stmt, succ)
	case cminorsel.Sstore:
		return t.translateStore(stmt, succ)
	case cminorsel.Scall:
		return t.translateCall(stmt, succ)
	case cminorsel.Stailcall:
		return t.translateTailcall(stmt)
	case cminorsel.Sbuiltin:
		return t.translateBuiltin(stmt, succ)
	case cminorsel.Sseq:
		return t.translateSeq(stmt, succ)
	case cminorsel.Sifthenelse:
		return t.translateIf(stmt, succ)
	case cminorsel.Sloop:
		return t.translateLoop(stmt, succ)
	case cminorsel.Sblock:
		return t.translateBlock(stmt, succ)
	case cminorsel.Sexit:
		return t.translateExit(stmt)
	case cminorsel.Sswitch:
		return t.translateSwitch(stmt, succ)
	case cminorsel.Sreturn:
		return t.translateReturn(stmt)
	case cminorsel.Slabel:
		return t.translateLabel(stmt, succ)
	case cminorsel.Sgoto:
		return t.translateGoto(stmt)
	default:
		return succ
	}
}

func (t *StmtTranslator) translateAssign(s cminorsel.Sassign, succ rtl.Node) rtl.Node {
	// Assign RHS to variable's register
	dest := t.regs.MapVar(s.Name)
	return t.expr.TranslateExpr(s.RHS, dest, succ)
}

func (t *StmtTranslator) translateStore(s cminorsel.Sstore, succ rtl.Node) rtl.Node {
	// Evaluate value and address, then store
	valueReg := t.regs.Fresh()
	addrRegs := make([]rtl.Reg, len(s.Args))
	for i := range s.Args {
		addrRegs[i] = t.regs.Fresh()
	}
	
	// Emit store instruction
	addr := TranslateAddressingMode(s.Mode)
	chunk := TranslateChunk(s.Chunk)
	storeNode := t.cfg.EmitInstr(rtl.Istore{
		Chunk: chunk,
		Addr:  addr,
		Args:  addrRegs,
		Src:   valueReg,
		Succ:  succ,
	})
	
	// Translate value expression -> store
	valueEntry := t.expr.TranslateExpr(s.Value, valueReg, storeNode)
	
	// Translate address expressions -> value
	entry := valueEntry
	for i := len(s.Args) - 1; i >= 0; i-- {
		entry = t.expr.TranslateExpr(s.Args[i], addrRegs[i], entry)
	}
	
	return entry
}

func (t *StmtTranslator) translateCall(s cminorsel.Scall, succ rtl.Node) rtl.Node {
	// Evaluate function and arguments, then call
	funcReg := t.regs.Fresh()
	argRegs := make([]rtl.Reg, len(s.Args))
	for i := range s.Args {
		argRegs[i] = t.regs.Fresh()
	}
	
	// Destination register (0 for void)
	var destReg rtl.Reg
	if s.Result != nil {
		destReg = t.regs.MapVar(*s.Result)
	}
	
	// Build signature
	var sig rtl.Sig
	if s.Sig != nil {
		sig = rtl.Sig{
			Args:   s.Sig.Args,
			Return: s.Sig.Return,
			VarArg: s.Sig.VarArg,
		}
	}
	
	// Determine function reference - could be symbol or register
	var fnRef rtl.FunRef
	
	// Emit call instruction
	callNode := t.cfg.EmitInstr(rtl.Icall{
		Sig:  sig,
		Fn:   rtl.FunReg{Reg: funcReg}, // Will be updated below if symbol
		Args: argRegs,
		Dest: destReg,
		Succ: succ,
	})
	
	// Check if function is a direct call to a symbol
	if econst, ok := s.Func.(cminorsel.Econst); ok {
		if sym, ok := econst.Const.(cminorsel.Oaddrsymbol); ok {
			fnRef = rtl.FunSymbol{Name: sym.Symbol}
			// Re-emit with symbol
			t.cfg.AddInstr(callNode, rtl.Icall{
				Sig:  sig,
				Fn:   fnRef,
				Args: argRegs,
				Dest: destReg,
				Succ: succ,
			})
			// No need to evaluate function expression
			return t.translateExprList(s.Args, argRegs, callNode)
		}
	}
	
	// Indirect call - evaluate function address
	funcEntry := t.expr.TranslateExpr(s.Func, funcReg, callNode)
	
	// Translate arguments -> function
	return t.translateExprListChain(s.Args, argRegs, funcEntry)
}

func (t *StmtTranslator) translateTailcall(s cminorsel.Stailcall) rtl.Node {
	// Evaluate function and arguments, then tail call
	funcReg := t.regs.Fresh()
	argRegs := make([]rtl.Reg, len(s.Args))
	for i := range s.Args {
		argRegs[i] = t.regs.Fresh()
	}
	
	var sig rtl.Sig
	if s.Sig != nil {
		sig = rtl.Sig{
			Args:   s.Sig.Args,
			Return: s.Sig.Return,
			VarArg: s.Sig.VarArg,
		}
	}
	
	// Emit tail call (no successor)
	tailNode := t.cfg.EmitInstr(rtl.Itailcall{
		Sig:  sig,
		Fn:   rtl.FunReg{Reg: funcReg},
		Args: argRegs,
	})
	
	// Check for direct call
	if econst, ok := s.Func.(cminorsel.Econst); ok {
		if sym, ok := econst.Const.(cminorsel.Oaddrsymbol); ok {
			t.cfg.AddInstr(tailNode, rtl.Itailcall{
				Sig:  sig,
				Fn:   rtl.FunSymbol{Name: sym.Symbol},
				Args: argRegs,
			})
			return t.translateExprList(s.Args, argRegs, tailNode)
		}
	}
	
	funcEntry := t.expr.TranslateExpr(s.Func, funcReg, tailNode)
	return t.translateExprListChain(s.Args, argRegs, funcEntry)
}

func (t *StmtTranslator) translateBuiltin(s cminorsel.Sbuiltin, succ rtl.Node) rtl.Node {
	argRegs := make([]rtl.Reg, len(s.Args))
	for i := range s.Args {
		argRegs[i] = t.regs.Fresh()
	}
	
	var destPtr *rtl.Reg
	if s.Result != nil {
		dest := t.regs.MapVar(*s.Result)
		destPtr = &dest
	}
	
	builtinNode := t.cfg.EmitInstr(rtl.Ibuiltin{
		Builtin: s.Builtin,
		Args:    argRegs,
		Dest:    destPtr,
		Succ:    succ,
	})
	
	return t.translateExprList(s.Args, argRegs, builtinNode)
}

func (t *StmtTranslator) translateSeq(s cminorsel.Sseq, succ rtl.Node) rtl.Node {
	// Execute first, then second
	// With backward chaining: second -> succ, first -> second
	secondEntry := t.TranslateStmt(s.Second, succ)
	return t.TranslateStmt(s.First, secondEntry)
}

func (t *StmtTranslator) translateIf(s cminorsel.Sifthenelse, succ rtl.Node) rtl.Node {
	// Translate both branches -> succ
	thenEntry := t.TranslateStmt(s.Then, succ)
	elseEntry := t.TranslateStmt(s.Else, succ)
	
	// Translate condition -> branches
	return t.expr.TranslateCond(s.Cond, thenEntry, elseEntry)
}

func (t *StmtTranslator) translateLoop(s cminorsel.Sloop, succ rtl.Node) rtl.Node {
	// Loop structure:
	//   header: body -> header (back edge)
	// Exit is via Sexit which jumps past the loop
	
	// Allocate header node
	header := t.cfg.AllocNode()
	
	// Push exit target (for Sexit(0) to jump past loop)
	t.ctx.Push(succ)
	
	// Translate body -> header (back edge)
	bodyEntry := t.TranslateStmt(s.Body, header)
	
	t.ctx.Pop()
	
	// Header instruction: nop -> body
	// Actually, header IS the body entry
	// We need: header -> body -> back to header
	// With our backward chaining, bodyEntry is where body starts
	
	// If body is empty (Sskip), bodyEntry == header (back edge)
	// Add nop at header to jump to body
	t.cfg.AddInstr(header, rtl.Inop{Succ: bodyEntry})
	
	return header
}

func (t *StmtTranslator) translateBlock(s cminorsel.Sblock, succ rtl.Node) rtl.Node {
	// Block provides exit target for Sexit
	t.ctx.Push(succ)
	entry := t.TranslateStmt(s.Body, succ)
	t.ctx.Pop()
	return entry
}

func (t *StmtTranslator) translateExit(s cminorsel.Sexit) rtl.Node {
	// Jump to the N-th enclosing block's exit
	target, ok := t.ctx.Get(s.N)
	if !ok {
		// Invalid exit - should not happen
		return t.cfg.AllocNode()
	}
	
	// Emit unconditional jump (nop) to target
	return t.cfg.EmitInstr(rtl.Inop{Succ: target})
}

func (t *StmtTranslator) translateSwitch(s cminorsel.Sswitch, succ rtl.Node) rtl.Node {
	// Switch on expression value
	// For now, implement as cascading if-else
	// A proper implementation would use jump tables
	
	exprReg := t.regs.Fresh()
	
	// Translate default -> succ
	defaultEntry := t.TranslateStmt(s.Default, succ)
	
	// Build cascading conditions for cases
	currentElse := defaultEntry
	for i := len(s.Cases) - 1; i >= 0; i-- {
		c := s.Cases[i]
		
		// Translate case body -> succ
		caseEntry := t.TranslateStmt(c.Body, succ)
		
		// Compare expr == case value
		// If true, go to case; else continue to next case
		var cond rtl.ConditionCode
		if s.IsLong {
			cond = rtl.Ccomplimm{Cond: rtl.Ceq, N: c.Value}
		} else {
			cond = rtl.Ccompimm{Cond: rtl.Ceq, N: int32(c.Value)}
		}
		
		condNode := t.cfg.EmitInstr(rtl.Icond{
			Cond:  cond,
			Args:  []rtl.Reg{exprReg},
			IfSo:  caseEntry,
			IfNot: currentElse,
		})
		
		currentElse = condNode
	}
	
	// Translate expression -> first condition
	return t.expr.TranslateExpr(s.Expr, exprReg, currentElse)
}

func (t *StmtTranslator) translateReturn(s cminorsel.Sreturn) rtl.Node {
	if s.Value == nil {
		// Void return
		return t.cfg.EmitInstr(rtl.Ireturn{Arg: nil})
	}
	
	// Return with value
	retReg := t.regs.Fresh()
	retNode := t.cfg.EmitInstr(rtl.Ireturn{Arg: &retReg})
	return t.expr.TranslateExpr(s.Value, retReg, retNode)
}

func (t *StmtTranslator) translateLabel(s cminorsel.Slabel, succ rtl.Node) rtl.Node {
	// Get or create node for label
	labelNode := t.cfg.GetOrCreateLabel(s.Label)
	
	// Translate body -> succ
	bodyEntry := t.TranslateStmt(s.Body, succ)
	
	// Label node jumps to body
	t.cfg.AddInstr(labelNode, rtl.Inop{Succ: bodyEntry})
	
	return labelNode
}

func (t *StmtTranslator) translateGoto(s cminorsel.Sgoto) rtl.Node {
	// Get or create node for target label
	targetNode := t.cfg.GetOrCreateLabel(s.Label)
	
	// Jump to target
	return t.cfg.EmitInstr(rtl.Inop{Succ: targetNode})
}

// Helper: translate expressions into specific registers
func (t *StmtTranslator) translateExprList(exprs []cminorsel.Expr, regs []rtl.Reg, succ rtl.Node) rtl.Node {
	entry := succ
	for i := len(exprs) - 1; i >= 0; i-- {
		entry = t.expr.TranslateExpr(exprs[i], regs[i], entry)
	}
	return entry
}

// Helper: chain expression evaluation
func (t *StmtTranslator) translateExprListChain(exprs []cminorsel.Expr, regs []rtl.Reg, succ rtl.Node) rtl.Node {
	return t.translateExprList(exprs, regs, succ)
}

// TranslateFunction translates a CminorSel function to RTL.
func TranslateFunction(fn cminorsel.Function) *rtl.Function {
	cfg := NewCFGBuilder()
	regs := NewRegAllocator()
	
	// Reset let binding state
	ResetLetBindings()
	
	// Map parameters to registers
	paramRegs := regs.MapParams(fn.Params)
	
	// Map local variables
	regs.MapVars(fn.Vars)
	
	// Set stack size
	cfg.SetStackSize(fn.Stackspace)
	
	// Create translator
	trans := NewStmtTranslator(cfg, regs)
	
	// Create return node (exit point)
	// Note: Sreturn creates its own return instruction
	// We use a dummy exit node for statements that fall through
	exitNode := cfg.AllocNode()
	cfg.AddInstr(exitNode, rtl.Ireturn{Arg: nil})
	
	// Translate body
	entryNode := trans.TranslateStmt(fn.Body, exitNode)
	
	// Build RTL function
	sig := rtl.Sig{
		Args:   fn.Sig.Args,
		Return: fn.Sig.Return,
		VarArg: fn.Sig.VarArg,
	}
	
	return &rtl.Function{
		Name:       fn.Name,
		Sig:        sig,
		Params:     paramRegs,
		Stacksize:  fn.Stackspace,
		Code:       cfg.GetCode(),
		Entrypoint: entryNode,
	}
}

// TranslateProgram translates a CminorSel program to RTL.
func TranslateProgram(prog cminorsel.Program) *rtl.Program {
	result := &rtl.Program{
		Globals:   make([]rtl.GlobVar, len(prog.Globals)),
		Functions: make([]rtl.Function, len(prog.Functions)),
	}
	
	// Copy globals
	for i, g := range prog.Globals {
		result.Globals[i] = rtl.GlobVar{
			Name: g.Name,
			Size: g.Size,
			Init: g.Init,
		}
	}
	
	// Translate functions
	for i, fn := range prog.Functions {
		translated := TranslateFunction(fn)
		result.Functions[i] = *translated
	}
	
	return result
}
