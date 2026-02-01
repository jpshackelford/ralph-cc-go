// Package cminorsel - Machine operators for instruction selection.
// This file defines machine-level operators that go beyond the high-level
// operators in Cminor. These include:
// - High-word multiplication (mulhs, mulhu)
// - Absolute value operations
// - Combined shift+arithmetic operations (ARM64)
// - Multiply-accumulate operations (ARM64)
// This mirrors CompCert's backend/Op.v and aarch64/Op.v
package cminorsel

// --- Machine-Level Unary Operators ---
// These extend the Cminor unary operators with target-specific operations.

// MachUnaryOp represents machine-level unary operators
type MachUnaryOp int

const (
	// Integer operations
	MOnegint MachUnaryOp = iota // integer negation
	MOnotint                    // integer bitwise not
	MOnegf                      // float64 negation
	MOnegs                      // float32 negation
	MOabsf                      // float64 absolute value
	MOabss                      // float32 absolute value

	// Long operations
	MOnegl  // long negation
	MOnotl  // long bitwise not
	MOabsfl // float64 abs for long result

	// Cast operations (sign/zero extend)
	MOcast8signed    // sign-extend 8-bit to 32-bit
	MOcast8unsigned  // zero-extend 8-bit to 32-bit
	MOcast16signed   // sign-extend 16-bit to 32-bit
	MOcast16unsigned // zero-extend 16-bit to 32-bit

	// Integer/Float conversions
	MOsingleoffloat // float64 -> float32
	MOfloatofsingle // float32 -> float64
	MOintoffloat    // float64 -> int (signed, round toward zero)
	MOintuoffloat   // float64 -> int (unsigned, round toward zero)
	MOfloatofint    // int -> float64 (signed)
	MOfloatofintu   // int -> float64 (unsigned)

	// Long/Float conversions
	MOlongoffloat  // float64 -> long (signed)
	MOlonguoffloat // float64 -> long (unsigned)
	MOfloatoflong  // long -> float64 (signed)
	MOfloatoflongu // long -> float64 (unsigned)

	// Long/Single conversions
	MOlongofsingle  // float32 -> long (signed)
	MOlonguofsingle // float32 -> long (unsigned)
	MOsingleoflong  // long -> float32 (signed)
	MOsingleoflongu // long -> float32 (unsigned)

	// Int/Long conversions
	MOintoflong  // long -> int (truncate)
	MOlongofint  // int -> long (signed)
	MOlongofintu // int -> long (unsigned)

	// ARM64-specific unary ops
	MOrbit   // reverse bits (rbit instruction)
	MOclz    // count leading zeros
	MOcls    // count leading sign bits
	MOrev    // byte reverse (endian swap)
	MOrev16  // byte reverse within halfwords
	MOsqrtf  // float64 square root
	MOsqrts  // float32 square root
	MOnegfs  // float negate and single conversion
	MOabsfs  // float abs and single conversion
)

func (op MachUnaryOp) String() string {
	names := []string{
		"negint", "notint", "negf", "negs", "absf", "abss",
		"negl", "notl", "absfl",
		"cast8s", "cast8u", "cast16s", "cast16u",
		"singleoffloat", "floatofsingle",
		"intoffloat", "intuoffloat", "floatofint", "floatofintu",
		"longoffloat", "longuoffloat", "floatoflong", "floatoflongu",
		"longofsingle", "longuofsingle", "singleoflong", "singleoflongu",
		"intoflong", "longofint", "longofintu",
		"rbit", "clz", "cls", "rev", "rev16",
		"sqrtf", "sqrts", "negfs", "absfs",
	}
	if int(op) < len(names) {
		return names[op]
	}
	return "?"
}

// --- Machine-Level Binary Operators ---
// These extend the Cminor binary operators with target-specific operations.

// MachBinaryOp represents machine-level binary operators
type MachBinaryOp int

