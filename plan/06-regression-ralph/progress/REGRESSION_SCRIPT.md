# Progress: Regression Script for Csmith Findings

## Task
Create `plan/06-regression-ralph/scripts/regression.sh` that quickly runs csmith to verify all existing `csmith-reports` findings stay fixed (add example seeds to script, don't rely on the folder).

## Approach
1. Extract all unique seeds from existing csmith-reports crash/mismatch files
2. Create a regression script that regenerates test cases from seeds
3. Compile with ralph-cc and compare against gcc
4. Report pass/fail for each known issue

## Seeds Embedded (18 total)
All seeds are embedded in the script itself, not relying on csmith-reports folder:
- 15 crash seeds
- 1 fail_compile seed
- 2 mismatch seeds

## Result
Created `plan/06-regression-ralph/scripts/regression.sh`:
- Uses same test generation logic as csmith-fuzz.sh
- Embeds all 18 known issue seeds directly
- Regenerates tests from seeds (deterministic via --seed)
- Compares ralph-cc output against gcc
- Reports FIXED/FAIL/SKIP for each issue
- Returns exit 0 if all pass, exit 1 if any regressions

Test run results:
- 14 FIXED (issues that are now resolved)
- 4 still failing (existing bugs not yet fixed)
- `make check` passes

## Status
COMPLETE
