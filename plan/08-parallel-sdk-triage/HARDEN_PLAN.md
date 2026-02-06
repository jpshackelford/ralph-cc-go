# Super Ralph: Roadmap to SQLite

## Mission

Take ralph-cc from ~95% csmith pass rate (constrained C subset) to compiling and running SQLite. Each phase builds on the last. Progress is tracked in STATE.md (sibling file, auto-created by RALPH.md).

---

## Phase 0: STABILIZE

**Entry**: regression.sh has any failures, or csmith base pass rate < 100%
**Work**: Fix csmith bugs one at a time using GCC oracle comparison
**Exit**: regression.sh all pass AND csmith-fuzz.sh 200 iterations at 100% pass rate (current constraints)

Key techniques:
- Use `--drtl`/`--dltl`/`--dmach`/`--dasm` to trace bugs through IR
- Consult plan/05-fix-research-ralph/COMMON_CAUSES.md for known patterns
- One fix per invocation, with regression test

---

## Phase 1: EXPAND CSMITH

**Entry**: Phase 0 exit criteria met
**Work**: Enable csmith features one at a time, fix what breaks

Expansion order (tracked by `expand_index` in STATE.md):

| Index | Remove Flag | Feature |
|-------|------------|---------|
| 0 | `--no-jumps` | goto/labels |
| 1 | `--no-safe-math` | overflow math |
| 2 | `--no-pointers` | pointer ops |
| 3 | `--no-arrays` | array indexing |
| 4 | `--no-structs` | struct types |
| 5 | `--no-unions` | union types |
| 6 | `--no-longlong` | 64-bit int |
| 7 | `--no-bitfields` | bitfield members |
| 8 | `--no-volatiles` | volatile qualifier |

For each feature:
- Run 100 csmith iterations with that flag removed
- Fix bugs until >=99% pass rate
- Lock in (update csmith-fuzz.sh and regression.sh to remove that flag)
- Run full regression to confirm no regressions

**Exit**: All 9 flags removed, csmith 200 iterations >= 98% pass rate with full feature set

---

## Phase 2: HARDEN (Feature Gaps)

**Entry**: Phase 1 exit criteria met
**Work**: Fix features csmith doesn't exercise

Feature checklist (each item should be one-invocation sized):
- [ ] Hex literals (0xFF, 0x1234) in lexer
- [ ] Octal literals (077) in lexer
- [ ] String literal assignment to char* variable
- [ ] String literal passed to function parameter
- [ ] Function pointer declaration and call through pointer
- [ ] Struct initializer (`struct s x = {1, 2};`)
- [ ] Array initializer (`int a[] = {1, 2, 3};`)
- [ ] Designated initializers (`.field = val`, `[i] = val`)
- [ ] Multi-dimensional arrays (`int a[2][3]`, indexing)
- [ ] Nested structs (struct with struct member)
- [ ] Nested unions (union inside struct)
- [ ] Enum with explicit values (`enum { A=5, B=10 }`)
- [ ] Float/double local variables and assignment
- [ ] Float arithmetic (+, -, *, /)
- [ ] Float-to-int and int-to-float casts
- [ ] Float function parameters and return values
- [ ] do { } while(0) macro pattern
- [ ] Comma operator in for loop increment
- [ ] Typedef of typedef (`typedef int myint; typedef myint myint2;`)
- [ ] Typedef of pointer-to-function
- [ ] Static local variables (persist across calls)
- [ ] External linkage (extern declaration matching global in another TU)

For each:
- Write e2e_runtime.yaml test case
- Compile with ralph-cc, compare against gcc
- Fix what fails

**Exit**: All checklist items passing in e2e_runtime.yaml

---

## Phase 3: REAL PROGRAMS

**Entry**: Phase 2 exit criteria met
**Work**: Compile progressively larger real C programs

Target progression (each includes acquisition instructions):

