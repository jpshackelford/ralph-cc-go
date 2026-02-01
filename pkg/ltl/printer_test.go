package ltl

import (
	"bytes"
	"strings"
	"testing"

	"github.com/raymyers/ralph-cc/pkg/rtl"
)

func TestPrintLocation(t *testing.T) {
	tests := []struct {
		name string
		loc  Loc
		want string
	}{
		{"register X0", R{Reg: X0}, "X0"},
		{"register X15", R{Reg: X15}, "X15"},
		{"register D0", R{Reg: D0}, "D0"},
		{"local slot", S{Slot: SlotLocal, Ofs: 0, Ty: Tint}, "S(Local, 0, Tint)"},
		{"incoming slot", S{Slot: SlotIncoming, Ofs: 8, Ty: Tlong}, "S(Incoming, 8, Tlong)"},
		{"outgoing slot", S{Slot: SlotOutgoing, Ofs: 16, Ty: Tfloat}, "S(Outgoing, 16, Tfloat)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printLoc(tt.loc)
			if got := buf.String(); got != tt.want {
				t.Errorf("printLoc = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintSimpleFunction(t *testing.T) {
	fn := &Function{
		Name:       "test",
		Entrypoint: 1,
		Params:     []Loc{R{Reg: X0}},
		Code: map[Node]*BBlock{
			1: {Body: []Instruction{
				Lop{Op: rtl.Ointconst{Value: 42}, Args: []Loc{}, Dest: R{Reg: X0}},
				Lbranch{Succ: 2},
			}},
			2: {Body: []Instruction{
				Lreturn{},
			}},
		},
	}

	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintFunction(fn)

	output := buf.String()

	// Check function header
	if !strings.Contains(output, "test(X0)") {
		t.Errorf("output should contain function header, got:\n%s", output)
	}

	// Check block 1
	if !strings.Contains(output, "1: {") {
		t.Errorf("output should contain block 1, got:\n%s", output)
	}

	// Check Lop
	if !strings.Contains(output, "Lop(Ointconst(42)") {
		t.Errorf("output should contain Lop instruction, got:\n%s", output)
	}

	// Check entry point
	if !strings.Contains(output, "entry: 1") {
		t.Errorf("output should contain entry point, got:\n%s", output)
	}
}

func TestPrintInstructions(t *testing.T) {
	tests := []struct {
		name  string
		instr Instruction
		want  string
	}{
		{"Lnop", Lnop{}, "Lnop"},
		{"Lbranch", Lbranch{Succ: 5}, "Lbranch 5"},
		{"Lreturn", Lreturn{}, "Lreturn"},
		{
			"Lop add",
			Lop{Op: rtl.Oadd{}, Args: []Loc{R{Reg: X0}, R{Reg: X1}}, Dest: R{Reg: X2}},
			"Lop(Oadd, [X0; X1], X2)",
		},
		{
			"Lload",
			Lload{Chunk: Mint64, Addr: Aindexed{Offset: 8}, Args: []Loc{R{Reg: X0}}, Dest: R{Reg: X1}},
			"Lload(Mint64, Aindexed(8), [X0], X1)",
		},
		{
			"Lstore",
			Lstore{Chunk: Mint32, Addr: Aindexed{Offset: 0}, Args: []Loc{R{Reg: X0}}, Src: R{Reg: X1}},
			"Lstore(Mint32, Aindexed(0), [X0], X1)",
		},
		{
			"Lcall",
			Lcall{Fn: FunSymbol{Name: "foo"}, Args: []Loc{R{Reg: X0}}},
			"Lcall(\"foo\", [X0])",
		},
		{
			"Ltailcall",
			Ltailcall{Fn: FunSymbol{Name: "bar"}, Args: []Loc{}},
			"Ltailcall(\"bar\", [])",
		},
		{
			"Lcond",
			Lcond{Cond: rtl.Ccompimm{Cond: rtl.Ceq, N: 0}, Args: []Loc{R{Reg: X0}}, IfSo: 3, IfNot: 4},
			"Lcond(Ccompimm(==, 0), 3, 4)",
		},
		{
			"Ljumptable",
			Ljumptable{Arg: R{Reg: X0}, Targets: []Node{1, 2, 3}},
			"Ljumptable(X0, [1; 2; 3])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printInstruction(tt.instr)
			if got := buf.String(); got != tt.want {
				t.Errorf("printInstruction = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintOperations(t *testing.T) {
	tests := []struct {
		op   Operation
		want string
	}{
		{rtl.Omove{}, "Omove"},
		{rtl.Ointconst{Value: 42}, "Ointconst(42)"},
		{rtl.Olongconst{Value: 100}, "Olongconst(100L)"},
		{rtl.Oadd{}, "Oadd"},
		{rtl.Oaddimm{N: 5}, "Oaddimm(5)"},
		{rtl.Osub{}, "Osub"},
		{rtl.Omul{}, "Omul"},
		{rtl.Odiv{}, "Odiv"},
		{rtl.Oand{}, "Oand"},
		{rtl.Oor{}, "Oor"},
		{rtl.Oxor{}, "Oxor"},
		{rtl.Oshl{}, "Oshl"},
		{rtl.Oshlimm{N: 2}, "Oshlimm(2)"},
		{rtl.Oshr{}, "Oshr"},
		{rtl.Oshru{}, "Oshru"},
		{rtl.Oaddl{}, "Oaddl"},
		{rtl.Osubl{}, "Osubl"},
		{rtl.Omull{}, "Omull"},
		{rtl.Oaddf{}, "Oaddf"},
		{rtl.Omulf{}, "Omulf"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printOperation(tt.op)
			if got := buf.String(); got != tt.want {
				t.Errorf("printOperation = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintProgram(t *testing.T) {
	prog := &Program{
		Globals: []GlobVar{
			{Name: "counter", Size: 8},
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

	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.PrintProgram(prog)

	output := buf.String()

	// Check global
	if !strings.Contains(output, "var \"counter\"[8]") {
		t.Errorf("output should contain global variable, got:\n%s", output)
	}

	// Check function
	if !strings.Contains(output, "main()") {
		t.Errorf("output should contain main function, got:\n%s", output)
	}
}

func TestPrintBuiltin(t *testing.T) {
	instr := Lbuiltin{
		Builtin: "__builtin_trap",
		Args:    []Loc{R{Reg: X0}},
		Dest:    nil,
	}

	var buf bytes.Buffer
	p := NewPrinter(&buf)
	p.printInstruction(instr)

	got := buf.String()
	if !strings.Contains(got, "Lbuiltin") || !strings.Contains(got, "__builtin_trap") {
		t.Errorf("printInstruction(builtin) = %q, should contain Lbuiltin and builtin name", got)
	}
}

func TestPrintConditionCodes(t *testing.T) {
	tests := []struct {
		cond ConditionCode
		want string
	}{
		{rtl.Ccomp{Cond: rtl.Ceq}, "Ccomp(==)"},
		{rtl.Ccomp{Cond: rtl.Cne}, "Ccomp(!=)"},
		{rtl.Ccomp{Cond: rtl.Clt}, "Ccomp(<)"},
		{rtl.Ccompu{Cond: rtl.Cge}, "Ccompu(>=)"},
		{rtl.Ccompimm{Cond: rtl.Ceq, N: 0}, "Ccompimm(==, 0)"},
		{rtl.Ccompl{Cond: rtl.Cgt}, "Ccompl(>)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printConditionCode(tt.cond, nil)
			if got := buf.String(); got != tt.want {
				t.Errorf("printConditionCode = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintAddressingModes(t *testing.T) {
	tests := []struct {
		addr AddressingMode
		want string
	}{
		{Aindexed{Offset: 0}, "Aindexed(0)"},
		{Aindexed{Offset: 8}, "Aindexed(8)"},
		{Aindexed2{}, "Aindexed2"},
		{Aindexed2shift{Shift: 3}, "Aindexed2shift(3)"},
		{Aglobal{Symbol: "foo", Offset: 4}, "Aglobal(\"foo\", 4)"},
		{Ainstack{Offset: 16}, "Ainstack(16)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			var buf bytes.Buffer
			p := NewPrinter(&buf)
			p.printAddressingMode(tt.addr)
			if got := buf.String(); got != tt.want {
				t.Errorf("printAddressingMode = %q, want %q", got, tt.want)
			}
		})
	}
}
