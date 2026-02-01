package selection

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/cminor"
	"github.com/raymyers/ralph-cc/pkg/cminorsel"
)

// Helper to create int constant expression
func intConst(v int32) cminor.Expr {
	return cminor.Econst{Const: cminor.Ointconst{Value: v}}
}

// Helper to create long constant expression
func longConst(v int64) cminor.Expr {
	return cminor.Econst{Const: cminor.Olongconst{Value: v}}
}

// Helper to create variable expression
func evar(name string) cminor.Expr {
	return cminor.Evar{Name: name}
}

// Helper to create add expression
func add(left, right cminor.Expr) cminor.Expr {
	return cminor.Ebinop{Op: cminor.Oadd, Left: left, Right: right}
}

// Helper to create long add expression
func addl(left, right cminor.Expr) cminor.Expr {
	return cminor.Ebinop{Op: cminor.Oaddl, Left: left, Right: right}
}

// Helper to create sub expression
func sub(left, right cminor.Expr) cminor.Expr {
	return cminor.Ebinop{Op: cminor.Osub, Left: left, Right: right}
}

// Helper to create shift left expression
func shl(base cminor.Expr, amt int32) cminor.Expr {
	return cminor.Ebinop{Op: cminor.Oshl, Left: base, Right: intConst(amt)}
}

func TestSelectAddressing_Fallback(t *testing.T) {
	// Simple variable - should fallback to Aindexed{0}
	addr := evar("ptr")
	result := SelectAddressing(addr, nil, nil)

	if _, ok := result.Mode.(cminorsel.Aindexed); !ok {
		t.Errorf("expected Aindexed, got %T", result.Mode)
	}
	if mode := result.Mode.(cminorsel.Aindexed); mode.Offset != 0 {
		t.Errorf("expected offset 0, got %d", mode.Offset)
	}
	if len(result.Args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(result.Args))
	}
}

func TestSelectAddressing_Aglobal(t *testing.T) {
	globals := map[string]bool{"global_var": true}

	tests := []struct {
		name     string
		addr     cminor.Expr
		wantSym  string
		wantOff  int64
		wantArgs int
	}{
		{
			name:     "direct global",
			addr:     evar("global_var"),
			wantSym:  "global_var",
			wantOff:  0,
			wantArgs: 0,
		},
		{
			name:     "global + constant",
			addr:     add(evar("global_var"), intConst(16)),
			wantSym:  "global_var",
			wantOff:  16,
			wantArgs: 0,
		},
		{
			name:     "constant + global (commutative)",
			addr:     add(intConst(8), evar("global_var")),
			wantSym:  "global_var",
			wantOff:  8,
			wantArgs: 0,
		},
		{
			name:     "global + long constant",
			addr:     addl(evar("global_var"), longConst(1024)),
			wantSym:  "global_var",
			wantOff:  1024,
			wantArgs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectAddressing(tt.addr, globals, nil)

			mode, ok := result.Mode.(cminorsel.Aglobal)
			if !ok {
				t.Fatalf("expected Aglobal, got %T", result.Mode)
			}
			if mode.Symbol != tt.wantSym {
				t.Errorf("expected symbol %q, got %q", tt.wantSym, mode.Symbol)
			}
			if mode.Offset != tt.wantOff {
				t.Errorf("expected offset %d, got %d", tt.wantOff, mode.Offset)
			}
			if len(result.Args) != tt.wantArgs {
				t.Errorf("expected %d args, got %d", tt.wantArgs, len(result.Args))
			}
		})
	}
}

func TestSelectAddressing_Ainstack(t *testing.T) {
	stackVars := map[string]int64{
		"local_arr": 0,
		"local_int": 32,
	}

	tests := []struct {
		name     string
		addr     cminor.Expr
		wantOff  int64
		wantArgs int
	}{
		{
			name:     "direct stack var",
			addr:     evar("local_arr"),
			wantOff:  0,
			wantArgs: 0,
		},
		{
			name:     "stack var with offset",
			addr:     evar("local_int"),
			wantOff:  32,
			wantArgs: 0,
		},
		{
			name:     "stack var + constant",
			addr:     add(evar("local_arr"), intConst(8)),
			wantOff:  8,
			wantArgs: 0,
		},
		{
			name:     "constant + stack var",
			addr:     add(intConst(16), evar("local_int")),
			wantOff:  48, // 32 + 16
			wantArgs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectAddressing(tt.addr, nil, stackVars)

			mode, ok := result.Mode.(cminorsel.Ainstack)
			if !ok {
				t.Fatalf("expected Ainstack, got %T", result.Mode)
			}
			if mode.Offset != tt.wantOff {
				t.Errorf("expected offset %d, got %d", tt.wantOff, mode.Offset)
			}
			if len(result.Args) != tt.wantArgs {
				t.Errorf("expected %d args, got %d", tt.wantArgs, len(result.Args))
			}
		})
	}
}

