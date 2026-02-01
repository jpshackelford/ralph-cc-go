package cshmgen

import (
	"testing"

	"github.com/raymyers/ralph-cc/pkg/clight"
	"github.com/raymyers/ralph-cc/pkg/csharpminor"
	"github.com/raymyers/ralph-cc/pkg/ctypes"
)

func TestTranslateUnaryOp_Neg(t *testing.T) {
	tests := []struct {
		name    string
		argType ctypes.Type
		want    csharpminor.UnaryOp
	}{
		{"int", ctypes.Int(), csharpminor.Onegint},
		{"unsigned int", ctypes.UInt(), csharpminor.Onegint},
		{"long", ctypes.Long(), csharpminor.Onegl},
		{"float", ctypes.Float(), csharpminor.Onegs},
		{"double", ctypes.Double(), csharpminor.Onegf},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateUnaryOp(clight.Oneg, tt.argType)
			if got != tt.want {
				t.Errorf("TranslateUnaryOp(Oneg, %v) = %v, want %v", tt.argType, got, tt.want)
			}
		})
	}
}

func TestTranslateUnaryOp_Bitnot(t *testing.T) {
	tests := []struct {
		name    string
		argType ctypes.Type
		want    csharpminor.UnaryOp
	}{
		{"int", ctypes.Int(), csharpminor.Onotint},
		{"long", ctypes.Long(), csharpminor.Onotl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateUnaryOp(clight.Onotint, tt.argType)
			if got != tt.want {
				t.Errorf("TranslateUnaryOp(Onotint, %v) = %v, want %v", tt.argType, got, tt.want)
			}
		})
	}
}

func TestTranslateUnaryOp_Notbool(t *testing.T) {
	got := TranslateUnaryOp(clight.Onotbool, ctypes.Int())
	if got != csharpminor.Onotbool {
		t.Errorf("TranslateUnaryOp(Onotbool, int) = %v, want Onotbool", got)
	}
}

func TestTranslateBinaryOp_Add(t *testing.T) {
	tests := []struct {
		name     string
		leftType ctypes.Type
		want     csharpminor.BinaryOp
	}{
		{"int", ctypes.Int(), csharpminor.Oadd},
		{"long", ctypes.Long(), csharpminor.Oaddl},
		{"float", ctypes.Float(), csharpminor.Oadds},
		{"double", ctypes.Double(), csharpminor.Oaddf},
		{"pointer", ctypes.Pointer(ctypes.Int()), csharpminor.Oaddl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := TranslateBinaryOp(clight.Oadd, tt.leftType, tt.leftType)
			if got != tt.want {
				t.Errorf("TranslateBinaryOp(Oadd, %v) = %v, want %v", tt.leftType, got, tt.want)
			}
		})
	}
}

func TestTranslateBinaryOp_Sub(t *testing.T) {
	tests := []struct {
		name     string
		leftType ctypes.Type
		want     csharpminor.BinaryOp
	}{
		{"int", ctypes.Int(), csharpminor.Osub},
		{"long", ctypes.Long(), csharpminor.Osubl},
		{"float", ctypes.Float(), csharpminor.Osubs},
		{"double", ctypes.Double(), csharpminor.Osubf},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := TranslateBinaryOp(clight.Osub, tt.leftType, tt.leftType)
			if got != tt.want {
				t.Errorf("TranslateBinaryOp(Osub, %v) = %v, want %v", tt.leftType, got, tt.want)
			}
		})
	}
}

func TestTranslateBinaryOp_Mul(t *testing.T) {
	tests := []struct {
		name     string
		leftType ctypes.Type
		want     csharpminor.BinaryOp
	}{
		{"int", ctypes.Int(), csharpminor.Omul},
		{"long", ctypes.Long(), csharpminor.Omull},
		{"float", ctypes.Float(), csharpminor.Omuls},
		{"double", ctypes.Double(), csharpminor.Omulf},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := TranslateBinaryOp(clight.Omul, tt.leftType, tt.leftType)
			if got != tt.want {
				t.Errorf("TranslateBinaryOp(Omul, %v) = %v, want %v", tt.leftType, got, tt.want)
			}
		})
	}
}

