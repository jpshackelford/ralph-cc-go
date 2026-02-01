package mach

import (
	"fmt"
	"io"

	"github.com/raymyers/ralph-cc/pkg/ltl"
	"github.com/raymyers/ralph-cc/pkg/rtl"
)

// Printer outputs Mach code in a format similar to CompCert's -dmach output
type Printer struct {
	w io.Writer
}

// NewPrinter creates a new Mach printer
func NewPrinter(w io.Writer) *Printer {
	return &Printer{w: w}
}

// PrintProgram prints an entire Mach program
func (p *Printer) PrintProgram(prog *Program) {
	// Print globals
	for _, g := range prog.Globals {
		p.printGlobal(g)
	}

	// Print functions
	for _, fn := range prog.Functions {
		p.PrintFunction(&fn)
	}
}

// printGlobal prints a global variable
func (p *Printer) printGlobal(g GlobVar) {
	fmt.Fprintf(p.w, "  \"%s\" [%d]\n", g.Name, g.Size)
}

// PrintFunction prints a single Mach function
func (p *Printer) PrintFunction(fn *Function) {
	// Function header
	fmt.Fprintf(p.w, "%s:\n", fn.Name)

	// Print stack frame info
	fmt.Fprintf(p.w, "  ; stack frame: %d bytes\n", fn.Stacksize)
	if len(fn.CalleeSaveRegs) > 0 {
		fmt.Fprintf(p.w, "  ; callee-save: ")
		for i, reg := range fn.CalleeSaveRegs {
			if i > 0 {
				fmt.Fprintf(p.w, ", ")
			}
			fmt.Fprintf(p.w, "%s", reg.String())
		}
		fmt.Fprintln(p.w)
	}

	// Print code
	for _, inst := range fn.Code {
		p.printInstruction(inst)
	}

	fmt.Fprintln(p.w) // blank line after function
}

// printInstruction prints a single Mach instruction
func (p *Printer) printInstruction(inst Instruction) {
	switch i := inst.(type) {
	case Mgetstack:
		fmt.Fprintf(p.w, "  %s = stack(%d) : %s\n", i.Dest.String(), i.Ofs, typString(i.Ty))

	case Msetstack:
		fmt.Fprintf(p.w, "  stack(%d) = %s : %s\n", i.Ofs, i.Src.String(), typString(i.Ty))

	case Mgetparam:
		fmt.Fprintf(p.w, "  %s = param(%d) : %s\n", i.Dest.String(), i.Ofs, typString(i.Ty))

	case Mop:
		p.printOp(i)

	case Mload:
		fmt.Fprintf(p.w, "  %s = %s[%s]\n", i.Dest.String(), chunkString(i.Chunk), p.addrString(i.Addr, i.Args))

	case Mstore:
		fmt.Fprintf(p.w, "  %s[%s] = %s\n", chunkString(i.Chunk), p.addrString(i.Addr, i.Args), i.Src.String())

	case Mcall:
		fmt.Fprintf(p.w, "  call %s\n", p.funRefString(i.Fn))

	case Mtailcall:
		fmt.Fprintf(p.w, "  tailcall %s\n", p.funRefString(i.Fn))

	case Mbuiltin:
		if i.Dest != nil {
			fmt.Fprintf(p.w, "  %s = builtin %s(%s)\n", i.Dest.String(), i.Builtin, p.regsString(i.Args))
		} else {
			fmt.Fprintf(p.w, "  builtin %s(%s)\n", i.Builtin, p.regsString(i.Args))
		}

	case Mlabel:
		fmt.Fprintf(p.w, "%d:\n", i.Lbl)

	case Mgoto:
		fmt.Fprintf(p.w, "  goto %d\n", i.Target)

	case Mcond:
		fmt.Fprintf(p.w, "  if %s(%s) goto %d\n", p.condString(i.Cond), p.regsString(i.Args), i.IfSo)

	case Mjumptable:
		fmt.Fprintf(p.w, "  jumptable %s [", i.Arg.String())
		for j, lbl := range i.Targets {
			if j > 0 {
				fmt.Fprintf(p.w, ", ")
			}
			fmt.Fprintf(p.w, "%d", lbl)
		}
		fmt.Fprintln(p.w, "]")

	case Mreturn:
		fmt.Fprintln(p.w, "  return")

	default:
		fmt.Fprintf(p.w, "  ; unknown instruction %T\n", inst)
	}
}

