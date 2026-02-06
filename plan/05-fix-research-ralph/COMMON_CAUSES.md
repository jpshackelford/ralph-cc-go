# Common Causes of Compilation Bugs

## Stack Frame Layout Issues

### Callee-Save Register Offset Sign Error (CONFIRMED)

**Symptom**: Runtime crashes (SIGSEGV/SIGBUS) in functions using callee-saved registers (x19-x28).

**Cause**: Callee-saved registers stored at positive offsets from FP, which go outside the allocated stack frame into invalid memory.

**Location**: `pkg/stacking/layout.go` CalleeSaveOffset calculation

**Fix**: Use negative offsets from FP for callee-saved register storage.

---

## Categories to Watch

1. **Stack layout** - offset calculations, frame pointer usage
2. **Type coercions** - signed/unsigned, width conversions  
3. **Operator semantics** - division, shifts, overflow
4. **Control flow** - branch conditions, fall-through
5. **ABI compliance** - calling conventions, register usage