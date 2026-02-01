package regalloc

import (
	"github.com/raymyers/ralph-cc/pkg/ltl"
)

// ARM64 calling convention definitions

// CallerSavedRegs are registers that may be clobbered by function calls
// The caller must save these if they need the values after the call
var CallerSavedRegs = []ltl.MReg{
	ltl.X0, ltl.X1, ltl.X2, ltl.X3, ltl.X4, ltl.X5, ltl.X6, ltl.X7,
	ltl.X8, ltl.X9, ltl.X10, ltl.X11, ltl.X12, ltl.X13, ltl.X14, ltl.X15,
	ltl.X16, ltl.X17, ltl.X18,
}

// CalleeSavedRegs are registers that functions must preserve
// If a function uses these, it must save/restore them
var CalleeSavedRegs = []ltl.MReg{
	ltl.X19, ltl.X20, ltl.X21, ltl.X22, ltl.X23, ltl.X24, ltl.X25, ltl.X26, ltl.X27, ltl.X28,
}

// CallerSavedFloatRegs are caller-saved floating-point registers
var CallerSavedFloatRegs = []ltl.MReg{
	ltl.D0, ltl.D1, ltl.D2, ltl.D3, ltl.D4, ltl.D5, ltl.D6, ltl.D7,
	ltl.D16, ltl.D17, ltl.D18, ltl.D19, ltl.D20, ltl.D21, ltl.D22, ltl.D23,
	ltl.D24, ltl.D25, ltl.D26, ltl.D27, ltl.D28, ltl.D29, ltl.D30, ltl.D31,
}

// CalleeSavedFloatRegs are callee-saved floating-point registers
var CalleeSavedFloatRegs = []ltl.MReg{
	ltl.D8, ltl.D9, ltl.D10, ltl.D11, ltl.D12, ltl.D13, ltl.D14, ltl.D15,
}

// IntArgRegs are registers used for integer arguments
var IntArgRegs = []ltl.MReg{ltl.X0, ltl.X1, ltl.X2, ltl.X3, ltl.X4, ltl.X5, ltl.X6, ltl.X7}

// FloatArgRegs are registers used for floating-point arguments
var FloatArgRegs = []ltl.MReg{ltl.D0, ltl.D1, ltl.D2, ltl.D3, ltl.D4, ltl.D5, ltl.D6, ltl.D7}

// IntReturnReg is the register for integer return values
const IntReturnReg = ltl.X0

// FloatReturnReg is the register for floating-point return values
const FloatReturnReg = ltl.D0

// AllocatableIntRegs are integer registers available for allocation
// Excludes: X29 (FP), X30 (LR), X16/X17 (IP0/IP1 used by linker)
var AllocatableIntRegs = []ltl.MReg{
	ltl.X0, ltl.X1, ltl.X2, ltl.X3, ltl.X4, ltl.X5, ltl.X6, ltl.X7,
	ltl.X8, ltl.X9, ltl.X10, ltl.X11, ltl.X12, ltl.X13, ltl.X14, ltl.X15,
	ltl.X18, // Platform register, but often available
	ltl.X19, ltl.X20, ltl.X21, ltl.X22, ltl.X23, ltl.X24, ltl.X25, ltl.X26, ltl.X27, ltl.X28,
}

// AllocatableFloatRegs are floating-point registers available for allocation
var AllocatableFloatRegs = []ltl.MReg{
	ltl.D0, ltl.D1, ltl.D2, ltl.D3, ltl.D4, ltl.D5, ltl.D6, ltl.D7,
	ltl.D8, ltl.D9, ltl.D10, ltl.D11, ltl.D12, ltl.D13, ltl.D14, ltl.D15,
	ltl.D16, ltl.D17, ltl.D18, ltl.D19, ltl.D20, ltl.D21, ltl.D22, ltl.D23,
	ltl.D24, ltl.D25, ltl.D26, ltl.D27, ltl.D28, ltl.D29, ltl.D30, ltl.D31,
}

// NumAllocatableIntRegs is the number of allocatable integer registers (K for graph coloring)
const NumAllocatableIntRegs = 27

// NumAllocatableFloatRegs is the number of allocatable float registers
const NumAllocatableFloatRegs = 32

// ArgLocation returns the location for the i-th integer argument
func ArgLocation(i int, isFloat bool) ltl.Loc {
	if isFloat {
		if i < len(FloatArgRegs) {
			return ltl.R{Reg: FloatArgRegs[i]}
		}
		// Stack slot for overflow arguments
		return ltl.S{Slot: ltl.SlotIncoming, Ofs: int64(i-len(FloatArgRegs)) * 8, Ty: ltl.Tfloat}
	}
	if i < len(IntArgRegs) {
		return ltl.R{Reg: IntArgRegs[i]}
	}
	// Stack slot for overflow arguments
	return ltl.S{Slot: ltl.SlotIncoming, Ofs: int64(i-len(IntArgRegs)) * 8, Ty: ltl.Tlong}
}

// ReturnLocation returns the location for the return value
func ReturnLocation(isFloat bool) ltl.Loc {
	if isFloat {
		return ltl.R{Reg: FloatReturnReg}
	}
	return ltl.R{Reg: IntReturnReg}
}

// IsCallerSaved returns true if the register is caller-saved
func IsCallerSaved(r ltl.MReg) bool {
	for _, cr := range CallerSavedRegs {
		if r == cr {
			return true
		}
	}
	for _, cr := range CallerSavedFloatRegs {
		if r == cr {
			return true
		}
	}
	return false
}

// IsCalleeSaved returns true if the register is callee-saved
func IsCalleeSaved(r ltl.MReg) bool {
	for _, cr := range CalleeSavedRegs {
		if r == cr {
			return true
		}
	}
	for _, cr := range CalleeSavedFloatRegs {
		if r == cr {
			return true
		}
	}
	return false
}
