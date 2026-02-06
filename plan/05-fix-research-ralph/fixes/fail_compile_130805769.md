# Fix: fail_compile_130805769 - Stack Slot Spill Handling Incomplete

## Summary

Compilation fails with panic when register allocator spills registers and the spilled locations appear in instructions other than `Lop`.

## Error Message

```
panic: stack slot in register position - regalloc incomplete
```

## Test Case

`csmith-reports/fail_compile_130805769.c` - A complex function with ~200 RTL nodes and high register pressure that causes spills.

## Root Cause

The register allocator (`pkg/regalloc/irc.go`) correctly spills registers to stack slots when there aren't enough physical registers. Spilled registers are assigned `ltl.S{Slot: ltl.SlotLocal, ...}` locations.

However, the stacking transform (`pkg/stacking/transform.go`) only handles stack slot spilling for `Lop` instructions (via `transformLop`). Other instruction types call `locToReg()` which panics on stack slots:

**Instructions that DON'T handle stack slots:**
1. `Lload` - `Dest` can be a stack slot (load result goes to stack)
2. `Lstore` - `Src` can be a stack slot (value to store is on stack)  
3. `Lcond` - `Args` can have stack slots (conditional branch comparison)
4. `Ljumptable` - `Arg` can be a stack slot (switch index)
5. `Lbuiltin` - `Args`/`Dest` can have stack slots

**Example from failing code (LTL output):**
```
474: { Lstore(Mint32, Aindexed(0), [X0], S(Local, 0, Tlong)); Lbranch 473 }
```

The `Src` of the store is `S(Local, 0, Tlong)` (a spilled register), but `transformInst` for `Lstore` calls `locToReg(i.Src)` which panics.

## Why This Happens

Complex functions like `func_4` in this test case have many live variables across function calls (`func_27`). The register allocator:

1. Detects variables live across calls must go in callee-saved registers (X19-X28)
2. With only 10 callee-saved registers, some must spill when there are more than 10 live-across-call variables
3. Spilled variables get `S(Local, ...)` locations

## Fix Plan

Extend the stacking transform to handle stack slots in ALL instruction types, not just `Lop`. For each instruction type that accesses locations:

### Option A: Per-Instruction Spill Handling (Recommended)

Add spill/reload logic similar to `transformLop` for each instruction type:

```go
case linear.Lstore:
    var result []mach.Instruction
    args := make([]ltl.MReg, len(i.Args))
    for j, arg := range i.Args {
        args[j] = t.ensureInReg(arg, &result, j)
    }
    src := t.ensureInReg(i.Src, &result, len(i.Args))
    result = append(result, mach.Mstore{
        Chunk: i.Chunk,
        Addr:  i.Addr,
        Args:  args,
        Src:   src,
    })
    return result
```

Where `ensureInReg` loads a stack slot into a temp register if needed:

```go
func (t *transformer) ensureInReg(loc linear.Loc, result *[]mach.Instruction, tempIdx int) ltl.MReg {
    switch l := loc.(type) {
    case linear.R:
        return l.Reg
    case linear.S:
        // Load from stack slot into temp register
        tempReg := stackingTempRegs[tempIdx % len(stackingTempRegs)]
        *result = append(*result, t.slotTrans.TranslateGetstack(linear.Lgetstack{
            Slot: l.Slot,
            Ofs:  l.Ofs,
            Ty:   l.Ty,
            Dest: tempReg,
        }))
        return tempReg
    }
    panic("unknown location type")
}
```

### Option B: Pre-pass Spill Insertion

Add a pass before stacking that inserts explicit `Lgetstack`/`Lsetstack` around spilled uses, ensuring all operands in load/store/cond/etc. are registers.

## Files to Modify

1. `pkg/stacking/transform.go` - Add spill handling for all instruction types

## Verification

After fix:
1. `fail_compile_130805769.c` should compile and generate valid assembly
2. All other csmith tests should still pass
3. Run `make test` to verify no regressions

## Related Issues

This is likely the cause of multiple other csmith failures, since high register pressure leading to spills is common in csmith-generated code.

## Notes

The current code has a misleading panic message "regalloc incomplete". The regalloc IS complete - it correctly decided to spill. The stacking transform just doesn't handle the spilled locations yet.
