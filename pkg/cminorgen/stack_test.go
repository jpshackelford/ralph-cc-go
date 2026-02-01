package cminorgen

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/csharpminor"
)

func TestAlignUp(t *testing.T) {
	tests := []struct {
		n, align, want int64
	}{
		{0, 4, 0},
		{1, 4, 4},
		{3, 4, 4},
		{4, 4, 4},
		{5, 4, 8},
		{7, 8, 8},
		{8, 8, 8},
		{9, 8, 16},
		{0, 1, 0},
		{5, 1, 5},
	}
	for _, tt := range tests {
		got := alignUp(tt.n, tt.align)
		if got != tt.want {
			t.Errorf("alignUp(%d, %d) = %d, want %d", tt.n, tt.align, got, tt.want)
		}
	}
}

func TestAlignmentForSize(t *testing.T) {
	tests := []struct {
		size, want int64
	}{
		{1, 1},  // char
		{2, 2},  // short
		{4, 4},  // int
		{8, 8},  // long, pointer
		{16, 8}, // larger types use 8-byte alignment
		{3, 4},  // odd 3-byte -> 4
	}
	for _, tt := range tests {
		got := alignmentForSize(tt.size)
		if got != tt.want {
			t.Errorf("alignmentForSize(%d) = %d, want %d", tt.size, got, tt.want)
		}
	}
}

func TestComputeStackLayoutEmpty(t *testing.T) {
	layout := ComputeStackLayout(nil)
	if len(layout.Slots) != 0 {
		t.Errorf("empty layout should have 0 slots, got %d", len(layout.Slots))
	}
	if layout.TotalSize != 0 {
		t.Errorf("empty layout should have size 0, got %d", layout.TotalSize)
	}
}

func TestComputeStackLayoutSingleInt(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "x", Size: 4},
	}
	layout := ComputeStackLayout(locals)

	if len(layout.Slots) != 1 {
		t.Fatalf("expected 1 slot, got %d", len(layout.Slots))
	}

	slot := layout.Slots[0]
	if slot.Name != "x" {
		t.Errorf("slot name = %q, want %q", slot.Name, "x")
	}
	if slot.Offset != 0 {
		t.Errorf("slot offset = %d, want 0", slot.Offset)
	}
	if slot.Size != 4 {
		t.Errorf("slot size = %d, want 4", slot.Size)
	}
	if slot.Alignment != 4 {
		t.Errorf("slot alignment = %d, want 4", slot.Alignment)
	}

	// Total size should be 8 (aligned to 8)
	if layout.TotalSize != 8 {
		t.Errorf("total size = %d, want 8", layout.TotalSize)
	}
}

func TestComputeStackLayoutMultipleVars(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "a", Size: 1}, // char at offset 0
		{Name: "b", Size: 4}, // int at offset 4 (aligned)
		{Name: "c", Size: 8}, // long at offset 8 (aligned)
	}
	layout := ComputeStackLayout(locals)

	if len(layout.Slots) != 3 {
		t.Fatalf("expected 3 slots, got %d", len(layout.Slots))
	}

	// Check a: offset 0, size 1
	if layout.Slots[0].Offset != 0 {
		t.Errorf("a offset = %d, want 0", layout.Slots[0].Offset)
	}

	// Check b: offset 4 (aligned to 4), size 4
	if layout.Slots[1].Offset != 4 {
		t.Errorf("b offset = %d, want 4", layout.Slots[1].Offset)
	}

	// Check c: offset 8 (aligned to 8), size 8
	if layout.Slots[2].Offset != 8 {
		t.Errorf("c offset = %d, want 8", layout.Slots[2].Offset)
	}

	// Total: 8 + 8 = 16
	if layout.TotalSize != 16 {
		t.Errorf("total size = %d, want 16", layout.TotalSize)
	}
}

func TestComputeStackLayoutAlignmentPadding(t *testing.T) {
	// Two ints followed by a long requires padding
	locals := []csharpminor.VarDecl{
		{Name: "i1", Size: 4}, // offset 0
		{Name: "i2", Size: 4}, // offset 4
		{Name: "l", Size: 8},  // offset 8 (already aligned)
	}
	layout := ComputeStackLayout(locals)

	if layout.Slots[0].Offset != 0 {
		t.Errorf("i1 offset = %d, want 0", layout.Slots[0].Offset)
	}
	if layout.Slots[1].Offset != 4 {
		t.Errorf("i2 offset = %d, want 4", layout.Slots[1].Offset)
	}
	if layout.Slots[2].Offset != 8 {
		t.Errorf("l offset = %d, want 8", layout.Slots[2].Offset)
	}
	if layout.TotalSize != 16 {
		t.Errorf("total size = %d, want 16", layout.TotalSize)
	}
}

func TestComputeStackLayoutCharThenLong(t *testing.T) {
	// Char followed by long needs 7 bytes padding
	locals := []csharpminor.VarDecl{
		{Name: "c", Size: 1}, // offset 0
		{Name: "l", Size: 8}, // offset 8 (aligned to 8)
	}
	layout := ComputeStackLayout(locals)

	if layout.Slots[0].Offset != 0 {
		t.Errorf("c offset = %d, want 0", layout.Slots[0].Offset)
	}
	if layout.Slots[1].Offset != 8 {
		t.Errorf("l offset = %d, want 8", layout.Slots[1].Offset)
	}
	// Total: 8 + 8 = 16
	if layout.TotalSize != 16 {
		t.Errorf("total size = %d, want 16", layout.TotalSize)
	}
}