const (
	// Integer arithmetic
	MOadd  MachBinaryOp = iota // int addition
	MOsub                      // int subtraction
	MOmul                      // int multiplication
	MOmulhs                    // int multiplication high (signed) - upper 32 bits
	MOmulhu                    // int multiplication high (unsigned) - upper 32 bits
	MOdiv                      // int signed division
	MOdivu                     // int unsigned division
	MOmod                      // int signed modulo
	MOmodu                     // int unsigned modulo

	// Float64 arithmetic
	MOaddf // float64 addition
	MOsubf // float64 subtraction
	MOmulf // float64 multiplication
	MOdivf // float64 division

	// Float32 arithmetic
	MOadds // float32 addition
	MOsubs // float32 subtraction
	MOmuls // float32 multiplication
	MOdivs // float32 division

	// Long arithmetic
	MOaddl  // long addition
	MOsubl  // long subtraction
	MOmull  // long multiplication
	MOmullhs // long multiplication high (signed)
	MOmullhu // long multiplication high (unsigned)
	MOdivl  // long signed division
	MOdivlu // long unsigned division
	MOmodl  // long signed modulo
	MOmodlu // long unsigned modulo

	// Integer bitwise
	MOand  // int bitwise and
	MOor   // int bitwise or
	MOxor  // int bitwise xor
	MOshl  // int shift left
	MOshr  // int shift right (signed/arithmetic)
	MOshru // int shift right (unsigned/logical)

	// Long bitwise
	MOandl  // long bitwise and
	MOorl   // long bitwise or
	MOxorl  // long bitwise xor
	MOshll  // long shift left
	MOshrl  // long shift right (signed)
	MOshrlu // long shift right (unsigned)

	// Integer comparisons - produce int 0 or 1
	MOcmp  // int signed comparison
	MOcmpu // int unsigned comparison

	// Float comparisons
	MOcmpf // float64 comparison
	MOcmps // float32 comparison

	// Long comparisons
	MOcmpl  // long signed comparison
	MOcmplu // long unsigned comparison

	// ARM64-specific combined operations (shift + arithmetic)
	MOaddshift   // a + (b << shift)
	MOsubshift   // a - (b << shift)
	MOandshift   // a & (b << shift)
	MOorshift    // a | (b << shift)
	MOxorshift   // a ^ (b << shift)
	MOaddlshift  // long: a + (b << shift)
	MOsublshift  // long: a - (b << shift)
	MOandlshift  // long: a & (b << shift)
	MOorlshift   // long: a | (b << shift)
	MOxorlshift  // long: a ^ (b << shift)

	// ARM64 multiply-accumulate
	MOmadd  // a + b * c (fused multiply-add for int)
	MOmsub  // a - b * c (fused multiply-sub for int)
	MOmaddl // long: a + b * c
	MOmsubl // long: a - b * c

	// ARM64 floating-point multiply-accumulate
	MOmaddf // float64: a + b * c (fused)
	MOmsubf // float64: a - b * c (fused)
	MOmadds // float32: a + b * c (fused)
	MOmsubs // float32: a - b * c (fused)
	// Negated forms
	MOnmaddf // float64: -(a + b * c)
	MOnmsubf // float64: -(a - b * c)
	MOnmadds // float32: -(a + b * c)
	MOnmsubs // float32: -(a - b * c)

	// ARM64 bit manipulation
	MObic   // bit clear: a & ~b
	MOorn   // or not: a | ~b
	MOeon   // exclusive or not: a ^ ~b
	MObicl  // long: a & ~b
	MOornl  // long: a | ~b
	MOeonl  // long: a ^ ~b

	// ARM64 conditional select
	MOcsel  // if cond then a else b
	MOcsell // long: if cond then a else b
)

func (op MachBinaryOp) String() string {
	names := []string{
		"add", "sub", "mul", "mulhs", "mulhu", "div", "divu", "mod", "modu",
		"addf", "subf", "mulf", "divf",
		"adds", "subs", "muls", "divs",
		"addl", "subl", "mull", "mullhs", "mullhu", "divl", "divlu", "modl", "modlu",
		"and", "or", "xor", "shl", "shr", "shru",
		"andl", "orl", "xorl", "shll", "shrl", "shrlu",
		"cmp", "cmpu", "cmpf", "cmps", "cmpl", "cmplu",
		"addshift", "subshift", "andshift", "orshift", "xorshift",
		"addlshift", "sublshift", "andlshift", "orlshift", "xorlshift",
		"madd", "msub", "maddl", "msubl",
		"maddf", "msubf", "madds", "msubs",
		"nmaddf", "nmsubf", "nmadds", "nmsubs",
		"bic", "orn", "eon", "bicl", "ornl", "eonl",
		"csel", "csell",
	}
	if int(op) < len(names) {
		return names[op]
	}
	return "?"
}

// --- Machine-Level Ternary Operators ---
// Operations that take three operands (e.g., multiply-accumulate)

// MachTernaryOp represents machine-level ternary operators
type MachTernaryOp int

const (
	// Multiply-accumulate: result = arg1 +/- arg2 * arg3
	MOTmadd  MachTernaryOp = iota // int: a + b * c
	MOTmsub                       // int: a - b * c
	MOTmaddl                      // long: a + b * c
	MOTmsubl                      // long: a - b * c

	// Floating-point multiply-accumulate
	MOTmaddf // float64: a + b * c
	MOTmsubf // float64: a - b * c
	MOTmadds // float32: a + b * c
	MOTmsubs // float32: a - b * c
)

func (op MachTernaryOp) String() string {
	names := []string{
		"madd", "msub", "maddl", "msubl",
		"maddf", "msubf", "madds", "msubs",
	}
	if int(op) < len(names) {
		return names[op]
	}
	return "?"
}

// --- Helper functions ---

// IsCommutative returns true if the binary operation is commutative
func (op MachBinaryOp) IsCommutative() bool {
	switch op {
	case MOadd, MOmul, MOaddf, MOmulf, MOadds, MOmuls,
		MOaddl, MOmull, MOand, MOor, MOxor, MOandl, MOorl, MOxorl:
		return true
	}
	return false
}

// IsCompare returns true if the operation is a comparison
func (op MachBinaryOp) IsCompare() bool {
	switch op {
	case MOcmp, MOcmpu, MOcmpf, MOcmps, MOcmpl, MOcmplu:
		return true
	}
	return false
}

// IsShiftCombined returns true if this is a combined shift+op
func (op MachBinaryOp) IsShiftCombined() bool {
	switch op {
	case MOaddshift, MOsubshift, MOandshift, MOorshift, MOxorshift,
		MOaddlshift, MOsublshift, MOandlshift, MOorlshift, MOxorlshift:
		return true
	}
	return false
}

// IsFusedMultiply returns true if this is a fused multiply-accumulate
func (op MachBinaryOp) IsFusedMultiply() bool {
	switch op {
	case MOmadd, MOmsub, MOmaddl, MOmsubl,
		MOmaddf, MOmsubf, MOmadds, MOmsubs,
		MOnmaddf, MOnmsubf, MOnmadds, MOnmsubs:
		return true
	}
	return false
}
