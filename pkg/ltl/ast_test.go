package ltl

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/rtl"
)

func TestLocationTypes(t *testing.T) {
	tests := []struct {
		name string
		loc  Loc
	}{
		{"register X0", R{Reg: X0}},
		{"register D0", R{Reg: D0}},
		{"local slot", S{Slot: SlotLocal, Ofs: 0, Ty: Tint}},
		{"incoming slot", S{Slot: SlotIncoming, Ofs: 8, Ty: Tlong}},
		{"outgoing slot", S{Slot: SlotOutgoing, Ofs: 16, Ty: Tfloat}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify it implements Loc interface
			var _ Loc = tt.loc
		})
	}
}

func TestMachineRegisterNames(t *testing.T) {
	tests := []struct {
		reg  MReg
		want string
	}{
		{X0, "X0"},
		{X1, "X1"},
		{X15, "X15"},
		{X29, "X29"},
		{X30, "X30"},
		{D0, "D0"},
		{D1, "D1"},
		{D15, "D15"},
		{D31, "D31"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.reg.String(); got != tt.want {
				t.Errorf("MReg.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMachineRegisterTypes(t *testing.T) {
	// Integer registers
	for r := X0; r <= X30; r++ {
		if !r.IsInteger() {
			t.Errorf("%s.IsInteger() = false, want true", r)
		}
		if r.IsFloat() {
			t.Errorf("%s.IsFloat() = true, want false", r)
		}
	}

	// Float registers
	for r := D0; r <= D31; r++ {
		if r.IsInteger() {
			t.Errorf("%s.IsInteger() = true, want false", r)
		}
		if !r.IsFloat() {
			t.Errorf("%s.IsFloat() = false, want true", r)
		}
	}
}

func TestSlotKindString(t *testing.T) {
	tests := []struct {
		kind SlotKind
		want string
	}{
		{SlotLocal, "Local"},
		{SlotIncoming, "Incoming"},
		{SlotOutgoing, "Outgoing"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("SlotKind.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTypString(t *testing.T) {
	tests := []struct {
		typ  Typ
		want string
	}{
		{Tint, "Tint"},
		{Tfloat, "Tfloat"},
		{Tlong, "Tlong"},
		{Tsingle, "Tsingle"},
		{Tany32, "Tany32"},
		{Tany64, "Tany64"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("Typ.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInstructionTypes(t *testing.T) {
	tests := []struct {
		name  string
		instr Instruction
	}{
		{"Lnop", Lnop{}},
		{"Lop", Lop{Op: rtl.Oadd{}, Args: []Loc{R{X0}, R{X1}}, Dest: R{X2}}},
		{"Lload", Lload{Chunk: Mint64, Addr: Aindexed{Offset: 0}, Args: []Loc{R{X0}}, Dest: R{X1}}},
		{"Lstore", Lstore{Chunk: Mint64, Addr: Aindexed{Offset: 0}, Args: []Loc{R{X0}}, Src: R{X1}}},
		{"Lcall", Lcall{Fn: FunSymbol{Name: "foo"}, Args: []Loc{R{X0}}}},
		{"Ltailcall", Ltailcall{Fn: FunSymbol{Name: "bar"}, Args: []Loc{}}},
		{"Lbuiltin", Lbuiltin{Builtin: "__builtin_trap", Args: []Loc{}}},
		{"Lbranch", Lbranch{Succ: 1}},
		{"Lcond", Lcond{Cond: rtl.Ccomp{Cond: rtl.Ceq}, Args: []Loc{R{X0}, R{X1}}, IfSo: 1, IfNot: 2}},
		{"Ljumptable", Ljumptable{Arg: R{X0}, Targets: []Node{1, 2, 3}}},
		{"Lreturn", Lreturn{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify it implements Instruction interface
			var _ Instruction = tt.instr
		})
	}
}

func TestBBlockConstruction(t *testing.T) {
	// Create a basic block with some instructions
	bb := &BBlock{
		Body: []Instruction{
			Lop{Op: rtl.Oadd{}, Args: []Loc{R{X0}, R{X1}}, Dest: R{X2}},
			Lbranch{Succ: 2},
		},
	}

	if len(bb.Body) != 2 {
		t.Errorf("BBlock.Body length = %d, want 2", len(bb.Body))
	}
}

func TestFunctionConstruction(t *testing.T) {
	fn := NewFunction("test", Sig{})

	if fn.Name != "test" {
		t.Errorf("Function.Name = %q, want %q", fn.Name, "test")
	}
	if fn.Code == nil {
		t.Error("Function.Code is nil, should be initialized")
	}

	// Add a basic block
	fn.Code[1] = &BBlock{
		Body: []Instruction{
			Lop{Op: rtl.Ointconst{Value: 42}, Args: []Loc{}, Dest: R{X0}},
			Lreturn{},
		},
	}
	fn.Entrypoint = 1

	if len(fn.Code) != 1 {
		t.Errorf("Function has %d blocks, want 1", len(fn.Code))
	}
	if fn.Entrypoint != 1 {
		t.Errorf("Function.Entrypoint = %d, want 1", fn.Entrypoint)
	}
}

func TestFunRefTypes(t *testing.T) {
	tests := []struct {
		name   string
		funref FunRef
	}{
		{"FunReg", FunReg{Loc: R{X0}}},
		{"FunSymbol", FunSymbol{Name: "printf"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var _ FunRef = tt.funref
		})
	}
}

func TestProgramConstruction(t *testing.T) {
	prog := &Program{
		Globals: []GlobVar{
			{Name: "counter", Size: 4},
		},
		Functions: []Function{
			{
				Name:       "main",
				Entrypoint: 1,
				Code: map[Node]*BBlock{
					1: {Body: []Instruction{Lreturn{}}},
				},
			},
		},
	}

	if len(prog.Globals) != 1 {
		t.Errorf("Program has %d globals, want 1", len(prog.Globals))
	}
	if len(prog.Functions) != 1 {
		t.Errorf("Program has %d functions, want 1", len(prog.Functions))
	}
}

func TestStackSlotWithTypes(t *testing.T) {
	// Test all slot types with all data types
	types := []Typ{Tint, Tfloat, Tlong, Tsingle, Tany32, Tany64}
	slots := []SlotKind{SlotLocal, SlotIncoming, SlotOutgoing}

	for _, slot := range slots {
		for _, typ := range types {
			s := S{Slot: slot, Ofs: 0, Ty: typ}
			var _ Loc = s // Verify implements Loc
		}
	}
}
