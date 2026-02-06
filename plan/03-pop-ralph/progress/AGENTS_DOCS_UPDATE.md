# Progress: Update AGENTS.md and Supporting Docs

## Task
Based on `plan/04-learn/ANALYSIS.md`, update AGENTS.md and supporting docs.

## Key Findings from ANALYSIS.md

1. **Make check timeout problem**: `make check` takes 16+ minutes; agents should use `make test` for quick iteration
2. **Excessive retry loops**: Agents didn't understand compiler IR phases, causing 3-11x retries
3. **Missing debugging knowledge**: IR dump flags, debugging flowchart needed
4. **Known gotchas**: FP addressing, callee-saved registers, struct types

## Changes Made

### AGENTS.md Updates
Added new `ralph-cc` section with:
- Quick Build Commands (with warning about `make check` being slow)
- Debugging the Compiler section with IR dump flags table
- Debugging Flowchart (symptom → which IR to inspect)
- Known Gotchas section (FP addressing, callee-saved regs, struct types, frame size)
- Test Data Patterns listing example-c files
- Key Documentation links

### docs/ Assessment
- TESTING.md already has clear fast vs slow guidance
- PHASES.md already documents IR stages comprehensively

## Status
- [x] Create updated AGENTS.md
- [x] Verify make check passes
- [x] Commit changes
