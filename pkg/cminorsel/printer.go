// Package cminorsel - Printer for CminorSel AST.
// This provides debugging output for the CminorSel intermediate representation.
// Note: CminorSel is an internal phase and has no corresponding CompCert dump flag.
package cminorsel

import (
	"fmt"
	"io"
	"strings"
)

// Printer outputs CminorSel AST in a readable format.
type Printer struct {
	w      io.Writer
	indent int
}

// NewPrinter creates a new printer writing to w.
func NewPrinter(w io.Writer) *Printer {
	return &Printer{w: w, indent: 0}
}

// Print outputs a complete program.
func (p *Printer) Print(prog Program) {
	fmt.Fprintf(p.w, "/* CminorSel program */\n\n")

	// Print globals
	for _, g := range prog.Globals {
		p.printGlobal(g)
	}
	if len(prog.Globals) > 0 {
		fmt.Fprintln(p.w)
	}

	// Print functions
	for _, f := range prog.Functions {
		p.printFunction(f)
		fmt.Fprintln(p.w)
	}
}

func (p *Printer) printGlobal(g GlobVar) {
	if g.Init != nil {
		fmt.Fprintf(p.w, "var %s[%d] = {...};\n", g.Name, g.Size)
	} else {
		fmt.Fprintf(p.w, "var %s[%d];\n", g.Name, g.Size)
	}
}

func (p *Printer) printFunction(f Function) {
	// Print signature
	fmt.Fprintf(p.w, "%s %s(", f.Sig.Return, f.Name)
	for i, param := range f.Params {
		if i > 0 {
			fmt.Fprint(p.w, ", ")
		}
		if i < len(f.Sig.Args) {
			fmt.Fprintf(p.w, "%s %s", f.Sig.Args[i], param)
		} else {
			fmt.Fprintf(p.w, "? %s", param)
		}
	}
	fmt.Fprintln(p.w, ") {")

	p.indent++

	// Print locals
	if len(f.Vars) > 0 {
		p.writeIndent()
		fmt.Fprint(p.w, "var ")
		fmt.Fprint(p.w, strings.Join(f.Vars, ", "))
		fmt.Fprintln(p.w, ";")
	}
	if f.Stackspace > 0 {
		p.writeIndent()
		fmt.Fprintf(p.w, "stackspace %d;\n", f.Stackspace)
	}

	// Print body
	p.printStmt(f.Body)

	p.indent--
	fmt.Fprintln(p.w, "}")
}

func (p *Printer) writeIndent() {
	for i := 0; i < p.indent; i++ {
		fmt.Fprint(p.w, "  ")
	}
}

