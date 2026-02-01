package stacking

import (
	"github.com/raymyers/ralph-cc/pkg/ltl"
	"github.com/raymyers/ralph-cc/pkg/mach"
	"github.com/raymyers/ralph-cc/pkg/rtl"
)

// ARM64 special registers
const (
	FP = ltl.X29 // Frame pointer
	LR = ltl.X30 // Link register (return address)
)

// GeneratePrologue generates the function prologue instructions
// ARM64 prologue:
//  1. Save LR and FP (using STP)
//  2. Set up new FP
//  3. Allocate stack frame
//  4. Save callee-saved registers
func GeneratePrologue(layout *FrameLayout, calleeSave *CalleeSaveInfo) []mach.Instruction {
	var prologue []mach.Instruction

	// The prologue performs:
	// sub sp, sp, #framesize    -- allocate frame
	// stp fp, lr, [sp, #offset] -- save FP and LR
	// add fp, sp, #offset       -- set up frame pointer

	// For simplicity, we represent these as Mop instructions with special ops.
	// In actual assembly generation, these would map to specific ARM64 instructions.

	// 1. Allocate stack frame: sub sp, sp, #TotalSize
	// Represented as addlimm with negative value (sp = sp + (-TotalSize))
	if layout.TotalSize > 0 {
		prologue = append(prologue, mach.Mop{
			Op:   rtl.Oaddlimm{N: -layout.TotalSize},
			Args: nil, // SP is implicit
			Dest: FP,  // Result conceptually goes to SP, represented via FP
		})
	}

	// 2. Save FP and LR
	// stp fp, lr, [sp, #FPoffset]
	// We represent this as two Msetstack operations
	fpSaveOffset := layout.TotalSize - 16 // FP/LR saved at top of frame
	prologue = append(prologue, mach.Msetstack{
		Src: FP,
		Ofs: fpSaveOffset,
		Ty:  ltl.Tlong,
	})
	prologue = append(prologue, mach.Msetstack{
		Src: LR,
		Ofs: fpSaveOffset + 8,
		Ty:  ltl.Tlong,
	})

	// 3. Set up FP: add fp, sp, #offset
	prologue = append(prologue, mach.Mop{
		Op:   rtl.Oaddlimm{N: fpSaveOffset},
		Args: nil, // SP is implicit source
		Dest: FP,
	})

	// 4. Save callee-saved registers
	// For paired saves (STP), we save two at a time
	regs := calleeSave.Regs
	for i := 0; i+1 < len(regs); i += 2 {
		// Save pair at offset from FP
		offset := calleeSave.SaveOffsets[i]
		prologue = append(prologue, mach.Msetstack{
			Src: regs[i],
			Ofs: offset,
			Ty:  ltl.Tlong,
		})
		prologue = append(prologue, mach.Msetstack{
			Src: regs[i+1],
			Ofs: offset + 8,
			Ty:  ltl.Tlong,
		})
	}

	return prologue
}

// GenerateEpilogue generates the function epilogue instructions
// ARM64 epilogue:
//  1. Restore callee-saved registers
//  2. Restore FP and LR
//  3. Deallocate stack frame
//  4. Return
func GenerateEpilogue(layout *FrameLayout, calleeSave *CalleeSaveInfo) []mach.Instruction {
	var epilogue []mach.Instruction

	// 1. Restore callee-saved registers (in reverse order)
	regs := calleeSave.Regs
	for i := len(regs) - 2; i >= 0; i -= 2 {
		offset := calleeSave.SaveOffsets[i]
		epilogue = append(epilogue, mach.Mgetstack{
			Ofs:  offset,
			Ty:   ltl.Tlong,
			Dest: regs[i],
		})
		if i+1 < len(regs) {
			epilogue = append(epilogue, mach.Mgetstack{
				Ofs:  offset + 8,
				Ty:   ltl.Tlong,
				Dest: regs[i+1],
			})
		}
	}

	// 2. Restore FP and LR
	fpSaveOffset := layout.TotalSize - 16
	epilogue = append(epilogue, mach.Mgetstack{
		Ofs:  fpSaveOffset,
		Ty:   ltl.Tlong,
		Dest: FP,
	})
	epilogue = append(epilogue, mach.Mgetstack{
		Ofs:  fpSaveOffset + 8,
		Ty:   ltl.Tlong,
		Dest: LR,
	})

	// 3. Deallocate stack frame: add sp, sp, #TotalSize
	if layout.TotalSize > 0 {
		epilogue = append(epilogue, mach.Mop{
			Op:   rtl.Oaddlimm{N: layout.TotalSize},
			Args: nil,
			Dest: FP, // Conceptually SP
		})
	}

	// 4. Return
	epilogue = append(epilogue, mach.Mreturn{})

	return epilogue
}

// GenerateTailEpilogue generates epilogue for tail calls (without return)
func GenerateTailEpilogue(layout *FrameLayout, calleeSave *CalleeSaveInfo) []mach.Instruction {
	epilogue := GenerateEpilogue(layout, calleeSave)
	// Remove the Mreturn at the end - tail call will replace it
	if len(epilogue) > 0 {
		epilogue = epilogue[:len(epilogue)-1]
	}
	return epilogue
}

// IsLeafFunction returns true if the function doesn't call other functions
// Leaf functions may be able to omit some prologue/epilogue operations
func IsLeafFunction(code []mach.Instruction) bool {
	for _, inst := range code {
		switch inst.(type) {
		case mach.Mcall, mach.Mtailcall:
			return false
		}
	}
	return true
}
