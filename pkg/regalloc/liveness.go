// Package regalloc implements register allocation for the RTL to LTL transformation.
// It uses liveness analysis and graph coloring (IRC algorithm).
package regalloc

import (
	"github.com/raymyers/ralph-cc/pkg/rtl"
)

// LivenessInfo holds liveness information for a function
type LivenessInfo struct {
	// LiveIn maps nodes to the set of registers live at entry
	LiveIn map[rtl.Node]RegSet
	// LiveOut maps nodes to the set of registers live at exit
	LiveOut map[rtl.Node]RegSet
	// Def maps nodes to registers defined at that node
	Def map[rtl.Node]RegSet
	// Use maps nodes to registers used at that node
	Use map[rtl.Node]RegSet
}

// RegSet represents a set of pseudo-registers
type RegSet map[rtl.Reg]bool

// NewRegSet creates a new empty register set
func NewRegSet() RegSet {
	return make(RegSet)
}

// Add adds a register to the set
func (s RegSet) Add(r rtl.Reg) {
	s[r] = true
}

// Contains returns true if the register is in the set
func (s RegSet) Contains(r rtl.Reg) bool {
	return s[r]
}

// Union returns the union of two sets
func (s RegSet) Union(other RegSet) RegSet {
	result := NewRegSet()
	for r := range s {
		result[r] = true
	}
	for r := range other {
		result[r] = true
	}
	return result
}

// Minus returns s - other (set difference)
func (s RegSet) Minus(other RegSet) RegSet {
	result := NewRegSet()
	for r := range s {
		if !other[r] {
			result[r] = true
		}
	}
	return result
}

// Equal returns true if two sets are equal
func (s RegSet) Equal(other RegSet) bool {
	if len(s) != len(other) {
		return false
	}
	for r := range s {
		if !other[r] {
			return false
		}
	}
	return true
}

// Copy returns a copy of the set
func (s RegSet) Copy() RegSet {
	result := NewRegSet()
	for r := range s {
		result[r] = true
	}
	return result
}

// Slice returns the registers as a slice
func (s RegSet) Slice() []rtl.Reg {
	result := make([]rtl.Reg, 0, len(s))
	for r := range s {
		result = append(result, r)
	}
	return result
}

// ComputeDefUse computes the def and use sets for each instruction in the function
func ComputeDefUse(fn *rtl.Function) (def, use map[rtl.Node]RegSet) {
	def = make(map[rtl.Node]RegSet)
	use = make(map[rtl.Node]RegSet)

	for node, instr := range fn.Code {
		def[node] = NewRegSet()
		use[node] = NewRegSet()

		switch i := instr.(type) {
		case rtl.Inop:
			// No def/use
		case rtl.Iop:
			// Uses args, defines dest
			for _, arg := range i.Args {
				use[node].Add(arg)
			}
			def[node].Add(i.Dest)
		case rtl.Iload:
			// Uses address args, defines dest
			for _, arg := range i.Args {
				use[node].Add(arg)
			}
			def[node].Add(i.Dest)
		case rtl.Istore:
			// Uses address args and src
			for _, arg := range i.Args {
				use[node].Add(arg)
			}
			use[node].Add(i.Src)
		case rtl.Icall:
			// Uses args and possibly function pointer, defines dest
			for _, arg := range i.Args {
				use[node].Add(arg)
			}
			if fr, ok := i.Fn.(rtl.FunReg); ok {
				use[node].Add(fr.Reg)
			}
			if i.Dest != 0 {
				def[node].Add(i.Dest)
			}
		case rtl.Itailcall:
			// Uses args and possibly function pointer
			for _, arg := range i.Args {
				use[node].Add(arg)
			}
			if fr, ok := i.Fn.(rtl.FunReg); ok {
				use[node].Add(fr.Reg)
			}
		case rtl.Ibuiltin:
			// Uses args, defines dest if present
			for _, arg := range i.Args {
				use[node].Add(arg)
			}
			if i.Dest != nil {
				def[node].Add(*i.Dest)
			}
		case rtl.Icond:
			// Uses args
			for _, arg := range i.Args {
				use[node].Add(arg)
			}
		case rtl.Ijumptable:
			// Uses index register
			use[node].Add(i.Arg)
		case rtl.Ireturn:
			// Uses return value register if present
			if i.Arg != nil {
				use[node].Add(*i.Arg)
			}
		}
	}

	return def, use
}

// AnalyzeLiveness computes liveness information for the function using fixed-point iteration
func AnalyzeLiveness(fn *rtl.Function) *LivenessInfo {
	def, use := ComputeDefUse(fn)

	// Initialize live sets
	liveIn := make(map[rtl.Node]RegSet)
	liveOut := make(map[rtl.Node]RegSet)

	for node := range fn.Code {
		liveIn[node] = NewRegSet()
		liveOut[node] = NewRegSet()
	}

	// Function parameters are live at entry
	// (handled implicitly by the dataflow equations)

	// Fixed-point iteration
	changed := true
	for changed {
		changed = false

		for node, instr := range fn.Code {
			// Compute new live_out: union of live_in of all successors
			newLiveOut := NewRegSet()
			for _, succ := range instr.Successors() {
				for r := range liveIn[succ] {
					newLiveOut[r] = true
				}
			}

			// Compute new live_in: use[n] ∪ (live_out[n] - def[n])
			newLiveIn := use[node].Union(newLiveOut.Minus(def[node]))

			// Check if anything changed
			if !liveIn[node].Equal(newLiveIn) || !liveOut[node].Equal(newLiveOut) {
				changed = true
				liveIn[node] = newLiveIn
				liveOut[node] = newLiveOut
			}
		}
	}

	return &LivenessInfo{
		LiveIn:  liveIn,
		LiveOut: liveOut,
		Def:     def,
		Use:     use,
	}
}