func TestTranslateBinaryOp_Div(t *testing.T) {
	tests := []struct {
		name     string
		leftType ctypes.Type
		want     csharpminor.BinaryOp
	}{
		{"signed int", ctypes.Int(), csharpminor.Odiv},
		{"unsigned int", ctypes.UInt(), csharpminor.Odivu},
		{"signed long", ctypes.Long(), csharpminor.Odivl},
		{"unsigned long", ctypes.Tlong{Sign: ctypes.Unsigned}, csharpminor.Odivlu},
		{"float", ctypes.Float(), csharpminor.Odivs},
		{"double", ctypes.Double(), csharpminor.Odivf},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := TranslateBinaryOp(clight.Odiv, tt.leftType, tt.leftType)
			if got != tt.want {
				t.Errorf("TranslateBinaryOp(Odiv, %v) = %v, want %v", tt.leftType, got, tt.want)
			}
		})
	}
}

func TestTranslateBinaryOp_Mod(t *testing.T) {
	tests := []struct {
		name     string
		leftType ctypes.Type
		want     csharpminor.BinaryOp
	}{
		{"signed int", ctypes.Int(), csharpminor.Omod},
		{"unsigned int", ctypes.UInt(), csharpminor.Omodu},
		{"signed long", ctypes.Long(), csharpminor.Omodl},
		{"unsigned long", ctypes.Tlong{Sign: ctypes.Unsigned}, csharpminor.Omodlu},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := TranslateBinaryOp(clight.Omod, tt.leftType, tt.leftType)
			if got != tt.want {
				t.Errorf("TranslateBinaryOp(Omod, %v) = %v, want %v", tt.leftType, got, tt.want)
			}
		})
	}
}

func TestTranslateBinaryOp_Bitwise(t *testing.T) {
	tests := []struct {
		name     string
		op       clight.BinaryOp
		leftType ctypes.Type
		want     csharpminor.BinaryOp
	}{
		{"and int", clight.Oand, ctypes.Int(), csharpminor.Oand},
		{"and long", clight.Oand, ctypes.Long(), csharpminor.Oandl},
		{"or int", clight.Oor, ctypes.Int(), csharpminor.Oor},
		{"or long", clight.Oor, ctypes.Long(), csharpminor.Oorl},
		{"xor int", clight.Oxor, ctypes.Int(), csharpminor.Oxor},
		{"xor long", clight.Oxor, ctypes.Long(), csharpminor.Oxorl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := TranslateBinaryOp(tt.op, tt.leftType, tt.leftType)
			if got != tt.want {
				t.Errorf("TranslateBinaryOp(%v, %v) = %v, want %v", tt.op, tt.leftType, got, tt.want)
			}
		})
	}
}

func TestTranslateBinaryOp_Shift(t *testing.T) {
	tests := []struct {
		name     string
		op       clight.BinaryOp
		leftType ctypes.Type
		want     csharpminor.BinaryOp
	}{
		{"shl int", clight.Oshl, ctypes.Int(), csharpminor.Oshl},
		{"shl long", clight.Oshl, ctypes.Long(), csharpminor.Oshll},
		{"shr signed int", clight.Oshr, ctypes.Int(), csharpminor.Oshr},
		{"shr unsigned int", clight.Oshr, ctypes.UInt(), csharpminor.Oshru},
		{"shr signed long", clight.Oshr, ctypes.Long(), csharpminor.Oshrl},
		{"shr unsigned long", clight.Oshr, ctypes.Tlong{Sign: ctypes.Unsigned}, csharpminor.Oshrlu},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := TranslateBinaryOp(tt.op, tt.leftType, ctypes.Int())
			if got != tt.want {
				t.Errorf("TranslateBinaryOp(%v, %v) = %v, want %v", tt.op, tt.leftType, got, tt.want)
			}
		})
	}
}

