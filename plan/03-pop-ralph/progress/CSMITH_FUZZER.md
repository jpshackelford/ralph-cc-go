# Csmith Fuzzer Progress

## Goal
Set up automated fuzzing with csmith to find bugs in ralph-cc by comparing output against gcc.

## Status: IN PROGRESS

## Approach
1. Generate simple C programs using csmith (no stdio/complex features ralph-cc doesn't support)
2. Compile with both gcc and ralph-cc
3. Compare exit codes
4. Report mismatches with seed for reproduction

## Constraints
- ralph-cc doesn't support: printf, stdio, complex types, pointers (limited), arrays (limited)
- csmith uses its own header with CRC checksums - need workaround
- Simple approach: use exit code as return value for computation

## Implementation
- Created `scripts/csmith-fuzz.sh` - headless fuzzing script
- Generates report in `csmith-reports/`

## Current state
- Building the fuzzing script