func (p *Printer) printStmt(s Stmt) {
	switch stmt := s.(type) {
	case Sskip:
		p.writeIndent()
		fmt.Fprintln(p.w, "/* skip */;")

	case Sassign:
		p.writeIndent()
		fmt.Fprintf(p.w, "%s = ", stmt.Name)
		p.printExpr(stmt.RHS)
		fmt.Fprintln(p.w, ";")

	case Sstore:
		p.writeIndent()
		fmt.Fprintf(p.w, "%s[", stmt.Chunk)
		p.printAddressMode(stmt.Mode, stmt.Args)
		fmt.Fprint(p.w, "] = ")
		p.printExpr(stmt.Value)
		fmt.Fprintln(p.w, ";")

	case Scall:
		p.writeIndent()
		if stmt.Result != nil {
			fmt.Fprintf(p.w, "%s = ", *stmt.Result)
		}
		p.printExpr(stmt.Func)
		fmt.Fprint(p.w, "(")
		for i, arg := range stmt.Args {
			if i > 0 {
				fmt.Fprint(p.w, ", ")
			}
			p.printExpr(arg)
		}
		fmt.Fprintln(p.w, ");")

	case Stailcall:
		p.writeIndent()
		fmt.Fprint(p.w, "tailcall ")
		p.printExpr(stmt.Func)
		fmt.Fprint(p.w, "(")
		for i, arg := range stmt.Args {
			if i > 0 {
				fmt.Fprint(p.w, ", ")
			}
			p.printExpr(arg)
		}
		fmt.Fprintln(p.w, ");")

	case Sbuiltin:
		p.writeIndent()
		if stmt.Result != nil {
			fmt.Fprintf(p.w, "%s = ", *stmt.Result)
		}
		fmt.Fprintf(p.w, "%s(", stmt.Builtin)
		for i, arg := range stmt.Args {
			if i > 0 {
				fmt.Fprint(p.w, ", ")
			}
			p.printExpr(arg)
		}
		fmt.Fprintln(p.w, ");")

	case Sseq:
		p.printStmt(stmt.First)
		p.printStmt(stmt.Second)

	case Sifthenelse:
		p.writeIndent()
		fmt.Fprint(p.w, "if (")
		p.printCond(stmt.Cond)
		fmt.Fprintln(p.w, ") {")
		p.indent++
		p.printStmt(stmt.Then)
		p.indent--
		p.writeIndent()
		fmt.Fprintln(p.w, "} else {")
		p.indent++
		p.printStmt(stmt.Else)
		p.indent--
		p.writeIndent()
		fmt.Fprintln(p.w, "}")

	case Sloop:
		p.writeIndent()
		fmt.Fprintln(p.w, "loop {")
		p.indent++
		p.printStmt(stmt.Body)
		p.indent--
		p.writeIndent()
		fmt.Fprintln(p.w, "}")

	case Sblock:
		p.writeIndent()
		fmt.Fprintln(p.w, "{{ /* block */")
		p.indent++
		p.printStmt(stmt.Body)
		p.indent--
		p.writeIndent()
		fmt.Fprintln(p.w, "}}")

	case Sexit:
		p.writeIndent()
		fmt.Fprintf(p.w, "exit %d;\n", stmt.N)

	case Sswitch:
		p.writeIndent()
		if stmt.IsLong {
			fmt.Fprint(p.w, "switchl (")
		} else {
			fmt.Fprint(p.w, "switch (")
		}
		p.printExpr(stmt.Expr)
		fmt.Fprintln(p.w, ") {")
		for _, c := range stmt.Cases {
			p.writeIndent()
			fmt.Fprintf(p.w, "case %d:\n", c.Value)
			p.indent++
			p.printStmt(c.Body)
			p.indent--
		}
		p.writeIndent()
		fmt.Fprintln(p.w, "default:")
		p.indent++
		p.printStmt(stmt.Default)
		p.indent--
		p.writeIndent()
		fmt.Fprintln(p.w, "}")

	case Sreturn:
		p.writeIndent()
		if stmt.Value == nil {
			fmt.Fprintln(p.w, "return;")
		} else {
			fmt.Fprint(p.w, "return ")
			p.printExpr(stmt.Value)
			fmt.Fprintln(p.w, ";")
		}

	case Slabel:
		// Labels are unindented
		fmt.Fprintf(p.w, "%s:\n", stmt.Label)
		p.printStmt(stmt.Body)

	case Sgoto:
		p.writeIndent()
		fmt.Fprintf(p.w, "goto %s;\n", stmt.Label)
	}
}

func (p *Printer) printExpr(e Expr) {
	switch expr := e.(type) {
	case Evar:
		fmt.Fprint(p.w, expr.Name)

	case Econst:
		p.printConst(expr.Const)

	case Eunop:
		fmt.Fprintf(p.w, "%s(", expr.Op)
		p.printExpr(expr.Arg)
		fmt.Fprint(p.w, ")")

	case Ebinop:
		fmt.Fprint(p.w, "(")
		p.printExpr(expr.Left)
		fmt.Fprintf(p.w, " %s ", expr.Op)
		p.printExpr(expr.Right)
		fmt.Fprint(p.w, ")")

	case Eload:
		fmt.Fprintf(p.w, "%s[", expr.Chunk)
		p.printAddressMode(expr.Mode, expr.Args)
		fmt.Fprint(p.w, "]")

	case Econdition:
		fmt.Fprint(p.w, "(")
		p.printCond(expr.Cond)
		fmt.Fprint(p.w, " ? ")
		p.printExpr(expr.Then)
		fmt.Fprint(p.w, " : ")
		p.printExpr(expr.Else)
		fmt.Fprint(p.w, ")")

	case Elet:
		fmt.Fprint(p.w, "(let ")
		p.printExpr(expr.Bind)
		fmt.Fprint(p.w, " in ")
		p.printExpr(expr.Body)
		fmt.Fprint(p.w, ")")

	case Eletvar:
		fmt.Fprintf(p.w, "#%d", expr.Index)

	case Eaddshift:
		fmt.Fprint(p.w, "(")
		p.printExpr(expr.Left)
		fmt.Fprintf(p.w, " + (%s ", expr.Op)
		p.printExpr(expr.Right)
		fmt.Fprintf(p.w, " %d)", expr.Shift)
		fmt.Fprint(p.w, ")")

	case Esubshift:
		fmt.Fprint(p.w, "(")
		p.printExpr(expr.Left)
		fmt.Fprintf(p.w, " - (%s ", expr.Op)
		p.printExpr(expr.Right)
		fmt.Fprintf(p.w, " %d)", expr.Shift)
		fmt.Fprint(p.w, ")")
	}
}

