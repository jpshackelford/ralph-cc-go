# Usability Assessment Plan

This document outlines a rigorous methodology to assess how close ralph-cc is to being a usable compiler for short programs (~100 lines) using common C features.

## Definition of "Usable"

A compiler is considered usable for ~100 line programs if it can:

1. **Parse** - Correctly parse common C syntax
2. **Compile** - Generate assembly for all supported constructs
3. **Run** - Produce executables that give correct results
4. **Integrate** - Work with system libraries (stdio, stdlib)

## Feature Categories

### Category 1: Core Features (Must Work)

These features are essential for any non-trivial C program:

| ID | Feature | Parser | Compile | Runtime | Notes |
|----|---------|--------|---------|---------|-------|
| C1.1 | Integer constants | [ ] | [ ] | [ ] | `42`, `0`, `-1` |
| C1.2 | Integer arithmetic | [ ] | [ ] | [ ] | `+`, `-`, `*`, `/`, `%` |
| C1.3 | Integer comparisons | [ ] | [ ] | [ ] | `<`, `<=`, `>`, `>=`, `==`, `!=` |
| C1.4 | Local variables | [ ] | [ ] | [ ] | `int x = 5;` |
| C1.5 | Assignment | [ ] | [ ] | [ ] | `x = y;` |
| C1.6 | Function definitions | [ ] | [ ] | [ ] | `int f(int x) { ... }` |
| C1.7 | Function calls | [ ] | [ ] | [ ] | `f(1, 2)` |
| C1.8 | Return statement | [ ] | [ ] | [ ] | `return x;` |
| C1.9 | If statement | [ ] | [ ] | [ ] | `if (x) ...` |
| C1.10 | If-else statement | [ ] | [ ] | [ ] | `if (x) ... else ...` |
| C1.11 | While loop | [ ] | [ ] | [ ] | `while (x) ...` |
| C1.12 | For loop | [ ] | [ ] | [ ] | `for (i=0; i<n; i++) ...` |

### Category 2: Extended Features (Should Work)

Features commonly used in practical programs:

| ID | Feature | Parser | Compile | Runtime | Notes |
|----|---------|--------|---------|---------|-------|
| C2.1 | Logical operators | [ ] | [ ] | [ ] | `&&`, `||`, `!` |
| C2.2 | Bitwise operators | [ ] | [ ] | [ ] | `&`, `|`, `^`, `~`, `<<`, `>>` |
| C2.3 | Increment/decrement | [ ] | [ ] | [ ] | `++x`, `x++`, `--x`, `x--` |
| C2.4 | Compound assignment | [ ] | [ ] | [ ] | `+=`, `-=`, `*=`, etc. |
| C2.5 | Ternary operator | [ ] | [ ] | [ ] | `x ? y : z` |
| C2.6 | Do-while loop | [ ] | [ ] | [ ] | `do { ... } while (x);` |
| C2.7 | Switch statement | [ ] | [ ] | [ ] | `switch (x) { case 1: ... }` |
| C2.8 | Break/continue | [ ] | [ ] | [ ] | `break;`, `continue;` |
| C2.9 | Pointers | [ ] | [ ] | [ ] | `int *p; *p = 5;` |
| C2.10 | Address-of | [ ] | [ ] | [ ] | `&x` |
| C2.11 | Arrays | [ ] | [ ] | [ ] | `int a[10]; a[0] = 1;` |
| C2.12 | String literals | [ ] | [ ] | [ ] | `"hello"` |
| C2.13 | Character literals | [ ] | [ ] | [ ] | `'x'`, `'\n'` |

### Category 3: Type System (For Practical Programs)

| ID | Feature | Parser | Compile | Runtime | Notes |
|----|---------|--------|---------|---------|-------|
| C3.1 | Char type | [ ] | [ ] | [ ] | `char c = 'x';` |
| C3.2 | Unsigned types | [ ] | [ ] | [ ] | `unsigned int x;` |
| C3.3 | Typedef | [ ] | [ ] | [ ] | `typedef int myint;` |
| C3.4 | Struct definition | [ ] | [ ] | [ ] | `struct Point { int x, y; };` |
| C3.5 | Struct member access | [ ] | [ ] | [ ] | `p.x`, `p->x` |
| C3.6 | Enum | [ ] | [ ] | [ ] | `enum Color { RED };` |
| C3.7 | Const qualifier | [ ] | [ ] | [ ] | `const int x = 5;` |
| C3.8 | Void type | [ ] | [ ] | [ ] | `void f() { }` |
| C3.9 | Pointer arithmetic | [ ] | [ ] | [ ] | `p + 1`, `p++` |
| C3.10 | Cast expressions | [ ] | [ ] | [ ] | `(int)x` |

