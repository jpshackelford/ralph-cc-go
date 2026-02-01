// Package ltl provides AST printing functionality for LTL IR
// Format matches CompCert's .ltl output
package ltl

import (
	"fmt"
	"io"
	"sort"

	"github.com/raymyers/ralph-cc/pkg/rtl"
)

// Printer outputs the LTL AST in CompCert-compatible format
type Printer struct {
	w io.Writer
}

// NewPrinter creates a new LTL AST printer
func NewPrinter(w io.Writer) *Printer {
	return &Printer{w: w}
}

// PrintProgram prints a complete LTL program
func (p *Printer) PrintProgram(prog *Program) {
	// Print global variables
	for _, g := range prog.Globals {
		fmt.Fprintf(p.w, "var \"%s\"[%d]\n", g.Name, g.Size)
	}
	if len(prog.Globals) > 0 {
		fmt.Fprintln(p.w)
	}

	// Print functions
	for i, fn := range prog.Functions {
		p.PrintFunction(&fn)
		if i < len(prog.Functions)-1 {
			fmt.Fprintln(p.w)
		}
	}
}

// PrintFunction prints a function in LTL format
func (p *Printer) PrintFunction(fn *Function) {
	// Function header: name(params)
	fmt.Fprintf(p.w, "%s(", fn.Name)
	for i, loc := range fn.Params {
		if i > 0 {
			fmt.Fprint(p.w, ", ")
		}
		p.printLoc(loc)
	}
	fmt.Fprintln(p.w, ") {")

	// Sort nodes for deterministic output
	nodes := make([]Node, 0, len(fn.Code))
	for n := range fn.Code {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i] < nodes[j]
	})

	// Print each basic block
	for _, n := range nodes {
		block := fn.Code[n]
		fmt.Fprintf(p.w, "  %d: { ", n)
		for i, instr := range block.Body {
			if i > 0 {
				fmt.Fprint(p.w, "; ")
			}
			p.printInstruction(instr)
		}
		fmt.Fprintln(p.w, " }")
	}

	fmt.Fprintln(p.w, "}")
	fmt.Fprintf(p.w, "entry: %d\n", fn.Entrypoint)
}

func (p *Printer) printInstruction(instr Instruction) {
	switch i := instr.(type) {
	case Lnop:
		fmt.Fprint(p.w, "Lnop")
	case Lop:
		fmt.Fprint(p.w, "Lop(")
		p.printOperation(i.Op)
		fmt.Fprint(p.w, ", [")
		for j, arg := range i.Args {
			if j > 0 {
				fmt.Fprint(p.w, "; ")
			}
			p.printLoc(arg)
		}
		fmt.Fprint(p.w, "], ")
		p.printLoc(i.Dest)
		fmt.Fprint(p.w, ")")
	case Lload:
		fmt.Fprintf(p.w, "Lload(%s, ", chunkName(i.Chunk))
		p.printAddressingMode(i.Addr)
		fmt.Fprint(p.w, ", [")
		for j, arg := range i.Args {
			if j > 0 {
				fmt.Fprint(p.w, "; ")
			}
			p.printLoc(arg)
		}
		fmt.Fprint(p.w, "], ")
		p.printLoc(i.Dest)
		fmt.Fprint(p.w, ")")
	case Lstore:
		fmt.Fprintf(p.w, "Lstore(%s, ", chunkName(i.Chunk))
		p.printAddressingMode(i.Addr)
		fmt.Fprint(p.w, ", [")
		for j, arg := range i.Args {
			if j > 0 {
				fmt.Fprint(p.w, "; ")
			}
			p.printLoc(arg)
		}
		fmt.Fprint(p.w, "], ")
		p.printLoc(i.Src)
		fmt.Fprint(p.w, ")")
	case Lcall:
		fmt.Fprint(p.w, "Lcall(")
		p.printFunRef(i.Fn)
		fmt.Fprint(p.w, ", [")
		for j, arg := range i.Args {
			if j > 0 {
				fmt.Fprint(p.w, "; ")
			}
			p.printLoc(arg)
		}
		fmt.Fprint(p.w, "])")
	case Ltailcall:
		fmt.Fprint(p.w, "Ltailcall(")
		p.printFunRef(i.Fn)
		fmt.Fprint(p.w, ", [")
		for j, arg := range i.Args {
			if j > 0 {
				fmt.Fprint(p.w, "; ")
			}
			p.printLoc(arg)
		}
		fmt.Fprint(p.w, "])")
	case Lbuiltin:
		fmt.Fprintf(p.w, "Lbuiltin(%q, [", i.Builtin)
		for j, arg := range i.Args {
			if j > 0 {
				fmt.Fprint(p.w, "; ")
			}
			p.printLoc(arg)
		}
		fmt.Fprint(p.w, "]")
		if i.Dest != nil {
			fmt.Fprint(p.w, ", ")
			p.printLoc(*i.Dest)
		}
		fmt.Fprint(p.w, ")")
	case Lbranch:
		fmt.Fprintf(p.w, "Lbranch %d", i.Succ)
	case Lcond:
		fmt.Fprint(p.w, "Lcond(")
		p.printConditionCode(i.Cond, i.Args)
		fmt.Fprintf(p.w, ", %d, %d)", i.IfSo, i.IfNot)
	case Ljumptable:
		fmt.Fprint(p.w, "Ljumptable(")
		p.printLoc(i.Arg)
		fmt.Fprint(p.w, ", [")
		for j, t := range i.Targets {
			if j > 0 {
				fmt.Fprint(p.w, "; ")
			}
			fmt.Fprintf(p.w, "%d", t)
		}
		fmt.Fprint(p.w, "])")
	case Lreturn:
		fmt.Fprint(p.w, "Lreturn")
	default:
		fmt.Fprint(p.w, "???")
	}
}