func (p *Printer) printConst(c Constant) {
	switch cnst := c.(type) {
	case Ointconst:
		fmt.Fprintf(p.w, "%d", cnst.Value)
	case Olongconst:
		fmt.Fprintf(p.w, "%dL", cnst.Value)
	case Ofloatconst:
		fmt.Fprintf(p.w, "%g", cnst.Value)
	case Osingleconst:
		fmt.Fprintf(p.w, "%gf", cnst.Value)
	case Oaddrsymbol:
		if cnst.Offset == 0 {
			fmt.Fprintf(p.w, "&%s", cnst.Symbol)
		} else {
			fmt.Fprintf(p.w, "&%s+%d", cnst.Symbol, cnst.Offset)
		}
	case Oaddrstack:
		fmt.Fprintf(p.w, "[sp+%d]", cnst.Offset)
	}
}

func (p *Printer) printAddressMode(mode AddressingMode, args []Expr) {
	switch m := mode.(type) {
	case Aindexed:
		if len(args) > 0 {
			p.printExpr(args[0])
		}
		if m.Offset != 0 {
			fmt.Fprintf(p.w, "+%d", m.Offset)
		}

	case Aindexed2:
		if len(args) >= 2 {
			p.printExpr(args[0])
			fmt.Fprint(p.w, "+")
			p.printExpr(args[1])
		}

	case Aindexed2shift:
		if len(args) >= 2 {
			p.printExpr(args[0])
			fmt.Fprint(p.w, "+")
			p.printExpr(args[1])
			fmt.Fprintf(p.w, "<<%d", m.Shift)
		}

	case Aindexed2ext:
		if len(args) >= 2 {
			p.printExpr(args[0])
			fmt.Fprintf(p.w, "+%s(", m.Extend)
			p.printExpr(args[1])
			fmt.Fprintf(p.w, ")<<%d", m.Shift)
		}

	case Aglobal:
		fmt.Fprintf(p.w, "&%s", m.Symbol)
		if m.Offset != 0 {
			fmt.Fprintf(p.w, "+%d", m.Offset)
		}

	case Ainstack:
		fmt.Fprintf(p.w, "[sp+%d]", m.Offset)
	}
}

func (p *Printer) printCond(c Condition) {
	switch cond := c.(type) {
	case CondTrue:
		fmt.Fprint(p.w, "true")

	case CondFalse:
		fmt.Fprint(p.w, "false")

	case CondCmp:
		p.printExpr(cond.Left)
		fmt.Fprintf(p.w, " %s ", cond.Cmp)
		p.printExpr(cond.Right)

	case CondCmpu:
		p.printExpr(cond.Left)
		fmt.Fprintf(p.w, " %su ", cond.Cmp)
		p.printExpr(cond.Right)

	case CondCmpf:
		p.printExpr(cond.Left)
		fmt.Fprintf(p.w, " %sf ", cond.Cmp)
		p.printExpr(cond.Right)

	case CondCmps:
		p.printExpr(cond.Left)
		fmt.Fprintf(p.w, " %ss ", cond.Cmp)
		p.printExpr(cond.Right)

	case CondCmpl:
		p.printExpr(cond.Left)
		fmt.Fprintf(p.w, " %sl ", cond.Cmp)
		p.printExpr(cond.Right)

	case CondCmplu:
		p.printExpr(cond.Left)
		fmt.Fprintf(p.w, " %slu ", cond.Cmp)
		p.printExpr(cond.Right)

	case CondNot:
		fmt.Fprint(p.w, "!(")
		p.printCond(cond.Cond)
		fmt.Fprint(p.w, ")")

	case CondAnd:
		fmt.Fprint(p.w, "(")
		p.printCond(cond.Left)
		fmt.Fprint(p.w, " && ")
		p.printCond(cond.Right)
		fmt.Fprint(p.w, ")")

	case CondOr:
		fmt.Fprint(p.w, "(")
		p.printCond(cond.Left)
		fmt.Fprint(p.w, " || ")
		p.printCond(cond.Right)
		fmt.Fprint(p.w, ")")
	}
}

// PrintExpr outputs a single expression (for debugging).
func (p *Printer) PrintExpr(e Expr) {
	p.printExpr(e)
}

// PrintStmt outputs a single statement (for debugging).
func (p *Printer) PrintStmt(s Stmt) {
	p.printStmt(s)
}
