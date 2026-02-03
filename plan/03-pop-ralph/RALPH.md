Execute these steps.

1. From the unfinished tasks in `plan/03-pop-ralph/PLAN.md`, choose a logical one to do next.
2. If a progress file is specified in the task, study it. Otherwise create one in `plan/03-pop-ralph/progress/` and reference in the plan file task.
2. Do ONLY that task, and related automated tests.
3. Verify (including `make check`).
4. If complete: update PLAN to mark complete, commit.
5. If not complete: update progress file and bail.

## Environments

If you need to install things, use `container-use` MCP or docker.

## Progress files

Choose unique short descriptive names 2-4 words formatted `LIKE_THIS.md`. If you need to avoid collision, numbers are fine.

The audience will be a coding agent that needs to continue your task with little help, or understand the history of the execution. Err towards terse mention of past steps (unless something went wrong), more detail on current state.

## Tech Guidelines

We have a prototype C compiler, which we are trying to get working on real programs.

Our CLI is in Go lang, but following the compcert design with goal of equivalent output on each IR. Optimizations are not required (compare with -O0).

Makefile has test, lint and check (doing both).

For tests prefer data-driven from cases in in `testdata/*.yaml` listing input/output for examples. Also for e2e we can some full programs in `testdata/example-c.*.c`.

Docs in `docs/` have useful information, updated when needed.