func TestTranslateBinaryOp_Comparison(t *testing.T) {
	tests := []struct {
		name     string
		op       clight.BinaryOp
		leftType ctypes.Type
		wantOp   csharpminor.BinaryOp
		wantCmp  csharpminor.Comparison
	}{
		{"eq signed int", clight.Oeq, ctypes.Int(), csharpminor.Ocmp, csharpminor.Ceq},
		{"ne unsigned int", clight.One, ctypes.UInt(), csharpminor.Ocmpu, csharpminor.Cne},
		{"lt signed long", clight.Olt, ctypes.Long(), csharpminor.Ocmpl, csharpminor.Clt},
		{"le unsigned long", clight.Ole, ctypes.Tlong{Sign: ctypes.Unsigned}, csharpminor.Ocmplu, csharpminor.Cle},
		{"gt float", clight.Ogt, ctypes.Float(), csharpminor.Ocmps, csharpminor.Cgt},
		{"ge double", clight.Oge, ctypes.Double(), csharpminor.Ocmpf, csharpminor.Cge},
		{"eq pointer", clight.Oeq, ctypes.Pointer(ctypes.Int()), csharpminor.Ocmplu, csharpminor.Ceq},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOp, gotCmp := TranslateBinaryOp(tt.op, tt.leftType, tt.leftType)
			if gotOp != tt.wantOp {
				t.Errorf("TranslateBinaryOp(%v, %v) op = %v, want %v", tt.op, tt.leftType, gotOp, tt.wantOp)
			}
			if gotCmp != tt.wantCmp {
				t.Errorf("TranslateBinaryOp(%v, %v) cmp = %v, want %v", tt.op, tt.leftType, gotCmp, tt.wantCmp)
			}
		})
	}
}

func TestTranslateCast_IntTruncation(t *testing.T) {
	tests := []struct {
		name string
		to   ctypes.Type
		want csharpminor.UnaryOp
	}{
		{"to signed char", ctypes.Char(), csharpminor.Ocast8signed},
		{"to unsigned char", ctypes.UChar(), csharpminor.Ocast8unsigned},
		{"to signed short", ctypes.Short(), csharpminor.Ocast16signed},
		{"to unsigned short", ctypes.Tint{Size: ctypes.I16, Sign: ctypes.Unsigned}, csharpminor.Ocast16unsigned},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := TranslateCast(ctypes.Int(), tt.to)
			if !ok {
				t.Errorf("TranslateCast(int, %v) returned ok=false", tt.to)
			}
			if got != tt.want {
				t.Errorf("TranslateCast(int, %v) = %v, want %v", tt.to, got, tt.want)
			}
		})
	}
}