// printOp prints an Mop instruction
func (p *Printer) printOp(op Mop) {
	opName := operationName(op.Op)

	if len(op.Args) == 0 {
		// No source args (e.g., constant load or SP manipulation)
		fmt.Fprintf(p.w, "  %s = %s\n", op.Dest.String(), opName)
	} else if len(op.Args) == 1 {
		// Unary operation
		fmt.Fprintf(p.w, "  %s = %s %s\n", op.Dest.String(), opName, op.Args[0].String())
	} else {
		// Binary operation
		fmt.Fprintf(p.w, "  %s = %s %s, %s\n", op.Dest.String(), opName, op.Args[0].String(), op.Args[1].String())
	}
}

// operationName returns a string representation of an operation
func operationName(op Operation) string {
	switch o := op.(type) {
	case rtl.Omove:
		return "move"
	case rtl.Oadd:
		return "add"
	case rtl.Oaddimm:
		return fmt.Sprintf("addimm(%d)", o.N)
	case rtl.Oaddlimm:
		return fmt.Sprintf("addlimm(%d)", o.N)
	case rtl.Osub:
		return "sub"
	case rtl.Omul:
		return "mul"
	case rtl.Odiv:
		return "div"
	case rtl.Odivu:
		return "divu"
	case rtl.Omod:
		return "mod"
	case rtl.Omodu:
		return "modu"
	case rtl.Oneg:
		return "neg"
	case rtl.Oand:
		return "and"
	case rtl.Oandimm:
		return fmt.Sprintf("andimm(%d)", o.N)
	case rtl.Oor:
		return "or"
	case rtl.Oorimm:
		return fmt.Sprintf("orimm(%d)", o.N)
	case rtl.Oxor:
		return "xor"
	case rtl.Oxorimm:
		return fmt.Sprintf("xorimm(%d)", o.N)
	case rtl.Onot:
		return "not"
	case rtl.Oshl:
		return "shl"
	case rtl.Oshlimm:
		return fmt.Sprintf("shlimm(%d)", o.N)
	case rtl.Oshr:
		return "shr"
	case rtl.Oshrimm:
		return fmt.Sprintf("shrimm(%d)", o.N)
	case rtl.Oshru:
		return "shru"
	case rtl.Oshruimm:
		return fmt.Sprintf("shruimm(%d)", o.N)
	case rtl.Oaddl:
		return "addl"
	case rtl.Osubl:
		return "subl"
	case rtl.Omull:
		return "mull"
	case rtl.Ointconst:
		return fmt.Sprintf("intconst(%d)", o.Value)
	case rtl.Olongconst:
		return fmt.Sprintf("longconst(%d)", o.Value)
	case rtl.Ofloatconst:
		return fmt.Sprintf("floatconst(%g)", o.Value)
	case rtl.Osingleconst:
		return fmt.Sprintf("singleconst(%g)", o.Value)
	case rtl.Oaddrsymbol:
		return fmt.Sprintf("addrsymbol(%q, %d)", o.Symbol, o.Offset)
	case rtl.Oaddrstack:
		return fmt.Sprintf("addrstack(%d)", o.Offset)
	default:
		return fmt.Sprintf("%T", op)
	}
}

// typString returns a string for a Typ
func typString(ty Typ) string {
	switch ty {
	case Tint:
		return "int"
	case Tfloat:
		return "float"
	case Tlong:
		return "long"
	case Tsingle:
		return "single"
	case Tany32:
		return "any32"
	case Tany64:
		return "any64"
	default:
		return "?"
	}
}

