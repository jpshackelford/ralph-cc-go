package regalloc

import (
	"sort"

	"github.com/raymyers/ralph-cc/pkg/ltl"
	"github.com/raymyers/ralph-cc/pkg/rtl"
)

// TransformFunction transforms an RTL function to LTL by applying register allocation
func TransformFunction(rtlFn *rtl.Function) *ltl.Function {
	// Perform register allocation
	allocation := AllocateFunction(rtlFn)

	ltlFn := ltl.NewFunction(rtlFn.Name, rtlFn.Sig)
	ltlFn.Stacksize = rtlFn.Stacksize + allocation.StackSize
	ltlFn.Entrypoint = ltl.Node(rtlFn.Entrypoint)

	// Transform parameter locations
	for _, param := range rtlFn.Params {
		if loc, ok := allocation.RegToLoc[param]; ok {
			ltlFn.Params = append(ltlFn.Params, loc)
		}
	}

	// Group instructions into basic blocks
	// For simplicity, we create one block per RTL node initially
	// (CompCert's Linearize pass will optimize this later)
	sortedNodes := getSortedNodes(rtlFn)
	for _, node := range sortedNodes {
		instr := rtlFn.Code[node]
		ltlBlock := transformInstruction(instr, allocation)
		ltlFn.Code[ltl.Node(node)] = ltlBlock
	}

	return ltlFn
}

func getSortedNodes(fn *rtl.Function) []rtl.Node {
	nodes := make([]rtl.Node, 0, len(fn.Code))
	for n := range fn.Code {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i] < nodes[j]
	})
	return nodes
}

func transformInstruction(instr rtl.Instruction, alloc *AllocationResult) *ltl.BBlock {
	switch i := instr.(type) {
	case rtl.Inop:
		return &ltl.BBlock{
			Body: []ltl.Instruction{
				ltl.Lbranch{Succ: ltl.Node(i.Succ)},
			},
		}

	case rtl.Iop:
		args := transformRegs(i.Args, alloc)
		dest := alloc.RegToLoc[i.Dest]
		return &ltl.BBlock{
			Body: []ltl.Instruction{
				ltl.Lop{Op: i.Op, Args: args, Dest: dest},
				ltl.Lbranch{Succ: ltl.Node(i.Succ)},
			},
		}

	case rtl.Iload:
		args := transformRegs(i.Args, alloc)
		dest := alloc.RegToLoc[i.Dest]
		return &ltl.BBlock{
			Body: []ltl.Instruction{
				ltl.Lload{Chunk: i.Chunk, Addr: i.Addr, Args: args, Dest: dest},
				ltl.Lbranch{Succ: ltl.Node(i.Succ)},
			},
		}

	case rtl.Istore:
		args := transformRegs(i.Args, alloc)
		src := alloc.RegToLoc[i.Src]
		return &ltl.BBlock{
			Body: []ltl.Instruction{
				ltl.Lstore{Chunk: i.Chunk, Addr: i.Addr, Args: args, Src: src},
				ltl.Lbranch{Succ: ltl.Node(i.Succ)},
			},
		}

	case rtl.Icall:
		args := transformRegs(i.Args, alloc)
		var fn ltl.FunRef
		switch f := i.Fn.(type) {
		case rtl.FunSymbol:
			fn = ltl.FunSymbol{Name: f.Name}
		case rtl.FunReg:
			fn = ltl.FunReg{Loc: alloc.RegToLoc[f.Reg]}
		}
		body := []ltl.Instruction{
			ltl.Lcall{Sig: i.Sig, Fn: fn, Args: args},
		}
		// If call has a destination, move the return value (X0) to it
		if i.Dest != 0 {
			destLoc := alloc.RegToLoc[i.Dest]
			retLoc := ReturnLocation(false) // TODO: handle float returns
			// Only add move if destination is not already X0
			if destLoc != retLoc {
				body = append(body, ltl.Lop{
					Op:   rtl.Omove{},
					Args: []ltl.Loc{retLoc},
					Dest: destLoc,
				})
			}
		}
		body = append(body, ltl.Lbranch{Succ: ltl.Node(i.Succ)})
		return &ltl.BBlock{Body: body}

	case rtl.Itailcall:
		args := transformRegs(i.Args, alloc)
		var fn ltl.FunRef
		switch f := i.Fn.(type) {
		case rtl.FunSymbol:
			fn = ltl.FunSymbol{Name: f.Name}
		case rtl.FunReg:
			fn = ltl.FunReg{Loc: alloc.RegToLoc[f.Reg]}
		}
		return &ltl.BBlock{
			Body: []ltl.Instruction{
				ltl.Ltailcall{Sig: i.Sig, Fn: fn, Args: args},
			},
		}

	case rtl.Ibuiltin:
		args := transformRegs(i.Args, alloc)
		var dest *ltl.Loc
		if i.Dest != nil {
			loc := alloc.RegToLoc[*i.Dest]
			dest = &loc
		}
		return &ltl.BBlock{
			Body: []ltl.Instruction{
				ltl.Lbuiltin{Builtin: i.Builtin, Args: args, Dest: dest},
				ltl.Lbranch{Succ: ltl.Node(i.Succ)},
			},
		}

	case rtl.Icond:
		args := transformRegs(i.Args, alloc)
		return &ltl.BBlock{
			Body: []ltl.Instruction{
				ltl.Lcond{
					Cond:  i.Cond,
					Args:  args,
					IfSo:  ltl.Node(i.IfSo),
					IfNot: ltl.Node(i.IfNot),
				},
			},
		}

	case rtl.Ijumptable:
		arg := alloc.RegToLoc[i.Arg]
		targets := make([]ltl.Node, len(i.Targets))
		for j, t := range i.Targets {
			targets[j] = ltl.Node(t)
		}
		return &ltl.BBlock{
			Body: []ltl.Instruction{
				ltl.Ljumptable{Arg: arg, Targets: targets},
			},
		}

	case rtl.Ireturn:
		var instrs []ltl.Instruction
		// If there's a return value, move it to the return register
		if i.Arg != nil {
			srcLoc := alloc.RegToLoc[*i.Arg]
			destLoc := ReturnLocation(false) // TODO: handle float returns
			// Only add move if not already in return register
			if srcLoc != destLoc {
				instrs = append(instrs, ltl.Lop{
					Op:   rtl.Omove{},
					Args: []ltl.Loc{srcLoc},
					Dest: destLoc,
				})
			}
		}
		instrs = append(instrs, ltl.Lreturn{})
		return &ltl.BBlock{Body: instrs}
	}

	// Fallback: empty block with nop
	return &ltl.BBlock{Body: []ltl.Instruction{ltl.Lnop{}}}
}

func transformRegs(regs []rtl.Reg, alloc *AllocationResult) []ltl.Loc {
	result := make([]ltl.Loc, len(regs))
	for i, r := range regs {
		result[i] = alloc.RegToLoc[r]
	}
	return result
}

// TransformProgram transforms an RTL program to LTL
func TransformProgram(rtlProg *rtl.Program) *ltl.Program {
	ltlProg := &ltl.Program{}

	// Transform globals
	for _, g := range rtlProg.Globals {
		ltlProg.Globals = append(ltlProg.Globals, ltl.GlobVar{
			Name:     g.Name,
			Size:     g.Size,
			Init:     g.Init,
			ReadOnly: g.ReadOnly,
		})
	}

	// Transform functions
	for _, fn := range rtlProg.Functions {
		ltlFn := TransformFunction(&fn)
		ltlProg.Functions = append(ltlProg.Functions, *ltlFn)
	}

	return ltlProg
}
