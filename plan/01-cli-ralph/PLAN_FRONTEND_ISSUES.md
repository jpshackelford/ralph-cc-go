# Frontend Issues

Issues discovered during frontend review.

## Issue 1: SimplLocals address-taken analysis doesn't handle `*cabs.Block`

**Severity:** Critical  
**Status:** ✅ FIXED

**Symptom:** Variables whose address is taken (e.g., `&x`) are incorrectly promoted to temporaries, causing panic in Csharpminor generation: "cannot take address of expression"

**Root Cause:**  
In `pkg/simpllocals/transform.go`, `AnalyzeStmt()` has a case for `cabs.Block` (value type), but `FunDef.Body` is `*cabs.Block` (pointer type). The switch case doesn't match, so the function body is never analyzed for address-taken variables.

**Location:** `pkg/simpllocals/transform.go:157`

**Fix Applied:**  
Added a case for `*cabs.Block` in `AnalyzeStmt`:
```go
case *cabs.Block:
    for _, item := range stmt.Items {
        t.AnalyzeStmt(item)
    }
```

**Tests Added:**
- `TestAnalyzeFunctionWithPointerBlock` - verifies function body (*Block) is analyzed
- `TestAnalyzeStmtBlockPointer` - direct test for *cabs.Block case

---

## Issue 2: Clight printer shows wrong variable name in address-of

**Severity:** Medium (cosmetic but confusing)  
**Status:** ✅ FIXED (by Issue 1 fix)

**Symptom:** After SimplLocals, the Clight output shows `$1 = &$1;` instead of `$1 = &x;`

**Root Cause:**  
Related to Issue 1. With Issue 1 fixed, `x` now correctly remains as `Evar{Name: "x"}` and is not promoted, making the output correct.

---

## Issue 3: all.c crashes on `-dcsharpminor` and `-dcminor`

**Severity:** Critical  
**Status:** ✅ FIXED (by Issue 1 fix)

**Symptom:** Running `bin/ralph-cc testdata/example-c/all.c -dcsharpminor` panics

**Root Cause:**  
Same as Issue 1. The `pointerOps` function takes `&x`.

**Fix:** Same as Issue 1.

---

## Notes

- CminorSel is documented as an internal phase with no CompCert dump flag, so no `-dcminorsel` CLI flag is needed
- The `pkg/selection` package is complete but wired only via `pkg/cminorsel` (internal use)
