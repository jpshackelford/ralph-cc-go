# Testing Strategy

This document describes ralph-cc's testing approach across multiple levels.

## Quick Reference

```bash
make test       # Fast tests (~2s), skips slow runtime tests
make test-slow  # Runtime tests only (~30s), requires as/ld
make test-all   # All tests
make check      # lint + test-all
make coverage   # Generate coverage report
```

## Test Categories

### 1. Unit Tests

Standard Go tests in each `pkg/*/` package. These run in-memory and test internal logic.

**Pattern**: Table-driven tests with structs defining inputs and expected outputs.

```go
// Example from pkg/lexer/lexer_test.go
tests := []struct {
    expectedType    TokenType
    expectedLiteral string
}{
    {TokenInt_, "int"},
    {TokenIdent, "main"},
    // ...
}
```

**Locations**: `pkg/*/*.go` (64 test files total)

### 2. Data-Driven YAML Tests

Test cases defined in YAML files in `testdata/`. Go test code loads and iterates over cases.

#### `testdata/parse.yaml`
Parser AST verification. Each test provides C input and expected AST structure.

```yaml
tests:
  - name: "return zero"
    input: |
      int f() { return 0; }
    ast:
      kind: FunDef
      name: f
      body:
        kind: Block
        items:
          - kind: Return
            expr:
              kind: Constant
              value: 0
```

#### `testdata/integration.yaml`
Tests comparing ralph-cc `-dparse` output against CompCert `ccomp -dparse`. Requires CompCert to be built.

```yaml
tests:
  - name: "empty main"
    input: |
      int main() {}
  - name: "return constant"
    input: |
      int f() { return 42; }
```

#### `testdata/e2e_asm.yaml`
End-to-end C to ARM64 assembly tests. Verifies specific strings appear in output.

```yaml
tests:
  - name: "hello world - return zero"
    input: |
      int main() { return 0; }
    expect:
      - ".text"
      - ".global\tmain"
      - "main:"
      - "ret"
    expect_order:          # Optional: must appear in this order
      - "mov\tw0, #42"
      - "ret"
    expect_unique:         # Optional: must appear exactly once
      - "main:"
    expect_not:            # Optional: must NOT appear
      - ".L1:"
```

#### `testdata/e2e_runtime.yaml`
Full pipeline tests: compile → assemble → link → run → check exit code.

```yaml
tests:
  - name: "C1.1 - integer constant 42"
    input: |
      int main() { return 42; }
    expected_exit: 42
```

**Note**: These tests are slow and require `as` and `ld` in PATH. They are skipped by `make test` and run by `make test-slow`.

### 3. Example C Files

`testdata/example-c/` contains sample C files with their expected IR outputs at various stages:

- `fib.c` → `fib.parsed.c`, `fib.light.c`, `fib.rtl.0`, `fib.ltl`, `fib.mach`, `fib.s`
- `hello.c` → `hello.cminor`, `hello.csharpminor`, `hello.mach`, `hello.s`

These serve as both documentation and regression baselines.

## Test Organization

### Fast vs Slow

The Makefile splits tests by speed:

```makefile
test:
    go test -skip 'TestE2ERuntimeYAML' ./...

test-slow:
    go test -run 'TestE2ERuntimeYAML' ./...
```

Fast tests (~2s) run the full pipeline except runtime execution. Slow tests (~30s) compile, assemble, link, and run actual binaries.

### Test File Structure

```
ralph-cc/
├── cmd/ralph-cc/
│   ├── main_test.go           # CLI flag tests
│   └── integration_test.go    # E2E and CompCert comparison
├── pkg/
│   ├── lexer/lexer_test.go    # Token scanning
│   ├── parser/parser_test.go  # AST construction
│   ├── regalloc/irc_test.go   # Register allocation
│   └── ...                    # (64 test files total)
└── testdata/
    ├── parse.yaml             # Parser test data
    ├── integration.yaml       # CompCert equivalence data
    ├── e2e_asm.yaml           # Assembly output tests
    ├── e2e_runtime.yaml       # Runtime execution tests
    └── example-c/             # Sample C programs
```

## Adding Tests

### Adding a Unit Test

Add table entries or new test functions to the appropriate `*_test.go` file:

```go
func TestNewFeature(t *testing.T) {
    // setup
    result := myFunction(input)
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

### Adding a YAML Test Case

For parser tests, add to `testdata/parse.yaml`:

```yaml
  - name: "my new feature"
    input: |
      int f() { /* new syntax */ }
    ast:
      kind: FunDef
      # ... expected AST
```

For runtime tests, add to `testdata/e2e_runtime.yaml`:

```yaml
  - name: "my feature - description"
    input: |
      int main() { /* code */ return result; }
    expected_exit: 42
```

### Skipping Tests

YAML tests support a `skip` field:

```yaml
  - name: "known broken feature"
    input: ...
    skip: "Not yet implemented"
```

## Coverage

```bash
make coverage
# Creates coverage/coverage.html
```

View the HTML report to identify untested code paths.

## Critique & Opportunities

### Strengths

1. **Data-driven tests**: YAML test data is readable, easy to add cases, separates test logic from test data.
2. **CompCert comparison**: `integration.yaml` provides ground truth for parser output.
3. **Fast/slow split**: Developers get quick feedback; CI can run full suite.
4. **Hierarchical coverage**: Unit tests for components, E2E tests for the pipeline.

### Weaknesses

1. **No fuzz testing**: Parser and lexer could benefit from fuzz testing to find edge cases.
2. **Limited error path coverage**: Few tests verify error messages or handling of invalid input.
3. **Manual golden files**: `testdata/example-c/*.s` files are manually maintained—could drift.
4. **No mutation testing**: Unknown how robust tests are at catching bugs.
5. **Platform-specific**: E2E runtime tests assume macOS ARM64 (`xcrun`, `as`, `ld`). Linux/x86 untested.

### Opportunities

1. **Add `go test -fuzz`**: Fuzz the lexer and parser with random inputs.
2. **Snapshot testing**: Auto-update golden files with `UPDATE_SNAPSHOTS=1 make test`.
3. **Error case YAML files**: Add `testdata/errors.yaml` with expected parse/compile errors.
4. **CI matrix**: Test on Linux ARM64, Linux x86_64, macOS ARM64.
5. **Benchmark tests**: Add `BenchmarkParse`, `BenchmarkCompile` for regression tracking.
6. **Property-based testing**: Use `testing/quick` for invariant checks (e.g., roundtrip printing).
7. **Test coverage gates**: Fail CI if coverage drops below threshold.
