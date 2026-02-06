# Multi-line Macro Arguments Fix

## Problem
The preprocessor processed tokens line-by-line, causing macro invocations that span multiple lines to fail with "unterminated macro argument list" error.

Example that failed:
```c
assert(x == 1 ? 1
              : 0);
```

SQLite uses many multi-line macro calls (e.g., line 24109 in sqlite3.c).

## Solution
Modified `preprocessContent()` in `pkg/cpp/preprocess.go` to track parenthesis depth and continue collecting tokens across newlines when inside unbalanced parentheses.

Key changes:
1. Added `parenDepth` counter to track `(` and `)` 
2. When hitting newline with parenDepth > 0 (and not a directive line), continue collecting instead of processing
3. Added `isDirectiveLine()` helper to avoid breaking directive handling

## Additional Fix: Macro Redefinition
System headers often redefine macros (e.g., CHAR_BIT defined in multiple limit headers). Changed macro redefinition from error to warning for compatibility.

## Tests Added
- `TestPreprocessor_MultiLineMacroArgs` - basic multi-line args
- `TestPreprocessor_NestedMultiLineMacroArgs` - nested parens across lines
- Updated `TestRedefinitionDifferent` to expect warning behavior

## Result
Milestone 1 verification PASSES:
```
./bin/ralph-cc -E checkouts/sqlite-amalgamation-3470200/sqlite3.c > /dev/null 2>&1 && echo PASS
```

## Files Changed
- `pkg/cpp/preprocess.go` - multi-line macro handling
- `pkg/cpp/macro.go` - redefinition warning instead of error  
- `pkg/cpp/macro_test.go` - updated test
- `pkg/cpp/preprocess_test.go` - added tests