func TestComputeStackLayoutArray(t *testing.T) {
	// Array of 10 ints (40 bytes)
	locals := []csharpminor.VarDecl{
		{Name: "arr", Size: 40}, // 10 ints
	}
	layout := ComputeStackLayout(locals)

	if layout.Slots[0].Offset != 0 {
		t.Errorf("arr offset = %d, want 0", layout.Slots[0].Offset)
	}
	if layout.Slots[0].Size != 40 {
		t.Errorf("arr size = %d, want 40", layout.Slots[0].Size)
	}
	// 40 aligned to 8 = 40
	if layout.TotalSize != 40 {
		t.Errorf("total size = %d, want 40", layout.TotalSize)
	}
}

func TestGetSlot(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "a", Size: 4},
		{Name: "b", Size: 8},
	}
	layout := ComputeStackLayout(locals)

	// Found
	slot := layout.GetSlot("b")
	if slot == nil {
		t.Fatal("GetSlot(b) returned nil")
	}
	if slot.Name != "b" {
		t.Errorf("slot name = %q, want %q", slot.Name, "b")
	}

	// Not found
	slot = layout.GetSlot("notexist")
	if slot != nil {
		t.Errorf("GetSlot(notexist) should return nil")
	}
}

func TestIsStackAllocated(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "x", Size: 4},
		{Name: "y", Size: 8},
	}

	if !IsStackAllocated("x", locals) {
		t.Error("x should be stack allocated")
	}
	if !IsStackAllocated("y", locals) {
		t.Error("y should be stack allocated")
	}
	if IsStackAllocated("z", locals) {
		t.Error("z should not be stack allocated")
	}
}

func TestFindAddressTakenEmpty(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "x", Size: 4},
	}
	result := FindAddressTaken(csharpminor.Sskip{}, locals)
	if len(result) != 0 {
		t.Errorf("expected 0 address-taken, got %d", len(result))
	}
}

func TestFindAddressTakenAddrof(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "x", Size: 4},
		{Name: "y", Size: 4},
	}

	// Statement that takes address of x (via Eaddrof which is for locals marked as stack)
	stmt := csharpminor.Sset{
		TempID: 0,
		RHS:    csharpminor.Eaddrof{Name: "x"},
	}

	result := FindAddressTaken(stmt, locals)
	if !result["x"] {
		t.Error("x should be address-taken")
	}
	if result["y"] {
		t.Error("y should not be address-taken")
	}
}

func TestFindAddressTakenNested(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "a", Size: 4},
		{Name: "b", Size: 4},
	}

	// Nested in if-then-else
	stmt := csharpminor.Sifthenelse{
		Cond: csharpminor.Econst{Const: csharpminor.Ointconst{Value: 1}},
		Then: csharpminor.Sset{TempID: 0, RHS: csharpminor.Eaddrof{Name: "a"}},
		Else: csharpminor.Sset{TempID: 1, RHS: csharpminor.Eaddrof{Name: "b"}},
	}

	result := FindAddressTaken(stmt, locals)
	if !result["a"] {
		t.Error("a should be address-taken")
	}
	if !result["b"] {
		t.Error("b should be address-taken")
	}
}

func TestFindAddressTakenInSequence(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "x", Size: 4},
	}

	stmt := csharpminor.Sseq{
		First:  csharpminor.Sskip{},
		Second: csharpminor.Sset{TempID: 0, RHS: csharpminor.Eaddrof{Name: "x"}},
	}

	result := FindAddressTaken(stmt, locals)
	if !result["x"] {
		t.Error("x should be address-taken")
	}
}

func TestFindAddressTakenInLoop(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "x", Size: 4},
	}

	stmt := csharpminor.Sloop{
		Body: csharpminor.Sset{TempID: 0, RHS: csharpminor.Eaddrof{Name: "x"}},
	}

	result := FindAddressTaken(stmt, locals)
	if !result["x"] {
		t.Error("x should be address-taken")
	}
}

func TestFindAddressTakenInBlock(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "x", Size: 4},
	}

	stmt := csharpminor.Sblock{
		Body: csharpminor.Sset{TempID: 0, RHS: csharpminor.Eaddrof{Name: "x"}},
	}

	result := FindAddressTaken(stmt, locals)
	if !result["x"] {
		t.Error("x should be address-taken")
	}
}

func TestFindAddressTakenInSwitch(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "x", Size: 4},
	}

	stmt := csharpminor.Sswitch{
		Expr: csharpminor.Econst{Const: csharpminor.Ointconst{Value: 0}},
		Cases: []csharpminor.SwitchCase{
			{Value: 0, Body: csharpminor.Sset{TempID: 0, RHS: csharpminor.Eaddrof{Name: "x"}}},
		},
		Default: csharpminor.Sskip{},
	}

	result := FindAddressTaken(stmt, locals)
	if !result["x"] {
		t.Error("x should be address-taken")
	}
}

func TestFindAddressTakenInLabel(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "x", Size: 4},
	}

	stmt := csharpminor.Slabel{
		Label: "L1",
		Body:  csharpminor.Sset{TempID: 0, RHS: csharpminor.Eaddrof{Name: "x"}},
	}

	result := FindAddressTaken(stmt, locals)
	if !result["x"] {
		t.Error("x should be address-taken")
	}
}

func TestFindAddressTakenGlobalNotCounted(t *testing.T) {
	locals := []csharpminor.VarDecl{
		{Name: "x", Size: 4},
	}

	// Address of global (not in locals list)
	stmt := csharpminor.Sset{
		TempID: 0,
		RHS:    csharpminor.Eaddrof{Name: "global_var"},
	}

	result := FindAddressTaken(stmt, locals)
	if len(result) != 0 {
		t.Errorf("no locals should be address-taken, got %v", result)
	}
}