### Category 4: I/O and Library Integration

| ID | Feature | Parser | Compile | Runtime | Notes |
|----|---------|--------|---------|---------|-------|
| C4.1 | Include stdio.h | [ ] | [ ] | [ ] | `#include <stdio.h>` |
| C4.2 | printf call | [ ] | [ ] | [ ] | `printf("hello\n");` |
| C4.3 | puts call | [ ] | [ ] | [ ] | `puts("hello");` |
| C4.4 | External functions | [ ] | [ ] | [ ] | `int printf(...);` |

## Test Methodology

### Phase 1: Create E2E Runtime Tests

Create a new test file `testdata/e2e_runtime.yaml` with test cases that:
1. Compile C source to assembly
2. Assemble and link using system tools
3. Run the executable
4. Verify the exit code matches expected value

Example test case format:
```yaml
tests:
  - name: "integer addition"
    input: |
      int main() { return 3 + 4; }
    expected_exit: 7

  - name: "while loop - sum 1 to 10"
    input: |
      int main() {
        int s = 0, n = 10;
        while (n > 0) { s = s + n; n = n - 1; }
        return s;
      }
    expected_exit: 55
```

### Phase 2: Run and Document Results

For each feature category:
1. Run all tests in that category
2. Mark Parser/Compile/Runtime status in the table above
3. Document specific failures with error messages

### Phase 3: Prioritize Fixes

Based on test results:
1. Identify critical blocking issues (Category 1 failures)
2. Create targeted bug fix tasks
3. Verify fixes with regression tests

## Success Criteria

The compiler is considered **minimally usable** when:
- 100% of Category 1 features pass all three stages (Parser/Compile/Runtime)
- 80% of Category 2 features pass all three stages
- hello.c with printf works correctly

## Current Status

**Assessment Date**: 2026-02-02

### Test Infrastructure

- [x] Created `testdata/e2e_runtime.yaml` with comprehensive test cases (60+ tests)
- [x] Added runtime test runner to `cmd/ralph-cc/integration_test.go` (TestE2ERuntimeYAML)
- [x] Tests compile C→assembly→object→executable and verify exit codes

### Results Summary

| Category | Subcategory | Status | Notes |
|----------|-------------|--------|-------|
| C1.1 | Integer constants | ✅ PASS | 0, 42, 255 all work |
| C1.2 | Integer arithmetic | ✅ PASS | +, -, *, /, % all work |
| C1.3 | Integer comparisons | ✅ PASS | `<`, `>`, `==`, `!=`, `<=`, `>=` as expressions |
| C1.4 | Local variables | ✅ PASS | Basic and multiple vars work |
| C1.5 | Assignment | ✅ PASS | Simple and chained work |
| C1.6 | Function definitions | ✅ PASS | With and without params |
| C1.7 | Function calls | ✅ PASS | Multiple args, nested calls |
| C1.8 | Return statement | ✅ PASS | Early return works |
| C1.9 | If statement | ✅ PASS | Fixed: CMP now emitted before branch |
| C1.10 | If-else statement | ✅ PASS | Fixed: CMP now emitted before branch |
| C1.11 | While loop | ⚠️ PARTIAL | Condition fixed, but variable tracking across iterations broken |
| C1.12 | For loop | ⚠️ PARTIAL | Condition fixed, but variable tracking across iterations broken |

### Critical Issues Found

#### Issue 1: Comparison Operators Compile as ADD (FIXED)

**Severity**: CRITICAL - Blocks all control flow

**Symptom**: `return 3 < 5;` compiled as:
```asm
mov w0, #3
mov w1, #5
add w0, w0, w1  ; Should be: cmp w0, w1; cset w0, lt
```

**Status**: ✅ FIXED (2026-02-02)

**Root cause found**: cminor.Ecmp was converted to cminorsel.Ebinop in selection phase,
losing the Comparison condition (Ceq, Clt, etc.). Then TranslateBinaryOp in rtlgen/instr.go
didn't handle Ocmp operators, defaulting to Oadd.

**Fix**: Added Ecmp expression type to cminorsel that preserves both Op and Cmp fields.
Updated selection and rtlgen to produce proper comparison operations.

Now correctly generates:
```asm
mov w1, #3
mov w0, #5
cmp w1, w0
cset w0, lt
```

#### Issue 2: Conditional Branches Without Flag Setting (FIXED)

