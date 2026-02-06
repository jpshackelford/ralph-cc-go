# Cast Pointer Types

## Task
Fix cast expression parsing to support pointer types like `(char*)`, `(void*)`, `(const char*)`.

## Problem
Parser's `parseCast()` only consumed single-token type names:
```go
typeName := p.curToken.Literal
p.nextToken() // consume type name
if !p.curTokenIs(lexer.TokenRParen) {
    p.addError(...)  // Fails here for (char*) because next is * not )
```

This fails on:
- `(char*)ptr` - sees `*` instead of `)`
- `(void*)0` - same issue
- `(const char*)str` - const qualifier not consumed

## Fix
Update `parseCast()` to:
1. Parse qualifiers (const, volatile)
2. Parse compound type specifier (handles struct, multi-word like `unsigned int`)
3. Parse pointer markers (`*`)
4. Build complete type string

## Current Status
- [ ] Analyzing cast parsing code
- [ ] Implementing fix
- [ ] Testing with sqlite3.c
