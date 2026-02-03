# Fib Segfault Debug

## Issue

`testdata/example-c/fib_fn.c` segfaulted when run. The program uses `printf` with `%lld` format and a `long long` argument.

## Root Cause

Two related bugs in the assembly generation (asmgen) and stack layout (stacking) packages:

### Bug 1: Prologue Generated Wrong Frame Layout

The `asmgen/transform.go` prologue was:
```asm
stp x29, x30, [sp, #-framesize]!   ; save FP/LR, pre-decrement SP
mov x29, sp                         ; FP = SP
```

But the Mach IR expects:
```asm
sub sp, sp, #framesize              ; allocate frame
stp x29, x30, [sp, #fpOffset]       ; save FP/LR at TOP of frame
add x29, sp, #fpOffset              ; FP = SP + fpOffset
```

With the Mach IR conventions:
- FP points at the saved FP/LR pair, NOT at SP
- Callee-saved registers and locals are at positive offsets from FP
- Outgoing arguments are at negative offsets from FP (near SP)

### Bug 2: OutgoingSlotOffset Was Wrong

`stacking/layout.go` `OutgoingSlotOffset()` returned `slotOffset` directly, assuming SP-relative addressing. But since all stack accesses in asmgen use FP as the base, outgoing slots need FP-relative (negative) offsets.

## Fix

### asmgen/transform.go

Changed `generatePrologue()` to emit:
```asm
sub sp, sp, #framesize           ; allocate frame
stp x29, x30, [sp, #fpOffset]    ; save at top
add x29, sp, #fpOffset           ; FP points to saved FP/LR
```

Where `fpOffset = frameSize - 16`.

Changed `generateEpilogue()` to match:
```asm
ldp x29, x30, [sp, #fpOffset]
add sp, sp, #framesize
ret
```

### stacking/layout.go

Changed `OutgoingSlotOffset()` to compute FP-relative offset:
```go
func (l *FrameLayout) OutgoingSlotOffset(slotOffset int64) int64 {
    fpOffset := l.TotalSize - 16
    return -fpOffset + slotOffset
}
```

This maps `SP + slotOffset` → `FP - fpOffset + slotOffset`.

## Verification

- `fib_fn.c` now runs correctly and prints Fibonacci numbers
- All unit tests pass
- E2E tests pass
- `hello.c` still works

## Test Updated

Updated `stacking/slots_test.go` `TestTranslateSlotOffset` to expect the new FP-relative offset computation.
