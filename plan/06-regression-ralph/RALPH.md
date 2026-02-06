Execute these steps.

1. Run csmith enough to find at least 1 bug in this compiler. Use `scripts/csmith-fuzz.sh` or similar if usefull.
2. Create `plan/06-regression-ralph/progress/...` markdown named after test case id, you'll document findings there
3. Any pending git changes are from a previous attempt. your choice to finish or reset.
4. Do ONLY that fix, and related automated tests.
5. Verify (including `make check`).
6. Add this case to the fuzz file, `plan/06-regression-ralph/scripts/regression.sh` and run it to make sure nothing broke.
7. Update progress file.
8. If everything passed, commit. Else bail.


## Progress files

The audience will be a coding agent that needs to continue your task with little help, or understand the history of the execution. Err towards terse mention of past steps (unless something went wrong), more detail on current state.

## Tech Guidelines

We have a prototype C compiler, which we are trying to get working on real programs.

Our CLI is in Go lang, but following the compcert design with goal of equivalent output on each IR. Optimizations are not required (compare with -O0).

Makefile has test, lint and check (doing both).

For tests prefer data-driven from cases in in `testdata/*.yaml` listing input/output for examples. Also for e2e we can some full programs in `testdata/example-c.*.c`.

Docs in `docs/` have useful information, updated when needed.