func TestSelectAddressing_Aindexed(t *testing.T) {
	tests := []struct {
		name    string
		addr    cminor.Expr
		wantOff int64
	}{
		{
			name:    "base + positive offset",
			addr:    add(evar("ptr"), intConst(16)),
			wantOff: 16,
		},
		{
			name:    "positive offset + base",
			addr:    add(intConst(32), evar("ptr")),
			wantOff: 32,
		},
		{
			name:    "base + long offset",
			addr:    addl(evar("ptr"), longConst(256)),
			wantOff: 256,
		},
		{
			name:    "base - offset (negative)",
			addr:    sub(evar("ptr"), intConst(8)),
			wantOff: -8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectAddressing(tt.addr, nil, nil)

			mode, ok := result.Mode.(cminorsel.Aindexed)
			if !ok {
				t.Fatalf("expected Aindexed, got %T", result.Mode)
			}
			if mode.Offset != tt.wantOff {
				t.Errorf("expected offset %d, got %d", tt.wantOff, mode.Offset)
			}
			if len(result.Args) != 1 {
				t.Errorf("expected 1 arg, got %d", len(result.Args))
			}
		})
	}
}

func TestSelectAddressing_Aindexed2(t *testing.T) {
	// base + index where both are variables
	addr := add(evar("base"), evar("index"))
	result := SelectAddressing(addr, nil, nil)

	if _, ok := result.Mode.(cminorsel.Aindexed2); !ok {
		t.Fatalf("expected Aindexed2, got %T", result.Mode)
	}
	if len(result.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(result.Args))
	}
}

func TestSelectAddressing_Aindexed2_NotForConstant(t *testing.T) {
	// base + constant should prefer Aindexed over Aindexed2
	addr := add(evar("base"), intConst(4))
	result := SelectAddressing(addr, nil, nil)

	// Should be Aindexed, not Aindexed2
	if _, ok := result.Mode.(cminorsel.Aindexed); !ok {
		t.Fatalf("expected Aindexed for base+const, got %T", result.Mode)
	}
}

func TestSelectAddressing_Aindexed2shift(t *testing.T) {
	tests := []struct {
		name      string
		addr      cminor.Expr
		wantShift int
	}{
		{
			name:      "base + (index << 0)",
			addr:      add(evar("base"), shl(evar("index"), 0)),
			wantShift: 0,
		},
		{
			name:      "base + (index << 1) - half-word",
			addr:      add(evar("base"), shl(evar("index"), 1)),
			wantShift: 1,
		},
		{
			name:      "base + (index << 2) - word",
			addr:      add(evar("base"), shl(evar("index"), 2)),
			wantShift: 2,
		},
		{
			name:      "base + (index << 3) - double-word",
			addr:      add(evar("base"), shl(evar("index"), 3)),
			wantShift: 3,
		},
		{
			name:      "(index << 2) + base - commutative",
			addr:      add(shl(evar("index"), 2), evar("base")),
			wantShift: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectAddressing(tt.addr, nil, nil)

			mode, ok := result.Mode.(cminorsel.Aindexed2shift)
			if !ok {
				t.Fatalf("expected Aindexed2shift, got %T", result.Mode)
			}
			if mode.Shift != tt.wantShift {
				t.Errorf("expected shift %d, got %d", tt.wantShift, mode.Shift)
			}
			if len(result.Args) != 2 {
				t.Errorf("expected 2 args, got %d", len(result.Args))
			}
		})
	}
}

func TestSelectAddressing_Aindexed2shift_InvalidShift(t *testing.T) {
	// Shift > 3 should not use Aindexed2shift (invalid for ARM64)
	addr := add(evar("base"), shl(evar("index"), 4))
	result := SelectAddressing(addr, nil, nil)

	// Should fall back to Aindexed2, not Aindexed2shift
	if _, ok := result.Mode.(cminorsel.Aindexed2shift); ok {
		t.Fatalf("shift > 3 should not use Aindexed2shift")
	}
}

func TestSelectAddressing_Priority(t *testing.T) {
	// Test that more specific modes are preferred
	globals := map[string]bool{"g": true}
	stackVars := map[string]int64{"s": 0}

	// Global takes priority over everything
	result := SelectAddressing(evar("g"), globals, stackVars)
	if _, ok := result.Mode.(cminorsel.Aglobal); !ok {
		t.Errorf("global should take priority, got %T", result.Mode)
	}

	// Stack takes priority when not global
	result = SelectAddressing(evar("s"), nil, stackVars)
	if _, ok := result.Mode.(cminorsel.Ainstack); !ok {
		t.Errorf("stack should take priority, got %T", result.Mode)
	}
}

func TestExtractConstantOffset(t *testing.T) {
	tests := []struct {
		name    string
		expr    cminor.Expr
		wantOff *int64
	}{
		{
			name:    "int constant",
			expr:    intConst(42),
			wantOff: ptr(42),
		},
		{
			name:    "long constant",
			expr:    longConst(100),
			wantOff: ptr(100),
		},
		{
			name:    "variable - not constant",
			expr:    evar("x"),
			wantOff: nil,
		},
		{
			name:    "float constant - not offset",
			expr:    cminor.Econst{Const: cminor.Ofloatconst{Value: 3.14}},
			wantOff: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractConstantOffset(tt.expr)
			if tt.wantOff == nil {
				if got != nil {
					t.Errorf("expected nil, got %d", *got)
				}
			} else {
				if got == nil {
					t.Errorf("expected %d, got nil", *tt.wantOff)
				} else if *got != *tt.wantOff {
					t.Errorf("expected %d, got %d", *tt.wantOff, *got)
				}
			}
		})
	}
}

func ptr(v int64) *int64 {
	return &v
}
