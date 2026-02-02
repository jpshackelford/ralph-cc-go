package cpp

import (
	"strings"
	"testing"
)

func TestMacroTableBasics(t *testing.T) {
	mt := NewMacroTable()

	// Test built-in macros exist
	if !mt.IsDefined("__FILE__") {
		t.Error("__FILE__ should be defined")
	}
	if !mt.IsDefined("__LINE__") {
		t.Error("__LINE__ should be defined")
	}
	if !mt.IsDefined("__DATE__") {
		t.Error("__DATE__ should be defined")
	}
	if !mt.IsDefined("__TIME__") {
		t.Error("__TIME__ should be defined")
	}
	if !mt.IsDefined("__STDC__") {
		t.Error("__STDC__ should be defined")
	}
	if !mt.IsDefined("__STDC_VERSION__") {
		t.Error("__STDC_VERSION__ should be defined")
	}
}

func TestDefineObjectMacro(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 1, Column: 1}

	// Define a simple object macro
	replacement := []Token{
		{Type: PP_NUMBER, Text: "42", Loc: loc},
	}
	err := mt.DefineObject("FOO", replacement, loc)
	if err != nil {
		t.Fatalf("DefineObject failed: %v", err)
	}

	// Lookup the macro
	m := mt.Lookup("FOO")
	if m == nil {
		t.Fatal("Lookup failed")
	}
	if m.Name != "FOO" {
		t.Errorf("Name = %q, want %q", m.Name, "FOO")
	}
	if m.Kind != MacroObject {
		t.Errorf("Kind = %v, want MacroObject", m.Kind)
	}
	if len(m.Replacement) != 1 || m.Replacement[0].Text != "42" {
		t.Errorf("Replacement = %v, want [{42}]", m.Replacement)
	}
}

func TestDefineFunctionMacro(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 1, Column: 1}

	// Define a function-like macro: #define MAX(a, b) ((a)>(b)?(a):(b))
	replacement := []Token{
		{Type: PP_PUNCTUATOR, Text: "(", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: "(", Loc: loc},
		{Type: PP_IDENTIFIER, Text: "a", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: ")", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: ">", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: "(", Loc: loc},
		{Type: PP_IDENTIFIER, Text: "b", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: ")", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: "?", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: "(", Loc: loc},
		{Type: PP_IDENTIFIER, Text: "a", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: ")", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: ":", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: "(", Loc: loc},
		{Type: PP_IDENTIFIER, Text: "b", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: ")", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: ")", Loc: loc},
	}
	err := mt.DefineFunction("MAX", []string{"a", "b"}, false, replacement, loc)
	if err != nil {
		t.Fatalf("DefineFunction failed: %v", err)
	}

	m := mt.Lookup("MAX")
	if m == nil {
		t.Fatal("Lookup failed")
	}
	if m.Kind != MacroFunction {
		t.Errorf("Kind = %v, want MacroFunction", m.Kind)
	}
	if len(m.Params) != 2 || m.Params[0] != "a" || m.Params[1] != "b" {
		t.Errorf("Params = %v, want [a, b]", m.Params)
	}
	if m.IsVariadic {
		t.Error("IsVariadic should be false")
	}
}

func TestDefineVariadicMacro(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 1, Column: 1}

	// Define variadic macro: #define PRINTF(fmt, ...) printf(fmt, __VA_ARGS__)
	replacement := []Token{
		{Type: PP_IDENTIFIER, Text: "printf", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: "(", Loc: loc},
		{Type: PP_IDENTIFIER, Text: "fmt", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: ",", Loc: loc},
		{Type: PP_IDENTIFIER, Text: "__VA_ARGS__", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: ")", Loc: loc},
	}
	err := mt.DefineFunction("PRINTF", []string{"fmt"}, true, replacement, loc)
	if err != nil {
		t.Fatalf("DefineFunction failed: %v", err)
	}

	m := mt.Lookup("PRINTF")
	if m == nil {
		t.Fatal("Lookup failed")
	}
	if !m.IsVariadic {
		t.Error("IsVariadic should be true")
	}
}

func TestDefineSimple(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "<cmdline>", Line: 1, Column: 1}

	// Simple define with value
	err := mt.DefineSimple("VERSION", "123", loc)
	if err != nil {
		t.Fatalf("DefineSimple failed: %v", err)
	}

	m := mt.Lookup("VERSION")
	if m == nil {
		t.Fatal("Lookup failed")
	}
	if len(m.Replacement) != 1 || m.Replacement[0].Text != "123" {
		t.Errorf("Replacement = %v, want [{123}]", m.Replacement)
	}

	// Simple define without value (like -DDEBUG with value "1")
	err = mt.DefineSimple("DEBUG", "1", loc)
	if err != nil {
		t.Fatalf("DefineSimple failed: %v", err)
	}

	m = mt.Lookup("DEBUG")
	if m == nil {
		t.Fatal("Lookup failed")
	}
	if len(m.Replacement) != 1 || m.Replacement[0].Text != "1" {
		t.Errorf("Replacement = %v, want [{1}]", m.Replacement)
	}
}

