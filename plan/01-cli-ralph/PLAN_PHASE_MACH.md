# Phase: Mach Code Generation (Stacking)

**Transformation:** Linear → Mach
**Prereqs:** Linear code generation (PLAN_PHASE_LINEAR.md)

Mach is a near-assembly representation with concrete activation record layout. This is the last intermediate language before assembly.

## Key CompCert Files to Study

| File | Purpose |
|------|---------|
| `backend/Mach.v` | Mach AST definition |
| `backend/Stacking.v` | Activation record layout |
| `backend/Stackingproof.v` | Correctness proof |
| `backend/Bounds.v` | Stack frame size computation |
| `aarch64/Stacklayout.v` | ARM64 stack layout |
| `backend/PrintMach.ml` | OCaml pretty-printer |
| `aarch64/Machregs.v` | Machine register definitions |

## Overview

Stacking transforms Linear to Mach by:
1. **Frame layout** - Compute activation record structure
2. **Stack access** - Replace abstract slots with concrete offsets
3. **Prologue/epilogue** - Add function entry/exit code
4. **Callee-save handling** - Save/restore preserved registers

## Milestone 1: Mach AST Definition

**Goal:** Define the Mach AST with concrete stack layout

### Tasks

- [x] Create `pkg/mach/ast.go` with node interfaces
- [x] Define machine registers (same as LTL)
- [x] Define Mach instructions:
  - [x] `Mgetstack` - Load from stack at concrete offset
  - [x] `Msetstack` - Store to stack at concrete offset
  - [x] `Mgetparam` - Load parameter from caller's frame
  - [x] `Mop` - Operation
  - [x] `Mload` - Memory load
  - [x] `Mstore` - Memory store
  - [x] `Mcall` - Function call
  - [x] `Mtailcall` - Tail call
  - [x] `Mbuiltin` - Builtin
  - [x] `Mlabel` - Label
  - [x] `Mgoto` - Unconditional jump
  - [x] `Mcond` - Conditional branch
  - [x] `Mjumptable` - Indexed jump
  - [x] `Mreturn` - Return
- [x] Define function structure:
  - [x] Code
  - [x] Stack frame size
  - [x] Used callee-save registers
- [x] Add tests for AST construction

## Milestone 2: Stack Frame Layout

**Goal:** Compute concrete stack frame layout

### Tasks

- [x] Create `pkg/stacking/layout.go`
- [x] Define frame structure (ARM64):
  ```
  +---------------------------+  <- old SP (caller's frame)
  | Return address (LR)       |
  | Saved FP                  |
  +---------------------------+  <- FP
  | Callee-saved registers    |
  | Local variables           |
  | Outgoing arguments        |
  +---------------------------+  <- SP (16-byte aligned)
  ```
- [x] Compute frame sections:
  - [x] Callee-save area size
  - [x] Local variable area size
  - [x] Outgoing argument area size
- [x] Handle alignment:
  - [x] 16-byte stack alignment (ARM64)
  - [x] Per-variable alignment
- [x] Compute total frame size
- [x] Add tests for layout computation

## Milestone 3: Stack Slot Translation

**Goal:** Translate abstract slots to concrete offsets

### Tasks

- [x] Create `pkg/stacking/slots.go`
- [x] Map Local slots to frame offsets
- [x] Map Outgoing slots to bottom of frame
- [x] Map Incoming slots to caller's frame:
  - [x] Above our frame pointer
  - [x] Depends on calling convention
- [x] Generate stack access instructions:
  - [x] `Lgetstack Local` → `Mgetstack fp+offset`
  - [x] `Lsetstack Local` → `Msetstack fp+offset`
  - [x] `Lgetstack Incoming` → `Mgetparam offset`
- [x] Add tests for slot translation

## Milestone 4: Callee-Save Register Handling

**Goal:** Save and restore callee-saved registers

### Tasks

- [x] Create `pkg/stacking/calleesave.go`
- [x] Identify used callee-saved registers:
  - [x] ARM64: X19-X28, D8-D15
  - [x] Scan function for uses
- [x] Compute save/restore locations:
  - [x] Sequential in callee-save area
  - [x] Paired stores for ARM64 (STP/LDP)
- [x] Generate prologue saves:
  - [x] At function entry
  - [x] After frame setup
- [x] Generate epilogue restores:
  - [x] Before return
  - [x] Before tail call
- [x] Add tests for callee-save handling

## Milestone 5: Prologue and Epilogue

**Goal:** Generate function entry and exit code

### Tasks

- [x] Create `pkg/stacking/prolog.go`
- [x] Generate prologue:
  - [x] Save link register (return address)
  - [x] Save frame pointer
  - [x] Set up new frame pointer
  - [x] Allocate stack frame
  - [x] Save callee-saved registers
- [x] Generate epilogue:
  - [x] Restore callee-saved registers
  - [x] Restore frame pointer
  - [x] Deallocate stack frame
  - [x] Return (restore PC from LR)
- [x] Handle leaf functions:
  - [x] May omit frame pointer setup
  - [x] May skip saving LR if not used
- [x] Add tests for prologue/epilogue

## Milestone 6: Instruction Translation

**Goal:** Translate Linear instructions to Mach

### Tasks

- [x] Create `pkg/stacking/transform.go`
- [x] Translate stack operations with concrete offsets
- [x] Translate other instructions (mostly unchanged)
- [x] Insert prologue at function entry
- [x] Insert epilogue before returns
- [x] Handle tail calls (epilogue before call)
- [x] Add tests for instruction translation

## Milestone 7: CLI Integration & Testing

**Goal:** Wire Mach generation to CLI, test against CompCert

### Tasks

- [x] Add `-dmach` flag implementation
- [x] Create `pkg/mach/printer.go` matching CompCert output format
- [x] Create test cases (unit tests in pkg/stacking and pkg/mach)
- [x] Add CLI tests for -dmach flag
- [ ] Test against CompCert output (using container-use) - optional verification
- [ ] Document any intentional deviations - optional

## Test Strategy

1. **Unit tests:** Frame layout, slot translation
2. **Stack correctness:** Verify offsets are correct
3. **Callee-save:** Verify all used regs saved/restored
4. **Alignment:** Verify 16-byte alignment maintained
5. **Golden tests:** Compare against CompCert's `-dmach`

## Expected Output Format

Mach output should match CompCert's `.mach` format:
```
f:
  sub sp, sp, #32
  stp fp, lr, [sp, #16]
  add fp, sp, #16
  ...
  ldp fp, lr, [sp, #16]
  add sp, sp, #32
  ret
```

## ARM64 Frame Notes

- FP (X29) points into frame
- SP must be 16-byte aligned
- LR (X30) contains return address
- Pairs of registers stored with STP/LDP

## Dependencies

- `pkg/linear` - Input AST (from PLAN_PHASE_LINEAR.md)
- Target architecture (ARM64)