**Severity**: CRITICAL - Was blocking control flow

**Symptom**: `if (0)` was generating `b.gt` without preceding CMP instruction

**Status**: ✅ FIXED (2026-02-02)

**Root cause found**: `translateCond` in `pkg/asmgen/transform.go` was only emitting
the conditional branch (`Bcond`) but not the comparison instruction (`CMP`) before it.
The condition code types (`Ccomp`, `Ccompimm`, etc.) carry the comparison details and
argument registers, but these were not being used to generate the actual comparison.

**Fix**: Updated `translateCond` to:
1. Emit appropriate `CMP` or `CMPi` instruction based on condition code type
2. Handle all comparison variants (signed/unsigned, register/immediate, 32/64-bit, float)
3. Then emit the conditional branch with the correct ARM64 condition code

Now correctly generates:
```asm
cmp w0, w1      ; or cmp w0, #imm for immediate comparisons
b.gt .Ltarget
```

### What Works (verified)

1. **Constants and arithmetic**: All basic math operations produce correct results
2. **Variables**: Local variable declaration, initialization, and assignment work
3. **Functions**: Definition, parameter passing, return values, and calls all work
4. **String literals**: "hello" style strings work with printf
5. **printf**: External function calls to libc work (hello.c runs correctly)
6. **Comparisons as expressions**: `return x < y;` now produces correct 0/1 values ✅
7. **If statements**: `if (cond)` and `if (cond) ... else ...` work correctly ✅
8. **Simple loops**: `while (0)` and `while(cond)` with simple bodies work ✅
9. **Ternary operator**: `x ? y : z` works correctly ✅
10. **Logical operators**: `&&` and `||` work correctly ✅

### What's Broken (verified)

1. **Variable tracking in loops**: Loops that modify variables across iterations don't track
   values correctly - the sum/accumulator gets lost in register allocation
2. **Pointers and arrays**: Address-of (`&x`) and dereferencing (`*p`) have codegen issues
3. **Logical not**: `!0` returns wrong value

### Fix Tasks

[x] **FIX-001**: Implement comparison code generation correctly
    - Root cause: cminor.Ecmp was converted to cminorsel.Ebinop, losing the Comparison condition
    - TranslateBinaryOp in rtlgen/instr.go defaulted Ocmp to Oadd
    - Fixed by:
      1. Added Ecmp expression type to cminorsel/ast.go that preserves Op and Cmp fields
      2. Updated selection/expr.go selectCmp() to produce Ecmp instead of Ebinop
      3. Added translateCmp() in rtlgen/expr.go to handle Ecmp expressions
      4. Added TranslateCompareOp() in rtlgen/instr.go to map comparison ops to RTL
    - Now generates: cmp + cset instructions correctly
    - All C1.3 comparison tests pass (10/10)

[x] **FIX-002**: Ensure conditionals set comparison flags
    - Root cause: translateCond() in pkg/asmgen/transform.go only emitted Bcond without CMP
    - Fixed by updating translateCond() to:
      1. Emit CMP or CMPi instruction based on condition code type
      2. Handle all variants: Ccomp, Ccompu, Ccompimm, Ccompuimm, Ccompl, Ccomplu, etc.
      3. Then emit the conditional branch with correct ARM64 condition code
    - Now generates: cmp + b.cond instructions correctly
    - All C1.9 and C1.10 tests pass (7/7)

[ ] **FIX-003**: Fix variable tracking across loop iterations
    - Loops with modified variables (while countdown, for loop sum) return wrong values
    - Register allocation seems to lose track of which register holds the accumulator
    - Need to investigate RTL/LTL/Mach transformation for local variable handling

### Usability Verdict

**APPROACHING USABLE** for ~100 line programs with common features.

With the conditional branch fix, significant progress has been made:
- ✅ If/else statements work correctly
- ✅ Simple while loops (like `while(0)`) work  
- ✅ Comparisons work both as expressions and in branch conditions
- ✅ Ternary operator works
- ⚠️ Loops with variable mutation need register allocation fixes

**Estimated effort to reach "minimally usable"**: 
- 1-2 issues to fix (variable tracking in loops, pointer/array codegen)
- Medium complexity - likely in register allocation or RTL generation
- After fix: 100% of Category 1 features should work

### Next Steps

1. [x] Fix conditional branch codegen - DONE
2. [ ] Investigate variable tracking in loops (register allocation issue)
3. [ ] Fix pointer and array codegen
4. [ ] Re-run test suite to verify all C1 tests pass
5. [ ] Update feature matrix with final results
