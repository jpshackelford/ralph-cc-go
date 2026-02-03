   - **FIXED** - fib.c now compiles and runs correctly!
    
    - Final fix (2026-02-02):
      - Implemented Darwin ARM64 variadic calling convention in pkg/linearize/linearize.go
      - Added `knownVariadicFuncs` map with common variadic functions and their fixed arg counts
        - printf(1), fprintf(2), sprintf(2), snprintf(3), scanf(1), etc.
      - Added `isVariadicCall()` function to detect calls to known variadic functions
      - Modified `convertCall()` to handle Darwin variadic calling convention:
        - Fixed args (up to fixedArgs) go in registers (X0, X1, ...)
        - Variadic args go on stack via Lsetstack with SlotOutgoing
        - Uses runtime.GOOS check to apply Darwin-specific behavior
      - Example: `printf("%lld ", first)` now places format string in X0, `first` on stack at [SP+0]
    
    - Progress history:
      - Implemented caller-saved register handling
        - Added LiveAcrossCalls tracking in interference graph (pkg/regalloc/interference.go)
        - Registers live across function calls now assigned to callee-saved registers (X19-X28)
        - Added FirstCalleeSavedColor constant to conventions.go
        - Fixed coalescing to propagate LiveAcrossCalls constraint (pkg/regalloc/irc.go)
        - Added move from X0 to destination after function calls (pkg/regalloc/transform.go)
      - Fixed frame layout bug (root cause of bus error):
        - Problem: callee-save stores at [FP+0..+24] overwrote saved FP/LR at [FP] and [FP+8]
        - Solution: Changed CalleeSaveOffset from 0 to 16 in pkg/stacking/layout.go
    
    - ROOT CAUSE (for reference):
      - macOS ARM64 variadic calling convention differs from Linux ARM64
      - On macOS ARM64, ALL variadic arguments must be passed on the STACK (not in registers)
      - Reference: https://developer.apple.com/documentation/xcode/writing-arm64-code-for-apple-platforms
    
    - Verified output:
      ```
      First 30 Fibonacci numbers:
      0 1 1 2 3 5 8 13 21 34 55 89 144 233 377 610 987 1597 2584 4181 6765 10946 17711 28657 46368 75025 121393 196418 317811 514229
      ```