// chunkString returns a string for a memory chunk
func chunkString(c Chunk) string {
	switch c {
	case Mint8signed:
		return "int8"
	case Mint8unsigned:
		return "uint8"
	case Mint16signed:
		return "int16"
	case Mint16unsigned:
		return "uint16"
	case Mint32:
		return "int32"
	case Mint64:
		return "int64"
	case Mfloat32:
		return "float32"
	case Mfloat64:
		return "float64"
	default:
		return "?"
	}
}

// addrString returns a string for an addressing mode
func (p *Printer) addrString(addr AddressingMode, args []MReg) string {
	switch a := addr.(type) {
	case ltl.Aindexed:
		if len(args) > 0 {
			return fmt.Sprintf("%s + %d", args[0].String(), a.Offset)
		}
		return fmt.Sprintf("?? + %d", a.Offset)
	case ltl.Aindexed2:
		if len(args) > 1 {
			return fmt.Sprintf("%s + %s", args[0].String(), args[1].String())
		}
		return "?? + ??"
	case ltl.Aglobal:
		return fmt.Sprintf("%q + %d", a.Symbol, a.Offset)
	case ltl.Ainstack:
		return fmt.Sprintf("sp + %d", a.Offset)
	case ltl.Aindexed2shift:
		if len(args) > 1 {
			return fmt.Sprintf("%s + %s << %d", args[0].String(), args[1].String(), a.Shift)
		}
		return fmt.Sprintf("?? + ?? << %d", a.Shift)
	default:
		return "??"
	}
}

// regsString returns a comma-separated list of registers
func (p *Printer) regsString(regs []MReg) string {
	if len(regs) == 0 {
		return ""
	}
	result := regs[0].String()
	for i := 1; i < len(regs); i++ {
		result += ", " + regs[i].String()
	}
	return result
}

// funRefString returns a string for a function reference
func (p *Printer) funRefString(fn FunRef) string {
	switch f := fn.(type) {
	case FunReg:
		return fmt.Sprintf("*%s", f.Reg.String())
	case FunSymbol:
		return f.Name
	default:
		return "?"
	}
}

// condString returns a string for a condition code
func (p *Printer) condString(cond ConditionCode) string {
	switch c := cond.(type) {
	case rtl.Ccomp:
		return fmt.Sprintf("cmp%s", conditionString(c.Cond))
	case rtl.Ccompu:
		return fmt.Sprintf("cmpu%s", conditionString(c.Cond))
	case rtl.Ccompimm:
		return fmt.Sprintf("cmpimm%s(%d)", conditionString(c.Cond), c.N)
	case rtl.Ccompuimm:
		return fmt.Sprintf("cmpuimm%s(%d)", conditionString(c.Cond), c.N)
	case rtl.Ccompl:
		return fmt.Sprintf("cmpl%s", conditionString(c.Cond))
	case rtl.Ccomplu:
		return fmt.Sprintf("cmplu%s", conditionString(c.Cond))
	case rtl.Ccomplimm:
		return fmt.Sprintf("cmplimm%s(%d)", conditionString(c.Cond), c.N)
	case rtl.Ccompluimm:
		return fmt.Sprintf("cmpluimm%s(%d)", conditionString(c.Cond), c.N)
	case rtl.Ccompf:
		return fmt.Sprintf("cmpf%s", conditionString(c.Cond))
	case rtl.Ccomps:
		return fmt.Sprintf("cmps%s", conditionString(c.Cond))
	default:
		return fmt.Sprintf("%T", cond)
	}
}

// conditionString returns a string for a comparison condition
func conditionString(cond rtl.Condition) string {
	switch cond {
	case rtl.Ceq:
		return "eq"
	case rtl.Cne:
		return "ne"
	case rtl.Clt:
		return "lt"
	case rtl.Cle:
		return "le"
	case rtl.Cgt:
		return "gt"
	case rtl.Cge:
		return "ge"
	default:
		return "?"
	}
}
