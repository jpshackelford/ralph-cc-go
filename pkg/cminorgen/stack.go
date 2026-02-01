// Package cminorgen implements the Cminorgen pass: Csharpminor → Cminor
// This file handles stack frame layout computation.
package cminorgen

import "github.com/raymyers/ralph-cc/pkg/csharpminor"

// StackSlot represents a variable's location on the stack frame.
type StackSlot struct {
	Name      string // variable name
	Offset    int64  // offset from stack frame base
	Size      int64  // size in bytes
	Alignment int64  // alignment requirement
}

// StackLayout represents the complete stack frame for a function.
type StackLayout struct {
	Slots     []StackSlot // stack slots in order
	TotalSize int64       // total stack size (including padding)
}

// GetSlot returns the slot for a variable by name, or nil if not found.
func (l *StackLayout) GetSlot(name string) *StackSlot {
	for i := range l.Slots {
		if l.Slots[i].Name == name {
			return &l.Slots[i]
		}
	}
	return nil
}

// ComputeStackLayout computes the stack frame layout for a function's locals.
// Variables are allocated in order with proper alignment.
func ComputeStackLayout(locals []csharpminor.VarDecl) StackLayout {
	layout := StackLayout{}
	offset := int64(0)

	for _, local := range locals {
		alignment := alignmentForSize(local.Size)

		// Align offset to required alignment
		offset = alignUp(offset, alignment)

		layout.Slots = append(layout.Slots, StackSlot{
			Name:      local.Name,
			Offset:    offset,
			Size:      local.Size,
			Alignment: alignment,
		})

		offset += local.Size
	}

	// Total size aligned to 8 bytes (aarch64 stack alignment)
	layout.TotalSize = alignUp(offset, 8)

	return layout
}

// alignmentForSize returns the natural alignment for a given size.
// On aarch64: 1-byte types have 1-byte alignment, 2-byte types have 2-byte,
// 4-byte have 4-byte, and 8-byte (and larger) have 8-byte alignment.
func alignmentForSize(size int64) int64 {
	switch {
	case size <= 1:
		return 1
	case size <= 2:
		return 2
	case size <= 4:
		return 4
	default:
		return 8
	}
}

// alignUp rounds n up to the nearest multiple of align.
func alignUp(n, align int64) int64 {
	if align <= 0 {
		return n
	}
	return (n + align - 1) / align * align
}

// IsStackAllocated returns true if the variable needs stack allocation.
// Currently all Csharpminor locals are treated as stack-allocated.
// In a more complete implementation, we would analyze address-taken variables.
func IsStackAllocated(name string, locals []csharpminor.VarDecl) bool {
	for _, l := range locals {
		if l.Name == name {
			return true
		}
	}
	return false
}

// FindAddressTaken scans a statement for address-of operations on locals.
// Returns a set of variable names that have their address taken.
// This is a placeholder for more sophisticated analysis.
func FindAddressTaken(stmt csharpminor.Stmt, locals []csharpminor.VarDecl) map[string]bool {
	result := make(map[string]bool)
	localSet := make(map[string]bool)
	for _, l := range locals {
		localSet[l.Name] = true
	}

	findAddressTakenInStmt(stmt, localSet, result)
	return result
}

func findAddressTakenInStmt(stmt csharpminor.Stmt, locals, result map[string]bool) {
	switch s := stmt.(type) {
	case csharpminor.Sskip:
		// nothing
	case csharpminor.Sset:
		findAddressTakenInExpr(s.RHS, locals, result)
	case csharpminor.Sstore:
		findAddressTakenInExpr(s.Addr, locals, result)
		findAddressTakenInExpr(s.Value, locals, result)
	case csharpminor.Scall:
		findAddressTakenInExpr(s.Func, locals, result)
		for _, arg := range s.Args {
			findAddressTakenInExpr(arg, locals, result)
		}
	case csharpminor.Stailcall:
		findAddressTakenInExpr(s.Func, locals, result)
		for _, arg := range s.Args {
			findAddressTakenInExpr(arg, locals, result)
		}
	case csharpminor.Sbuiltin:
		for _, arg := range s.Args {
			findAddressTakenInExpr(arg, locals, result)
		}
	case csharpminor.Sseq:
		findAddressTakenInStmt(s.First, locals, result)
		findAddressTakenInStmt(s.Second, locals, result)
	case csharpminor.Sifthenelse:
		findAddressTakenInExpr(s.Cond, locals, result)
		findAddressTakenInStmt(s.Then, locals, result)
		findAddressTakenInStmt(s.Else, locals, result)
	case csharpminor.Sloop:
		findAddressTakenInStmt(s.Body, locals, result)
	case csharpminor.Sblock:
		findAddressTakenInStmt(s.Body, locals, result)
	case csharpminor.Sexit:
		// nothing
	case csharpminor.Sswitch:
		findAddressTakenInExpr(s.Expr, locals, result)
		for _, c := range s.Cases {
			findAddressTakenInStmt(c.Body, locals, result)
		}
		findAddressTakenInStmt(s.Default, locals, result)
	case csharpminor.Sreturn:
		if s.Value != nil {
			findAddressTakenInExpr(s.Value, locals, result)
		}
	case csharpminor.Slabel:
		findAddressTakenInStmt(s.Body, locals, result)
	case csharpminor.Sgoto:
		// nothing
	}
}

func findAddressTakenInExpr(expr csharpminor.Expr, locals, result map[string]bool) {
	switch e := expr.(type) {
	case csharpminor.Evar:
		// Global variable reference - not address-taken for locals
	case csharpminor.Etempvar:
		// Temp variable - not address-taken
	case csharpminor.Eaddrof:
		// This is address-of for globals in Csharpminor
		// If it were a local, it would be address-taken
		if locals[e.Name] {
			result[e.Name] = true
		}
	case csharpminor.Econst:
		// nothing
	case csharpminor.Eunop:
		findAddressTakenInExpr(e.Arg, locals, result)
	case csharpminor.Ebinop:
		findAddressTakenInExpr(e.Left, locals, result)
		findAddressTakenInExpr(e.Right, locals, result)
	case csharpminor.Ecmp:
		findAddressTakenInExpr(e.Left, locals, result)
		findAddressTakenInExpr(e.Right, locals, result)
	case csharpminor.Eload:
		findAddressTakenInExpr(e.Addr, locals, result)
	}
}