1. [ ] **Handwritten integration test** (~50 lines)
   Create `testdata/real/integration_basic.c` — a single file using structs, pointers, arrays, function pointers, enums, string ops, and control flow together. Write it yourself to exercise everything from Phase 2.

2. [ ] **jsmn** — minimal JSON parser (~500 lines, single header)
   ```bash
   mkdir -p testdata/real
   curl -L 'https://raw.githubusercontent.com/zserge/jsmn/master/jsmn.h' -o testdata/real/jsmn.h
   ```
   Write a small test driver `testdata/real/jsmn_test.c` that parses `{"key":"value"}`.

3. [ ] **miniz** — single-file zlib replacement (~4k lines of C)
   ```bash
   curl -L 'https://raw.githubusercontent.com/richgel999/miniz/master/miniz.c' -o testdata/real/miniz.c
   curl -L 'https://raw.githubusercontent.com/richgel999/miniz/master/miniz.h' -o testdata/real/miniz.h
   ```
   Write a test driver that compresses and decompresses a small string.

4. [ ] **sbase/cat.c** — simple Unix cat utility from suckless (~100 lines with deps)
   Alternatively, write a small `testdata/real/minicat.c` that reads stdin and writes stdout, exercising FILE* and character I/O.

5. [ ] **Lua 5.4 core** (~20k lines)
   ```bash
   curl -L 'https://www.lua.org/ftp/lua-5.4.7.tar.gz' -o /tmp/lua.tar.gz
   tar xzf /tmp/lua.tar.gz -C testdata/real/
   ```
   Preprocess, compile with ralph-cc, run `print(40+2)`.

Method for all: `gcc -E -w -P program.c -o program_pp.c && ./bin/ralph-cc -dasm program_pp.c`
Compare output against gcc-compiled binary.

**Exit**: Can compile and correctly run at least jsmn_test and one program >= 1k lines

---

## Phase 4: SQLITE

**Entry**: Phase 3 exit criteria met
**Work**: Compile SQLite amalgamation

Acquisition:
```bash
mkdir -p testdata/sqlite
curl -L 'https://www.sqlite.org/2024/sqlite-amalgamation-3460100.zip' -o /tmp/sqlite.zip
unzip -o /tmp/sqlite.zip -d /tmp/sqlite_src
cp /tmp/sqlite_src/sqlite-amalgamation-*/sqlite3.c testdata/sqlite/
cp /tmp/sqlite_src/sqlite-amalgamation-*/sqlite3.h testdata/sqlite/
cp /tmp/sqlite_src/sqlite-amalgamation-*/shell.c testdata/sqlite/
```

Steps:
1. [ ] Preprocess: `gcc -E -P -DSQLITE_THREADSAFE=0 -DSQLITE_OMIT_LOAD_EXTENSION -DSQLITE_OMIT_WAL sqlite3.c -o sqlite3_pp.c`
2. [ ] Compile with ralph-cc (fix errors iteratively, one per invocation)
3. [ ] Assemble + link
4. [ ] Run: `echo "CREATE TABLE t(x); INSERT INTO t VALUES(42); SELECT * FROM t;" | ./sqlite3_ralph`
5. [ ] Fix runtime issues until queries return correct results

**Exit**: SQLite shell can create table, insert rows, and query them correctly

---

## Estimated Scale

| Phase | Invocations | Cumulative |
|-------|------------|------------|
| 0: STABILIZE | 5-15 | 15 |
| 1: EXPAND | 30-80 | 95 |
| 2: HARDEN | 20-40 | 135 |
| 3: REAL PROGRAMS | 30-100 | 235 |
| 4: SQLITE | 50-200 | 435 |

Some bugs are one-liners, some require architectural changes.

---

## Invariants (Never Break These)

1. `make check` must pass after every commit
2. `regression.sh` must pass after every commit (never regress)
3. Every fix must have an automated test (e2e_runtime.yaml or regression seed)
4. GCC is the oracle — if gcc and ralph-cc disagree, ralph-cc is wrong
5. One logical fix per invocation — keep diffs small and reviewable