func (p *Printer) printLoc(loc Loc) {
	switch l := loc.(type) {
	case R:
		fmt.Fprint(p.w, l.Reg.String())
	case S:
		fmt.Fprintf(p.w, "S(%s, %d, %s)", l.Slot, l.Ofs, l.Ty)
	default:
		fmt.Fprint(p.w, "?loc")
	}
}

func (p *Printer) printOperation(op Operation) {
	switch o := op.(type) {
	case rtl.Omove:
		fmt.Fprint(p.w, "Omove")
	case rtl.Ointconst:
		fmt.Fprintf(p.w, "Ointconst(%d)", o.Value)
	case rtl.Olongconst:
		fmt.Fprintf(p.w, "Olongconst(%dL)", o.Value)
	case rtl.Ofloatconst:
		fmt.Fprintf(p.w, "Ofloatconst(%v)", o.Value)
	case rtl.Osingleconst:
		fmt.Fprintf(p.w, "Osingleconst(%vf)", o.Value)
	case rtl.Oaddrsymbol:
		fmt.Fprintf(p.w, "Oaddrsymbol(\"%s\", %d)", o.Symbol, o.Offset)
	case rtl.Oaddrstack:
		fmt.Fprintf(p.w, "Oaddrstack(%d)", o.Offset)
	case rtl.Oadd:
		fmt.Fprint(p.w, "Oadd")
	case rtl.Oaddimm:
		fmt.Fprintf(p.w, "Oaddimm(%d)", o.N)
	case rtl.Oneg:
		fmt.Fprint(p.w, "Oneg")
	case rtl.Osub:
		fmt.Fprint(p.w, "Osub")
	case rtl.Omul:
		fmt.Fprint(p.w, "Omul")
	case rtl.Omulimm:
		fmt.Fprintf(p.w, "Omulimm(%d)", o.N)
	case rtl.Odiv:
		fmt.Fprint(p.w, "Odiv")
	case rtl.Odivu:
		fmt.Fprint(p.w, "Odivu")
	case rtl.Omod:
		fmt.Fprint(p.w, "Omod")
	case rtl.Omodu:
		fmt.Fprint(p.w, "Omodu")
	case rtl.Oand:
		fmt.Fprint(p.w, "Oand")
	case rtl.Oandimm:
		fmt.Fprintf(p.w, "Oandimm(%d)", o.N)
	case rtl.Oor:
		fmt.Fprint(p.w, "Oor")
	case rtl.Oorimm:
		fmt.Fprintf(p.w, "Oorimm(%d)", o.N)
	case rtl.Oxor:
		fmt.Fprint(p.w, "Oxor")
	case rtl.Oxorimm:
		fmt.Fprintf(p.w, "Oxorimm(%d)", o.N)
	case rtl.Onot:
		fmt.Fprint(p.w, "Onot")
	case rtl.Oshl:
		fmt.Fprint(p.w, "Oshl")
	case rtl.Oshlimm:
		fmt.Fprintf(p.w, "Oshlimm(%d)", o.N)
	case rtl.Oshr:
		fmt.Fprint(p.w, "Oshr")
	case rtl.Oshrimm:
		fmt.Fprintf(p.w, "Oshrimm(%d)", o.N)
	case rtl.Oshru:
		fmt.Fprint(p.w, "Oshru")
	case rtl.Oshruimm:
		fmt.Fprintf(p.w, "Oshruimm(%d)", o.N)
	case rtl.Oaddl:
		fmt.Fprint(p.w, "Oaddl")
	case rtl.Oaddlimm:
		fmt.Fprintf(p.w, "Oaddlimm(%dL)", o.N)
	case rtl.Onegl:
		fmt.Fprint(p.w, "Onegl")
	case rtl.Osubl:
		fmt.Fprint(p.w, "Osubl")
	case rtl.Omull:
		fmt.Fprint(p.w, "Omull")
	case rtl.Odivl:
		fmt.Fprint(p.w, "Odivl")
	case rtl.Odivlu:
		fmt.Fprint(p.w, "Odivlu")
	case rtl.Omodl:
		fmt.Fprint(p.w, "Omodl")
	case rtl.Omodlu:
		fmt.Fprint(p.w, "Omodlu")
	case rtl.Oandl:
		fmt.Fprint(p.w, "Oandl")
	case rtl.Oorl:
		fmt.Fprint(p.w, "Oorl")
	case rtl.Oxorl:
		fmt.Fprint(p.w, "Oxorl")
	case rtl.Onotl:
		fmt.Fprint(p.w, "Onotl")
	case rtl.Oshll:
		fmt.Fprint(p.w, "Oshll")
	case rtl.Oshllimm:
		fmt.Fprintf(p.w, "Oshllimm(%d)", o.N)
	case rtl.Oshrl:
		fmt.Fprint(p.w, "Oshrl")
	case rtl.Oshrlimm:
		fmt.Fprintf(p.w, "Oshrlimm(%d)", o.N)
	case rtl.Oshrlu:
		fmt.Fprint(p.w, "Oshrlu")
	case rtl.Oshrluimm:
		fmt.Fprintf(p.w, "Oshrluimm(%d)", o.N)
	case rtl.Ocast8signed:
		fmt.Fprint(p.w, "Ocast8signed")
	case rtl.Ocast8unsigned:
		fmt.Fprint(p.w, "Ocast8unsigned")
	case rtl.Ocast16signed:
		fmt.Fprint(p.w, "Ocast16signed")
	case rtl.Ocast16unsigned:
		fmt.Fprint(p.w, "Ocast16unsigned")
	case rtl.Olongofint:
		fmt.Fprint(p.w, "Olongofint")
	case rtl.Olongofintu:
		fmt.Fprint(p.w, "Olongofintu")
	case rtl.Ointoflong:
		fmt.Fprint(p.w, "Ointoflong")
	case rtl.Onegf:
		fmt.Fprint(p.w, "Onegf")
	case rtl.Oabsf:
		fmt.Fprint(p.w, "Oabsf")
	case rtl.Oaddf:
		fmt.Fprint(p.w, "Oaddf")
	case rtl.Osubf:
		fmt.Fprint(p.w, "Osubf")
	case rtl.Omulf:
		fmt.Fprint(p.w, "Omulf")
	case rtl.Odivf:
		fmt.Fprint(p.w, "Odivf")
	case rtl.Onegs:
		fmt.Fprint(p.w, "Onegs")
	case rtl.Oabss:
		fmt.Fprint(p.w, "Oabss")
	case rtl.Oadds:
		fmt.Fprint(p.w, "Oadds")
	case rtl.Osubs:
		fmt.Fprint(p.w, "Osubs")
	case rtl.Omuls:
		fmt.Fprint(p.w, "Omuls")
	case rtl.Odivs:
		fmt.Fprint(p.w, "Odivs")
	case rtl.Osingleoffloat:
		fmt.Fprint(p.w, "Osingleoffloat")
	case rtl.Ofloatofsingle:
		fmt.Fprint(p.w, "Ofloatofsingle")
	case rtl.Ointoffloat:
		fmt.Fprint(p.w, "Ointoffloat")
	case rtl.Ointuoffloat:
		fmt.Fprint(p.w, "Ointuoffloat")
	case rtl.Ofloatofint:
		fmt.Fprint(p.w, "Ofloatofint")
	case rtl.Ofloatofintu:
		fmt.Fprint(p.w, "Ofloatofintu")
	case rtl.Olongoffloat:
		fmt.Fprint(p.w, "Olongoffloat")
	case rtl.Olonguoffloat:
		fmt.Fprint(p.w, "Olonguoffloat")
	case rtl.Ofloatoflong:
		fmt.Fprint(p.w, "Ofloatoflong")
	case rtl.Ofloatoflongu:
		fmt.Fprint(p.w, "Ofloatoflongu")
	case rtl.Ocmp:
		fmt.Fprintf(p.w, "Ocmp(%s)", o.Cond)
	case rtl.Ocmpu:
		fmt.Fprintf(p.w, "Ocmpu(%s)", o.Cond)
	case rtl.Ocmpimm:
		fmt.Fprintf(p.w, "Ocmpimm(%s, %d)", o.Cond, o.N)
	case rtl.Ocmpuimm:
		fmt.Fprintf(p.w, "Ocmpuimm(%s, %d)", o.Cond, o.N)
	case rtl.Ocmpl:
		fmt.Fprintf(p.w, "Ocmpl(%s)", o.Cond)
	case rtl.Ocmplu:
		fmt.Fprintf(p.w, "Ocmplu(%s)", o.Cond)
	case rtl.Ocmplimm:
		fmt.Fprintf(p.w, "Ocmplimm(%s, %dL)", o.Cond, o.N)
	case rtl.Ocmpluimm:
		fmt.Fprintf(p.w, "Ocmpluimm(%s, %dL)", o.Cond, o.N)
	case rtl.Ocmpf:
		fmt.Fprintf(p.w, "Ocmpf(%s)", o.Cond)
	case rtl.Ocmps:
		fmt.Fprintf(p.w, "Ocmps(%s)", o.Cond)
	default:
		fmt.Fprintf(p.w, "op?(%T)", op)
	}
}

