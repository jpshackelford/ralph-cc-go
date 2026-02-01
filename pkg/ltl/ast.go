// Package ltl defines the LTL (Location Transfer Language) intermediate representation.
// LTL replaces RTL's infinite pseudo-registers with physical registers and stack slots.
// This is the output of register allocation. Instructions are grouped into basic blocks.
// This mirrors CompCert's backend/LTL.v
package ltl

import "github.com/raymyers/ralph-cc/pkg/rtl"

// Re-export types from rtl that are used in LTL
type (
	Chunk          = rtl.Chunk
	AddressingMode = rtl.AddressingMode
	Sig            = rtl.Sig
	Operation      = rtl.Operation
	Condition      = rtl.Condition
	ConditionCode  = rtl.ConditionCode
)

// Re-export addressing mode types
type (
	Aindexed       = rtl.Aindexed
	Aindexed2      = rtl.Aindexed2
	Aglobal        = rtl.Aglobal
	Ainstack       = rtl.Ainstack
	Aindexed2shift = rtl.Aindexed2shift
)

// Re-export chunk constants
const (
	Mint8signed    = rtl.Mint8signed
	Mint8unsigned  = rtl.Mint8unsigned
	Mint16signed   = rtl.Mint16signed
	Mint16unsigned = rtl.Mint16unsigned
	Mint32         = rtl.Mint32
	Mint64         = rtl.Mint64
	Mfloat32       = rtl.Mfloat32
	Mfloat64       = rtl.Mfloat64
)

// Node represents a program point in the CFG (basic block label)
type Node int

// --- Location Types ---
// In LTL, values live in locations rather than pseudo-registers.
// A location is either a machine register or a stack slot.

// Loc represents a location (register or stack slot)
type Loc interface {
	implLoc()
}

// R represents a machine register location
type R struct {
	Reg MReg
}

// S represents a stack slot location
type S struct {
	Slot SlotKind
	Ofs  int64
	Ty   Typ
}

func (R) implLoc() {}
func (S) implLoc() {}

// --- Machine Registers (ARM64/AArch64) ---

// MReg represents a physical machine register
type MReg int

// ARM64 integer registers
const (
	X0 MReg = iota
	X1
	X2
	X3
	X4
	X5
	X6
	X7
	X8
	X9
	X10
	X11
	X12
	X13
	X14
	X15
	X16
	X17
	X18
	X19
	X20
	X21
	X22
	X23
	X24
	X25
	X26
	X27
	X28
	X29 // FP (frame pointer)
	X30 // LR (link register)
)

// ARM64 floating-point registers (start at 64 to avoid collision)
const (
	D0 MReg = iota + 64
	D1
	D2
	D3
	D4
	D5
	D6
	D7
	D8
	D9
	D10
	D11
	D12
	D13
	D14
	D15
	D16
	D17
	D18
	D19
	D20
	D21
	D22
	D23
	D24
	D25
	D26
	D27
	D28
	D29
	D30
	D31
)

// IsInteger returns true if the register is an integer register
func (r MReg) IsInteger() bool {
	return r <= X30
}

// IsFloat returns true if the register is a floating-point register
func (r MReg) IsFloat() bool {
	return r >= D0 && r <= D31
}

// String returns the name of the register
func (r MReg) String() string {
	if r <= X30 {
		names := []string{
			"X0", "X1", "X2", "X3", "X4", "X5", "X6", "X7",
			"X8", "X9", "X10", "X11", "X12", "X13", "X14", "X15",
			"X16", "X17", "X18", "X19", "X20", "X21", "X22", "X23",
			"X24", "X25", "X26", "X27", "X28", "X29", "X30",
		}
		return names[r]
	}
	if r >= D0 && r <= D31 {
		names := []string{
			"D0", "D1", "D2", "D3", "D4", "D5", "D6", "D7",
			"D8", "D9", "D10", "D11", "D12", "D13", "D14", "D15",
			"D16", "D17", "D18", "D19", "D20", "D21", "D22", "D23",
			"D24", "D25", "D26", "D27", "D28", "D29", "D30", "D31",
		}
		return names[r-D0]
	}
	return "?"
}

// --- Stack Slot Kinds ---

// SlotKind represents the kind of stack slot
type SlotKind int

const (
	SlotLocal    SlotKind = iota // Local variable slot
	SlotIncoming                 // Incoming argument on stack
	SlotOutgoing                 // Outgoing argument on stack
)

func (s SlotKind) String() string {
	switch s {
	case SlotLocal:
		return "Local"
	case SlotIncoming:
		return "Incoming"
	case SlotOutgoing:
		return "Outgoing"
	}
	return "?"
}

