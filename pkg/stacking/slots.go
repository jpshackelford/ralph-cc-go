package stacking

import (
	"github.com/raymyers/ralph-cc/pkg/linear"
	"github.com/raymyers/ralph-cc/pkg/mach"
)

// SlotTranslator translates abstract stack slots to concrete offsets
type SlotTranslator struct {
	layout *FrameLayout
}

// NewSlotTranslator creates a translator with the given layout
func NewSlotTranslator(layout *FrameLayout) *SlotTranslator {
	return &SlotTranslator{layout: layout}
}

// TranslateGetstack converts Lgetstack to Mgetstack or Mgetparam
func (t *SlotTranslator) TranslateGetstack(inst linear.Lgetstack) mach.Instruction {
	switch inst.Slot {
	case linear.SlotLocal:
		// Local slot: access via FP with concrete offset
		return mach.Mgetstack{
			Ofs:  t.layout.LocalSlotOffset(inst.Ofs),
			Ty:   inst.Ty,
			Dest: inst.Dest,
		}
	case linear.SlotIncoming:
		// Incoming parameter: use Mgetparam for clarity
		return mach.Mgetparam{
			Ofs:  t.layout.IncomingSlotOffset(inst.Ofs),
			Ty:   inst.Ty,
			Dest: inst.Dest,
		}
	case linear.SlotOutgoing:
		// Outgoing arg: access via SP
		return mach.Mgetstack{
			Ofs:  t.layout.OutgoingSlotOffset(inst.Ofs),
			Ty:   inst.Ty,
			Dest: inst.Dest,
		}
	default:
		panic("unknown slot kind")
	}
}

// TranslateSetstack converts Lsetstack to Msetstack
func (t *SlotTranslator) TranslateSetstack(inst linear.Lsetstack) mach.Msetstack {
	var ofs int64
	switch inst.Slot {
	case linear.SlotLocal:
		ofs = t.layout.LocalSlotOffset(inst.Ofs)
	case linear.SlotIncoming:
		// Writing to incoming slot is unusual but possible
		ofs = t.layout.IncomingSlotOffset(inst.Ofs)
	case linear.SlotOutgoing:
		ofs = t.layout.OutgoingSlotOffset(inst.Ofs)
	default:
		panic("unknown slot kind")
	}

	return mach.Msetstack{
		Src: inst.Src,
		Ofs: ofs,
		Ty:  inst.Ty,
	}
}

// TranslateSlotInLoc translates a stack slot location to concrete offset
// Returns the offset from the appropriate base (FP or SP)
func (t *SlotTranslator) TranslateSlotOffset(slot linear.SlotKind, ofs int64) int64 {
	switch slot {
	case linear.SlotLocal:
		return t.layout.LocalSlotOffset(ofs)
	case linear.SlotIncoming:
		return t.layout.IncomingSlotOffset(ofs)
	case linear.SlotOutgoing:
		return t.layout.OutgoingSlotOffset(ofs)
	default:
		panic("unknown slot kind")
	}
}