func TestTranslateCast_FloatConversion(t *testing.T) {
	tests := []struct {
		name     string
		from, to ctypes.Type
		want     csharpminor.UnaryOp
	}{
		{"double to float", ctypes.Double(), ctypes.Float(), csharpminor.Osingleoffloat},
		{"float to double", ctypes.Float(), ctypes.Double(), csharpminor.Ofloatofsingle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := TranslateCast(tt.from, tt.to)
			if !ok {
				t.Errorf("TranslateCast(%v, %v) returned ok=false", tt.from, tt.to)
			}
			if got != tt.want {
				t.Errorf("TranslateCast(%v, %v) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestTranslateCast_IntFloatConversion(t *testing.T) {
	tests := []struct {
		name     string
		from, to ctypes.Type
		want     csharpminor.UnaryOp
	}{
		{"signed int to double", ctypes.Int(), ctypes.Double(), csharpminor.Ofloatofint},
		{"unsigned int to double", ctypes.UInt(), ctypes.Double(), csharpminor.Ofloatofintu},
		{"double to signed int", ctypes.Double(), ctypes.Int(), csharpminor.Ointoffloat},
		{"double to unsigned int", ctypes.Double(), ctypes.UInt(), csharpminor.Ointuoffloat},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := TranslateCast(tt.from, tt.to)
			if !ok {
				t.Errorf("TranslateCast(%v, %v) returned ok=false", tt.from, tt.to)
			}
			if got != tt.want {
				t.Errorf("TranslateCast(%v, %v) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestTranslateCast_LongFloatConversion(t *testing.T) {
	tests := []struct {
		name     string
		from, to ctypes.Type
		want     csharpminor.UnaryOp
	}{
		{"signed long to double", ctypes.Long(), ctypes.Double(), csharpminor.Ofloatoflong},
		{"unsigned long to double", ctypes.Tlong{Sign: ctypes.Unsigned}, ctypes.Double(), csharpminor.Ofloatoflongu},
		{"signed long to float", ctypes.Long(), ctypes.Float(), csharpminor.Osingleoflong},
		{"unsigned long to float", ctypes.Tlong{Sign: ctypes.Unsigned}, ctypes.Float(), csharpminor.Osingleoflongu},
		{"double to signed long", ctypes.Double(), ctypes.Long(), csharpminor.Olongoffloat},
		{"double to unsigned long", ctypes.Double(), ctypes.Tlong{Sign: ctypes.Unsigned}, csharpminor.Olonguoffloat},
		{"float to signed long", ctypes.Float(), ctypes.Long(), csharpminor.Olongofsingle},
		{"float to unsigned long", ctypes.Float(), ctypes.Tlong{Sign: ctypes.Unsigned}, csharpminor.Olonguofsingle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := TranslateCast(tt.from, tt.to)
			if !ok {
				t.Errorf("TranslateCast(%v, %v) returned ok=false", tt.from, tt.to)
			}
			if got != tt.want {
				t.Errorf("TranslateCast(%v, %v) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestTranslateCast_IntLongConversion(t *testing.T) {
	tests := []struct {
		name     string
		from, to ctypes.Type
		want     csharpminor.UnaryOp
	}{
		{"signed int to long", ctypes.Int(), ctypes.Long(), csharpminor.Olongofint},
		{"unsigned int to long", ctypes.UInt(), ctypes.Long(), csharpminor.Olongofintu},
		{"long to int", ctypes.Long(), ctypes.Int(), csharpminor.Ointoflong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := TranslateCast(tt.from, tt.to)
			if !ok {
				t.Errorf("TranslateCast(%v, %v) returned ok=false", tt.from, tt.to)
			}
			if got != tt.want {
				t.Errorf("TranslateCast(%v, %v) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestTranslateCast_SameType(t *testing.T) {
	types := []ctypes.Type{
		ctypes.Int(),
		ctypes.Long(),
		ctypes.Float(),
		ctypes.Double(),
	}
	for _, typ := range types {
		t.Run(typ.String(), func(t *testing.T) {
			_, ok := TranslateCast(typ, typ)
			if ok {
				t.Errorf("TranslateCast(%v, %v) should return ok=false for same type", typ, typ)
			}
		})
	}
}

func TestIsComparisonOp(t *testing.T) {
	comparisons := []clight.BinaryOp{clight.Oeq, clight.One, clight.Olt, clight.Ogt, clight.Ole, clight.Oge}
	nonComparisons := []clight.BinaryOp{clight.Oadd, clight.Osub, clight.Omul, clight.Odiv, clight.Omod, clight.Oand, clight.Oor, clight.Oxor, clight.Oshl, clight.Oshr}

	for _, op := range comparisons {
		if !IsComparisonOp(op) {
			t.Errorf("IsComparisonOp(%v) = false, want true", op)
		}
	}

	for _, op := range nonComparisons {
		if IsComparisonOp(op) {
			t.Errorf("IsComparisonOp(%v) = true, want false", op)
		}
	}
}