// --- Type for Stack Slots ---

// Typ represents the type stored in a stack slot
type Typ int

const (
	Tint Typ = iota
	Tfloat
	Tlong
	Tsingle
	Tany32
	Tany64
)

func (t Typ) String() string {
	names := []string{"Tint", "Tfloat", "Tlong", "Tsingle", "Tany32", "Tany64"}
	if int(t) < len(names) {
		return names[t]
	}
	return "?"
}

// --- Instruction Types ---
// LTL instructions operate on locations (registers/slots) instead of pseudo-registers.
// Instructions are grouped in basic blocks, not individual CFG nodes.

// Instruction is the interface for LTL instructions
type Instruction interface {
	implInstruction()
}

// Lnop is a no-operation (placeholder, typically eliminated)
type Lnop struct{}

// Lop performs an operation: dest = op(args...)
type Lop struct {
	Op   Operation // the operation
	Args []Loc     // source locations
	Dest Loc       // destination location
}

// Lload loads from memory: dest = Mem[addr(args...)]
type Lload struct {
	Chunk Chunk          // memory access size/type
	Addr  AddressingMode // addressing mode
	Args  []Loc          // locations for addressing
	Dest  Loc            // destination location
}

// Lstore stores to memory: Mem[addr(args...)] = src
type Lstore struct {
	Chunk Chunk          // memory access size/type
	Addr  AddressingMode // addressing mode
	Args  []Loc          // locations for addressing (including value to store)
	Src   Loc            // source location (value to store)
}

// Lcall performs a function call
type Lcall struct {
	Sig  Sig    // function signature
	Fn   FunRef // function to call (reg or symbol)
	Args []Loc  // argument locations
}

// Ltailcall performs a tail call (no return to caller)
type Ltailcall struct {
	Sig  Sig    // function signature
	Fn   FunRef // function to call
	Args []Loc  // argument locations
}

// Lbuiltin calls a builtin function
type Lbuiltin struct {
	Builtin string // builtin function name
	Args    []Loc  // argument locations
	Dest    *Loc   // destination location (nil if no result)
}

// Lbranch is an unconditional branch
type Lbranch struct {
	Succ Node // branch target
}

// Lcond is a conditional branch
type Lcond struct {
	Cond  ConditionCode // condition to evaluate
	Args  []Loc         // argument locations
	IfSo  Node          // branch target if condition is true
	IfNot Node          // branch target if condition is false
}

// Ljumptable is an indexed jump (switch)
type Ljumptable struct {
	Arg     Loc    // location containing index
	Targets []Node // jump targets
}

// Lreturn returns from the function
type Lreturn struct{}

// Marker methods for Instruction interface
func (Lnop) implInstruction()       {}
func (Lop) implInstruction()        {}
func (Lload) implInstruction()      {}
func (Lstore) implInstruction()     {}
func (Lcall) implInstruction()      {}
func (Ltailcall) implInstruction()  {}
func (Lbuiltin) implInstruction()   {}
func (Lbranch) implInstruction()    {}
func (Lcond) implInstruction()      {}
func (Ljumptable) implInstruction() {}
func (Lreturn) implInstruction()    {}

// --- Basic Block ---
// In LTL, instructions are grouped into basic blocks

// BBlock represents a basic block: a sequence of instructions ending with a control flow
type BBlock struct {
	Body []Instruction // instructions in the block (last is terminator)
}

// --- Function Reference ---

// FunRef represents a function reference (either register or symbol)
type FunRef interface {
	implFunRef()
}

// FunReg is a function pointer in a register
type FunReg struct {
	Loc Loc
}

// FunSymbol is a named function symbol
type FunSymbol struct {
	Name string
}

func (FunReg) implFunRef()    {}
func (FunSymbol) implFunRef() {}

// --- Function and Program ---

// Function represents an LTL function
type Function struct {
	Name       string            // function name
	Sig        Sig               // function signature
	Params     []Loc             // parameter locations
	Stacksize  int64             // stack frame size
	Code       map[Node]*BBlock  // CFG: node -> basic block
	Entrypoint Node              // entry node
}

// GlobVar represents a global variable
type GlobVar struct {
	Name string
	Size int64
	Init []byte
}

// Program represents a complete LTL program
type Program struct {
	Globals   []GlobVar
	Functions []Function
}

// NewFunction creates a new LTL function with initialized code map
func NewFunction(name string, sig Sig) *Function {
	return &Function{
		Name: name,
		Sig:  sig,
		Code: make(map[Node]*BBlock),
	}
}