func TestUndefine(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 1, Column: 1}

	// Define and then undefine
	err := mt.DefineSimple("FOO", "42", loc)
	if err != nil {
		t.Fatalf("DefineSimple failed: %v", err)
	}
	if !mt.IsDefined("FOO") {
		t.Error("FOO should be defined")
	}

	mt.Undefine("FOO")
	if mt.IsDefined("FOO") {
		t.Error("FOO should be undefined after Undefine")
	}

	// Undefining non-existent macro should not panic
	mt.Undefine("NONEXISTENT")
}

func TestUndefineBuiltins(t *testing.T) {
	mt := NewMacroTable()

	// __FILE__ and __LINE__ cannot be undefined
	mt.Undefine("__FILE__")
	if !mt.IsDefined("__FILE__") {
		t.Error("__FILE__ should not be undefine-able")
	}

	mt.Undefine("__LINE__")
	if !mt.IsDefined("__LINE__") {
		t.Error("__LINE__ should not be undefine-able")
	}

	// Other built-ins can be undefined
	mt.Undefine("__STDC__")
	if mt.IsDefined("__STDC__") {
		t.Error("__STDC__ should be undefine-able")
	}
}

func TestRedefinitionIdentical(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 1, Column: 1}

	// Define a macro
	replacement := []Token{
		{Type: PP_NUMBER, Text: "42", Loc: loc},
	}
	err := mt.DefineObject("FOO", replacement, loc)
	if err != nil {
		t.Fatalf("First DefineObject failed: %v", err)
	}

	// Redefine with identical definition - should be OK
	err = mt.DefineObject("FOO", replacement, loc)
	if err != nil {
		t.Errorf("Identical redefinition should be OK, got: %v", err)
	}
}

func TestRedefinitionDifferent(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 1, Column: 1}

	// Define a macro
	replacement1 := []Token{
		{Type: PP_NUMBER, Text: "42", Loc: loc},
	}
	err := mt.DefineObject("FOO", replacement1, loc)
	if err != nil {
		t.Fatalf("First DefineObject failed: %v", err)
	}

	// Redefine with different definition - should error
	replacement2 := []Token{
		{Type: PP_NUMBER, Text: "100", Loc: loc},
	}
	err = mt.DefineObject("FOO", replacement2, loc)
	if err == nil {
		t.Error("Different redefinition should fail")
	}
}

func TestBuiltinMacroExpansion(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 42, Column: 1}

	// Test __STDC__
	m := mt.Lookup("__STDC__")
	if m == nil || m.BuiltinFunc == nil {
		t.Fatal("__STDC__ should have BuiltinFunc")
	}
	tokens := m.BuiltinFunc(loc)
	if len(tokens) != 1 || tokens[0].Text != "1" {
		t.Errorf("__STDC__ = %v, want [{1}]", tokens)
	}

	// Test __STDC_VERSION__
	m = mt.Lookup("__STDC_VERSION__")
	tokens = m.BuiltinFunc(loc)
	if len(tokens) != 1 || tokens[0].Text != "201112L" {
		t.Errorf("__STDC_VERSION__ = %v, want [{201112L}]", tokens)
	}

	// Test __DATE__
	m = mt.Lookup("__DATE__")
	tokens = m.BuiltinFunc(loc)
	if len(tokens) != 1 || tokens[0].Type != PP_STRING {
		t.Errorf("__DATE__ should return a string token")
	}

	// Test __TIME__
	m = mt.Lookup("__TIME__")
	tokens = m.BuiltinFunc(loc)
	if len(tokens) != 1 || tokens[0].Type != PP_STRING {
		t.Errorf("__TIME__ should return a string token")
	}
}

func TestGetFileAndLineTokens(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "myfile.c", Line: 123, Column: 1}

	// Test GetFileToken
	tokens := mt.GetFileToken(loc)
	if len(tokens) != 1 || tokens[0].Text != "\"myfile.c\"" {
		t.Errorf("GetFileToken = %v, want [\"myfile.c\"]", tokens)
	}

	// Test GetLineToken
	tokens = mt.GetLineToken(loc)
	if len(tokens) != 1 || tokens[0].Text != "123" {
		t.Errorf("GetLineToken = %v, want [123]", tokens)
	}
}

func TestApplyCmdlineDefines(t *testing.T) {
	mt := NewMacroTable()

	defines := []string{
		"DEBUG",       // No value, defaults to 1
		"VERSION=123", // With value
		"NAME=hello",
	}
	undefines := []string{
		"__STDC_HOSTED__",
	}

	err := mt.ApplyCmdlineDefines(defines, undefines)
	if err != nil {
		t.Fatalf("ApplyCmdlineDefines failed: %v", err)
	}

	// Check DEBUG=1
	m := mt.Lookup("DEBUG")
	if m == nil || len(m.Replacement) != 1 || m.Replacement[0].Text != "1" {
		t.Errorf("DEBUG should be defined as 1")
	}

	// Check VERSION=123
	m = mt.Lookup("VERSION")
	if m == nil || len(m.Replacement) != 1 || m.Replacement[0].Text != "123" {
		t.Errorf("VERSION should be defined as 123")
	}

	// Check NAME=hello
	m = mt.Lookup("NAME")
	if m == nil || len(m.Replacement) != 1 || m.Replacement[0].Text != "hello" {
		t.Errorf("NAME should be defined as hello")
	}

	// Check __STDC_HOSTED__ is undefined
	if mt.IsDefined("__STDC_HOSTED__") {
		t.Error("__STDC_HOSTED__ should be undefined")
	}
}

