package stacking

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/linear"
	"github.com/raymyers/ralph-cc/pkg/ltl"
	"github.com/raymyers/ralph-cc/pkg/mach"
)

func TestTranslateGetstackLocal(t *testing.T) {
	fn := linear.NewFunction("test", linear.Sig{})
	fn.Append(linear.Lgetstack{
		Slot: linear.SlotLocal,
		Ofs:  0,
		Ty:   linear.Tlong,
		Dest: ltl.X0,
	})

	layout := ComputeLayout(fn, 0)
	trans := NewSlotTranslator(layout)

	inst := trans.TranslateGetstack(linear.Lgetstack{
		Slot: linear.SlotLocal,
		Ofs:  0,
		Ty:   linear.Tlong,
		Dest: ltl.X0,
	})

	getstack, ok := inst.(mach.Mgetstack)
	if !ok {
		t.Fatalf("expected Mgetstack, got %T", inst)
	}

	// Offset should be LocalOffset + 0
	if getstack.Ofs != layout.LocalOffset {
		t.Errorf("Ofs = %d, want %d", getstack.Ofs, layout.LocalOffset)
	}
	if getstack.Ty != linear.Tlong {
		t.Errorf("Ty = %v, want Tlong", getstack.Ty)
	}
	if getstack.Dest != ltl.X0 {
		t.Errorf("Dest = %v, want X0", getstack.Dest)
	}
}

func TestTranslateGetstackIncoming(t *testing.T) {
	fn := linear.NewFunction("test", linear.Sig{})
	layout := ComputeLayout(fn, 0)
	trans := NewSlotTranslator(layout)

	inst := trans.TranslateGetstack(linear.Lgetstack{
		Slot: linear.SlotIncoming,
		Ofs:  0,
		Ty:   linear.Tlong,
		Dest: ltl.X1,
	})

	getparam, ok := inst.(mach.Mgetparam)
	if !ok {
		t.Fatalf("expected Mgetparam, got %T", inst)
	}

	// Incoming at offset 0 should be at FP+16
	if getparam.Ofs != 16 {
		t.Errorf("Ofs = %d, want 16", getparam.Ofs)
	}
	if getparam.Ty != linear.Tlong {
		t.Errorf("Ty = %v, want Tlong", getparam.Ty)
	}
	if getparam.Dest != ltl.X1 {
		t.Errorf("Dest = %v, want X1", getparam.Dest)
	}
}

func TestTranslateGetstackOutgoing(t *testing.T) {
	fn := linear.NewFunction("test", linear.Sig{})
	layout := ComputeLayout(fn, 0)
	trans := NewSlotTranslator(layout)

	inst := trans.TranslateGetstack(linear.Lgetstack{
		Slot: linear.SlotOutgoing,
		Ofs:  8,
		Ty:   linear.Tint,
		Dest: ltl.X2,
	})

	getstack, ok := inst.(mach.Mgetstack)
	if !ok {
		t.Fatalf("expected Mgetstack, got %T", inst)
	}

	// Outgoing at offset 8 should be at SP+8
	if getstack.Ofs != 8 {
		t.Errorf("Ofs = %d, want 8", getstack.Ofs)
	}
}

func TestTranslateSetstackLocal(t *testing.T) {
	fn := linear.NewFunction("test", linear.Sig{})
	fn.Append(linear.Lsetstack{
		Src:  ltl.X0,
		Slot: linear.SlotLocal,
		Ofs:  0,
		Ty:   linear.Tlong,
	})

	layout := ComputeLayout(fn, 0)
	trans := NewSlotTranslator(layout)

	setstack := trans.TranslateSetstack(linear.Lsetstack{
		Src:  ltl.X0,
		Slot: linear.SlotLocal,
		Ofs:  0,
		Ty:   linear.Tlong,
	})

	if setstack.Ofs != layout.LocalOffset {
		t.Errorf("Ofs = %d, want %d", setstack.Ofs, layout.LocalOffset)
	}
	if setstack.Src != ltl.X0 {
		t.Errorf("Src = %v, want X0", setstack.Src)
	}
	if setstack.Ty != linear.Tlong {
		t.Errorf("Ty = %v, want Tlong", setstack.Ty)
	}
}

func TestTranslateSetstackOutgoing(t *testing.T) {
	fn := linear.NewFunction("test", linear.Sig{})
	layout := ComputeLayout(fn, 0)
	trans := NewSlotTranslator(layout)

	setstack := trans.TranslateSetstack(linear.Lsetstack{
		Src:  ltl.X1,
		Slot: linear.SlotOutgoing,
		Ofs:  16,
		Ty:   linear.Tint,
	})

	// Outgoing at offset 16 should be at SP+16
	if setstack.Ofs != 16 {
		t.Errorf("Ofs = %d, want 16", setstack.Ofs)
	}
}

func TestTranslateSlotOffset(t *testing.T) {
	fn := linear.NewFunction("test", linear.Sig{})
	fn.Append(linear.Lgetstack{
		Slot: linear.SlotLocal,
		Ofs:  0,
		Ty:   linear.Tlong,
		Dest: ltl.X0,
	})

	layout := ComputeLayout(fn, 2) // 2 callee-save regs
	trans := NewSlotTranslator(layout)

	// Test local slot translation
	localOfs := trans.TranslateSlotOffset(linear.SlotLocal, 0)
	if localOfs != layout.LocalOffset {
		t.Errorf("TranslateSlotOffset(Local, 0) = %d, want %d", localOfs, layout.LocalOffset)
	}

	// Test incoming slot translation
	incomingOfs := trans.TranslateSlotOffset(linear.SlotIncoming, 0)
	if incomingOfs != 16 {
		t.Errorf("TranslateSlotOffset(Incoming, 0) = %d, want 16", incomingOfs)
	}

	// Test outgoing slot translation
	// With frame layout: FP = SP + (TotalSize - 16)
	// Outgoing slot N is at SP + N = FP - (TotalSize - 16) + N
	// For TotalSize=48 and N=8: FP offset = -(48-16) + 8 = -24
	outgoingOfs := trans.TranslateSlotOffset(linear.SlotOutgoing, 8)
	expectedOutgoing := -(layout.TotalSize - 16) + 8
	if outgoingOfs != expectedOutgoing {
		t.Errorf("TranslateSlotOffset(Outgoing, 8) = %d, want %d", outgoingOfs, expectedOutgoing)
	}
}
