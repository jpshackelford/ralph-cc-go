# Preprocessor Expression Parsing

## Status: COMPLETE

## Summary

Fixed `#if` expression parsing to handle macros containing `defined()` operator and added support for additional clang-specific operators.

## Changes Made

### 1. Post-expansion `defined` handling (conditional.go)

Added `processDefinedOperator()` function that processes `defined` operator AFTER macro expansion. This fixes cases where macros expand to expressions containing `defined()`.

Example that now works:
```c
#define __is_modern_darwin(ios, macos) \
    (__IPHONE_OS_VERSION_MIN_REQUIRED >= (ios) || \
     __MAC_OS_X_VERSION_MIN_REQUIRED >= (macos) || \
     defined(__DRIVERKIT_VERSION_MIN_REQUIRED))

#if __is_modern_darwin(70000, 1090)  // Now correctly evaluates
```

### 2. Added `__building_module` and `__has_c_attribute` operators (conditional.go)

These clang-specific operators appear in system headers:
- `__building_module(X)` - Always returns 0 (we don't build modules)
- `__has_c_attribute(X)` - Always returns 0 (C2x attribute check)

### 3. Added `__APPLE_CC__` built-in macro (macro.go)

System headers check `__GNUC__ && __APPLE_CC__` to identify Apple toolchain. Value set to 6000.

## Verification

```bash
make test   # All fast tests pass
```

Preprocessing of `sqlite3.c` now progresses beyond the initial errors until hitting a different issue (multiline macro arguments).

## Next Blocker

Preprocessing now fails at line 24109 with "unterminated macro argument list" - this is because macro invocations spanning multiple lines are not supported. This is a separate issue from `#if` expression parsing.