func TestClone(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 1, Column: 1}

	err := mt.DefineSimple("FOO", "42", loc)
	if err != nil {
		t.Fatalf("DefineSimple failed: %v", err)
	}

	// Clone the table
	cloned := mt.Clone()

	// Original and clone both have FOO
	if !cloned.IsDefined("FOO") {
		t.Error("Clone should have FOO")
	}

	// Modify clone
	cloned.Undefine("FOO")

	// Original still has FOO
	if !mt.IsDefined("FOO") {
		t.Error("Original should still have FOO after modifying clone")
	}
}

func TestDefineFromDirective(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 1, Column: 1}

	// Test object-like macro
	dir := &Directive{
		Type:      DIR_DEFINE,
		Loc:       loc,
		MacroName: "PI",
		MacroBody: []Token{
			{Type: PP_NUMBER, Text: "3.14159", Loc: loc},
		},
	}
	err := mt.DefineFromDirective(dir)
	if err != nil {
		t.Fatalf("DefineFromDirective failed: %v", err)
	}

	m := mt.Lookup("PI")
	if m == nil || m.Kind != MacroObject {
		t.Error("PI should be object-like macro")
	}

	// Test function-like macro
	dir = &Directive{
		Type:        DIR_DEFINE,
		Loc:         loc,
		MacroName:   "SQUARE",
		MacroParams: []string{"x"},
		MacroBody: []Token{
			{Type: PP_PUNCTUATOR, Text: "(", Loc: loc},
			{Type: PP_IDENTIFIER, Text: "x", Loc: loc},
			{Type: PP_PUNCTUATOR, Text: "*", Loc: loc},
			{Type: PP_IDENTIFIER, Text: "x", Loc: loc},
			{Type: PP_PUNCTUATOR, Text: ")", Loc: loc},
		},
	}
	err = mt.DefineFromDirective(dir)
	if err != nil {
		t.Fatalf("DefineFromDirective failed: %v", err)
	}

	m = mt.Lookup("SQUARE")
	if m == nil || m.Kind != MacroFunction {
		t.Error("SQUARE should be function-like macro")
	}
}

func TestMacroString(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 1, Column: 1}

	// Object macro
	mt.DefineObject("FOO", []Token{
		{Type: PP_NUMBER, Text: "42", Loc: loc},
	}, loc)
	m := mt.Lookup("FOO")
	s := m.String()
	if !strings.Contains(s, "#define FOO 42") {
		t.Errorf("String() = %q, want contains '#define FOO 42'", s)
	}

	// Function macro
	mt.DefineFunction("MAX", []string{"a", "b"}, false, []Token{
		{Type: PP_IDENTIFIER, Text: "a", Loc: loc},
		{Type: PP_PUNCTUATOR, Text: "+", Loc: loc},
		{Type: PP_IDENTIFIER, Text: "b", Loc: loc},
	}, loc)
	m = mt.Lookup("MAX")
	s = m.String()
	if !strings.Contains(s, "#define MAX(a, b)") {
		t.Errorf("String() = %q, want contains '#define MAX(a, b)'", s)
	}

	// Built-in macro
	m = mt.Lookup("__STDC__")
	s = m.String()
	if !strings.Contains(s, "builtin") {
		t.Errorf("String() = %q, want contains 'builtin'", s)
	}
}

func TestIsFunctionMacroIsObjectMacro(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 1, Column: 1}

	mt.DefineObject("OBJ", []Token{}, loc)
	mt.DefineFunction("FUNC", []string{}, false, []Token{}, loc)

	if !mt.IsObjectMacro("OBJ") {
		t.Error("OBJ should be object macro")
	}
	if mt.IsFunctionMacro("OBJ") {
		t.Error("OBJ should not be function macro")
	}

	if !mt.IsFunctionMacro("FUNC") {
		t.Error("FUNC should be function macro")
	}
	if mt.IsObjectMacro("FUNC") {
		t.Error("FUNC should not be object macro")
	}
}

func TestNames(t *testing.T) {
	mt := NewMacroTable()
	loc := SourceLoc{File: "test.c", Line: 1, Column: 1}

	mt.DefineSimple("FOO", "1", loc)
	mt.DefineSimple("BAR", "2", loc)

	names := mt.Names()
	// Should include built-ins + FOO + BAR
	if len(names) < 2 {
		t.Errorf("Names() = %v, want at least FOO and BAR", names)
	}

	// Check FOO and BAR are in the list
	hasFoo := false
	hasBar := false
	for _, n := range names {
		if n == "FOO" {
			hasFoo = true
		}
		if n == "BAR" {
			hasBar = true
		}
	}
	if !hasFoo || !hasBar {
		t.Errorf("Names() should contain FOO and BAR, got %v", names)
	}
}
