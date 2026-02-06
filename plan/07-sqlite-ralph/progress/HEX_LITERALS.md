# Hex Literals Support

## Task
Fix lexer to handle hex literals (`0x...`, `0X...`) properly.

## Problem Found
The `readNumber()` function in `pkg/lexer/lexer.go` only handles decimal digits.
For input `0x09`, it reads `0` and stops. The `x09` becomes a separate identifier token.

## Test Case
```c
int x = 0x09;  // ERROR: expected ;, got IDENT
```

## Fix Applied
1. Updated `readNumber()` in lexer to detect and read hex (`0x`/`0X`) and octal (`0...`) prefixes
2. Added `isHexDigit()` and `isOctalDigit()` helper functions
3. Also handle integer suffixes (`u`, `U`, `l`, `L`, etc.)
4. Updated `parseIntegerLiteral()` in parser to use `strconv.ParseInt(lit, 0, 64)` which auto-detects base

## Tests Added
- `TestHexAndOctalLiterals` in `pkg/lexer/lexer_test.go`
- Tests hex, octal, decimal, and suffix combinations

## Verification
- `make test` passes
- SQLite parsing progresses past hex literals (no longer infinite loop on hex)
- New blocker: `__builtin_va_list` and other compiler builtins

## Status
- [x] Identify root cause
- [x] Implement fix
- [x] Add tests
- [x] Verify SQLite parses further (next blocker identified)
