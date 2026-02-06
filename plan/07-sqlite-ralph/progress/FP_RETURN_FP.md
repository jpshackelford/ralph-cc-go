# Function Pointer Returning Function Pointer

## Problem

Parser fails on struct fields with function pointers that return function pointers:
```c
void (*(*xDlSym)(sqlite3_vfs*, void*, const char *zSymbol))(void);
```

This pattern appears in SQLite's `sqlite3_vfs` struct for the dynamic library loading interface.

## Solution

Added `parseNestedFunctionPointerField()` in parser.go to handle the nested case. When `parseFunctionPointerField()` sees `(*` followed by another `(*`, it delegates to the nested handler.

Key changes:
1. Extracted `parseFuncPtrParamTypes()` to share parameter parsing logic
2. `parseNestedFunctionPointerField()` parses: `returnType (*(*name)(innerParams))(outerParams)`
3. Builds type string: `returnType(*)(outerParams)(*)(innerParams)`

## Test

Added test case "function pointer returning function pointer" in `TestFunctionPointerInStructField`:
```c
struct S { void (*(*xDlSym)(void*, const char*))(void); };
```

## Status

✅ COMPLETE
- Added test case
- Implemented fix  
- All tests pass (`make check`)
