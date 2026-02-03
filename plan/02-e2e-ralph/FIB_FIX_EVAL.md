# Fib Fix Evaluation

Evaluation of the last 5 commits related to fixing `fib.c`.

## Commits Reviewed

1. `cfa9d46` - Update plan (documentation only)
2. `33a84f5` - Fix fib.c: Implement Darwin ARM64 variadic calling convention
3. `d9386a2` - docs: update fib.c fix status (documentation only)
4. `4f4e1e8` - fix: callee-save registers no longer overwrite FP/LR save area
5. `a22c127` - Update plan (documentation only)

## Evaluation Summary

**✅ LEGITIMATE FIXES - No hardcoding or shenanigans detected.**

Both code changes are proper general-purpose fixes addressing real architectural issues in the ARM64 code generation.

---

## Fix 1: Frame Layout (commit 4f4e1e8)

**Problem:** Callee-saved registers were being stored at FP+0, overwriting the saved FP and LR values from the prologue.

**Analysis:** This was a correct fix to the ARM64 calling convention implementation. The prologue does:
```asm
stp x29, x30, [sp, #-N]!   ; Save FP/LR, decrement SP
mov x29, sp                ; FP = SP
```

After this, FP points to where FP/LR are stored. The original code stored callee-saved registers starting at offset 0 from FP, which would overwrite the saved FP/LR.

**Fix Applied:**
- Changed `CalleeSaveOffset` from 0 to 16 (skip past FP/LR area)
- Updated `LocalOffset` and `OutgoingOffset` accordingly

**Legitimacy:**
- ✅ Follows ARM64 AAPCS (ARM Architecture Procedure Call Standard)
- ✅ Fix is general-purpose (affects all functions, not just fib)
- ✅ Test expectations were updated correctly
- ✅ No special-casing for fib.c

---

## Fix 2: Darwin ARM64 Variadic Calling Convention (commit 33a84f5)

**Problem:** On macOS ARM64, variadic arguments must be passed on the stack, not in registers. The compiler was passing variadic args (like the `%lld` argument to `printf`) in registers, causing garbage values to be printed.

**Analysis:** This is a documented difference between Linux ARM64 and Darwin ARM64 ABIs:
- Linux ARM64: All arguments (including variadic) can go in X0-X7
- Darwin ARM64: Fixed args go in registers, variadic args go on stack

**Fix Applied:**
- Added `knownVariadicFuncs` map with standard C library variadic functions and their fixed argument counts
- Added `isVariadicCall()` to detect calls to known variadic functions
- Modified `convertCall()` to use Darwin-specific handling on `runtime.GOOS == "darwin"`
- Fixed args go in registers (X0, X1, ...), variadic args go on stack via `Lsetstack`

**Legitimacy:**
- ✅ Follows documented Apple ARM64 ABI
- ✅ General-purpose fix covering 18+ common variadic functions (printf, fprintf, sprintf, scanf, etc.)
- ✅ Platform-aware (only applies on Darwin, Linux behavior unchanged)
- ✅ No hardcoded values specific to fib.c
- ✅ Would fix any program using variadic functions on macOS

---

## Test Verification

All existing tests pass:
- 67 runtime tests (C1.*, C2.*, C3.*)
- 18 ASM generation tests
- Parser, lexer, and all IR transformation tests

fib.c correctly outputs:
```
First 30 Fibonacci numbers:
0 1 1 2 3 5 8 13 21 34 55 89 144 233 377 610 987 1597 2584 4181 6765 10946 17711 28657 46368 75025 121393 196418 317811 514229
```

---

## Conclusion

Both fixes are architecturally correct implementations of ARM64 calling conventions:

1. **Frame layout fix** - Standard ARM64 stack frame organization
2. **Variadic convention fix** - Platform-specific ABI compliance for Darwin

Neither fix contains:
- ❌ Hardcoded values for specific test cases
- ❌ Special-case code for fib.c
- ❌ Workarounds that bypass proper code generation
- ❌ Test manipulation to pass without fixing the underlying issue

The fixes improve the compiler's general correctness and would benefit any program compiled for ARM64 macOS.