func (p *Printer) printAddressingMode(addr AddressingMode) {
	switch a := addr.(type) {
	case Aindexed:
		fmt.Fprintf(p.w, "Aindexed(%d)", a.Offset)
	case Aindexed2:
		fmt.Fprint(p.w, "Aindexed2")
	case Aindexed2shift:
		fmt.Fprintf(p.w, "Aindexed2shift(%d)", a.Shift)
	case Aglobal:
		fmt.Fprintf(p.w, "Aglobal(\"%s\", %d)", a.Symbol, a.Offset)
	case Ainstack:
		fmt.Fprintf(p.w, "Ainstack(%d)", a.Offset)
	default:
		fmt.Fprint(p.w, "addr?")
	}
}

func (p *Printer) printFunRef(fn FunRef) {
	switch f := fn.(type) {
	case FunSymbol:
		fmt.Fprintf(p.w, "\"%s\"", f.Name)
	case FunReg:
		p.printLoc(f.Loc)
	}
}

func (p *Printer) printConditionCode(cc ConditionCode, args []Loc) {
	switch c := cc.(type) {
	case rtl.Ccomp:
		fmt.Fprintf(p.w, "Ccomp(%s)", c.Cond)
	case rtl.Ccompu:
		fmt.Fprintf(p.w, "Ccompu(%s)", c.Cond)
	case rtl.Ccompimm:
		fmt.Fprintf(p.w, "Ccompimm(%s, %d)", c.Cond, c.N)
	case rtl.Ccompuimm:
		fmt.Fprintf(p.w, "Ccompuimm(%s, %d)", c.Cond, c.N)
	case rtl.Ccompl:
		fmt.Fprintf(p.w, "Ccompl(%s)", c.Cond)
	case rtl.Ccomplu:
		fmt.Fprintf(p.w, "Ccomplu(%s)", c.Cond)
	case rtl.Ccomplimm:
		fmt.Fprintf(p.w, "Ccomplimm(%s, %dL)", c.Cond, c.N)
	case rtl.Ccompluimm:
		fmt.Fprintf(p.w, "Ccompluimm(%s, %dL)", c.Cond, c.N)
	case rtl.Ccompf:
		fmt.Fprintf(p.w, "Ccompf(%s)", c.Cond)
	case rtl.Cnotcompf:
		fmt.Fprintf(p.w, "Cnotcompf(%s)", c.Cond)
	case rtl.Ccomps:
		fmt.Fprintf(p.w, "Ccomps(%s)", c.Cond)
	case rtl.Cnotcomps:
		fmt.Fprintf(p.w, "Cnotcomps(%s)", c.Cond)
	default:
		fmt.Fprint(p.w, "cond?")
	}
}

func chunkName(c Chunk) string {
	switch c {
	case Mint8signed:
		return "Mint8signed"
	case Mint8unsigned:
		return "Mint8unsigned"
	case Mint16signed:
		return "Mint16signed"
	case Mint16unsigned:
		return "Mint16unsigned"
	case Mint32:
		return "Mint32"
	case Mint64:
		return "Mint64"
	case Mfloat32:
		return "Mfloat32"
	case Mfloat64:
		return "Mfloat64"
	default:
		return "mem?"
	}
}
