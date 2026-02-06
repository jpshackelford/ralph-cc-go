// macro.go implements macro definition storage and lookup.
package cpp

import (
	"fmt"
	"time"
)

// MacroKind indicates whether a macro is object-like or function-like.
type MacroKind int

const (
	MacroObject   MacroKind = iota // Object-like: #define NAME value
	MacroFunction                  // Function-like: #define NAME(a,b) value
	MacroBuiltin                   // Built-in macro (__FILE__, __LINE__, etc.)
)

// Macro represents a preprocessor macro definition.
type Macro struct {
	Name         string    // Macro name
	Kind         MacroKind // Object, function, or built-in
	Params       []string  // Parameters (for function-like macros)
	IsVariadic   bool      // True if last param is ... (or macro ends with ...)
	Replacement  []Token   // Replacement token list
	Loc          SourceLoc // Where the macro was defined
	BuiltinFunc  func(loc SourceLoc) []Token // For built-in macros
}

// MacroTable stores macro definitions and provides lookup.
type MacroTable struct {
	macros    map[string]*Macro
	compDate  string // Cached compilation date for __DATE__
	compTime  string // Cached compilation time for __TIME__
}

// NewMacroTable creates a new macro table with built-in macros.
func NewMacroTable() *MacroTable {
	now := time.Now()
	mt := &MacroTable{
		macros:   make(map[string]*Macro),
		compDate: now.Format("Jan _2 2006"),
		compTime: now.Format("15:04:05"),
	}
	mt.initBuiltins()
	return mt
}

// initBuiltins registers the standard built-in macros.
func (mt *MacroTable) initBuiltins() {
	// __FILE__ - current filename (handled dynamically during expansion)
	mt.macros["__FILE__"] = &Macro{
		Name:        "__FILE__",
		Kind:        MacroBuiltin,
		BuiltinFunc: nil, // Set during expansion with current context
	}

	// __LINE__ - current line number (handled dynamically during expansion)
	mt.macros["__LINE__"] = &Macro{
		Name:        "__LINE__",
		Kind:        MacroBuiltin,
		BuiltinFunc: nil, // Set during expansion with current context
	}

	// __DATE__ - compilation date
	mt.macros["__DATE__"] = &Macro{
		Name: "__DATE__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_STRING, Text: fmt.Sprintf("\"%s\"", mt.compDate), Loc: loc}}
		},
	}

	// __TIME__ - compilation time
	mt.macros["__TIME__"] = &Macro{
		Name: "__TIME__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_STRING, Text: fmt.Sprintf("\"%s\"", mt.compTime), Loc: loc}}
		},
	}

	// __STDC__ - always 1
	mt.macros["__STDC__"] = &Macro{
		Name: "__STDC__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1", Loc: loc}}
		},
	}

	// __STDC_VERSION__ - 201112L for C11
	mt.macros["__STDC_VERSION__"] = &Macro{
		Name: "__STDC_VERSION__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "201112L", Loc: loc}}
		},
	}

	// __STDC_HOSTED__ - 1 for hosted implementation
	mt.macros["__STDC_HOSTED__"] = &Macro{
		Name: "__STDC_HOSTED__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1", Loc: loc}}
		},
	}

	// GCC compatibility macros - pretend to be GCC 4.2 (widely compatible)
	mt.macros["__GNUC__"] = &Macro{
		Name: "__GNUC__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "4", Loc: loc}}
		},
	}
	mt.macros["__GNUC_MINOR__"] = &Macro{
		Name: "__GNUC_MINOR__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "2", Loc: loc}}
		},
	}
	mt.macros["__GNUC_PATCHLEVEL__"] = &Macro{
		Name: "__GNUC_PATCHLEVEL__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1", Loc: loc}}
		},
	}
	mt.macros["__GNUC_STDC_INLINE__"] = &Macro{
		Name: "__GNUC_STDC_INLINE__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1", Loc: loc}}
		},
	}

	// Type size macros for ARM64
	mt.macros["__SIZEOF_INT__"] = &Macro{
		Name: "__SIZEOF_INT__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "4", Loc: loc}}
		},
	}
	mt.macros["__SIZEOF_LONG__"] = &Macro{
		Name: "__SIZEOF_LONG__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "8", Loc: loc}}
		},
	}
	mt.macros["__SIZEOF_LONG_LONG__"] = &Macro{
		Name: "__SIZEOF_LONG_LONG__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "8", Loc: loc}}
		},
	}
	mt.macros["__SIZEOF_SHORT__"] = &Macro{
		Name: "__SIZEOF_SHORT__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "2", Loc: loc}}
		},
	}
	mt.macros["__SIZEOF_POINTER__"] = &Macro{
		Name: "__SIZEOF_POINTER__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "8", Loc: loc}}
		},
	}
	mt.macros["__SIZEOF_SIZE_T__"] = &Macro{
		Name: "__SIZEOF_SIZE_T__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "8", Loc: loc}}
		},
	}
	mt.macros["__SIZEOF_PTRDIFF_T__"] = &Macro{
		Name: "__SIZEOF_PTRDIFF_T__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "8", Loc: loc}}
		},
	}
	mt.macros["__SIZEOF_FLOAT__"] = &Macro{
		Name: "__SIZEOF_FLOAT__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "4", Loc: loc}}
		},
	}
	mt.macros["__SIZEOF_DOUBLE__"] = &Macro{
		Name: "__SIZEOF_DOUBLE__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "8", Loc: loc}}
		},
	}

	// Byte order macros for ARM64 (little endian)
	mt.macros["__BYTE_ORDER__"] = &Macro{
		Name: "__BYTE_ORDER__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1234", Loc: loc}} // Little endian
		},
	}
	mt.macros["__ORDER_LITTLE_ENDIAN__"] = &Macro{
		Name: "__ORDER_LITTLE_ENDIAN__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1234", Loc: loc}}
		},
	}
	mt.macros["__ORDER_BIG_ENDIAN__"] = &Macro{
		Name: "__ORDER_BIG_ENDIAN__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "4321", Loc: loc}}
		},
	}
	mt.macros["__LITTLE_ENDIAN__"] = &Macro{
		Name: "__LITTLE_ENDIAN__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1", Loc: loc}}
		},
	}

	// Platform macros for ARM64 macOS
	mt.macros["__LP64__"] = &Macro{
		Name: "__LP64__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1", Loc: loc}}
		},
	}
	mt.macros["__aarch64__"] = &Macro{
		Name: "__aarch64__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1", Loc: loc}}
		},
	}
	mt.macros["__arm64__"] = &Macro{
		Name: "__arm64__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1", Loc: loc}}
		},
	}
	mt.macros["__APPLE__"] = &Macro{
		Name: "__APPLE__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1", Loc: loc}}
		},
	}
	mt.macros["__MACH__"] = &Macro{
		Name: "__MACH__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "1", Loc: loc}}
		},
	}
	// Apple compiler compatibility
	mt.macros["__APPLE_CC__"] = &Macro{
		Name: "__APPLE_CC__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "6000", Loc: loc}}
		},
	}

	// Type limits macros
	mt.macros["__CHAR_BIT__"] = &Macro{
		Name: "__CHAR_BIT__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "8", Loc: loc}}
		},
	}
	mt.macros["__SCHAR_MAX__"] = &Macro{
		Name: "__SCHAR_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "127", Loc: loc}}
		},
	}
	mt.macros["__SHRT_MAX__"] = &Macro{
		Name: "__SHRT_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "32767", Loc: loc}}
		},
	}
	mt.macros["__INT_MAX__"] = &Macro{
		Name: "__INT_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "2147483647", Loc: loc}}
		},
	}
	mt.macros["__LONG_MAX__"] = &Macro{
		Name: "__LONG_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "9223372036854775807L", Loc: loc}}
		},
	}
	mt.macros["__LONG_LONG_MAX__"] = &Macro{
		Name: "__LONG_LONG_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "9223372036854775807LL", Loc: loc}}
		},
	}
	mt.macros["__WCHAR_MAX__"] = &Macro{
		Name: "__WCHAR_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "2147483647", Loc: loc}}
		},
	}
	mt.macros["__WCHAR_WIDTH__"] = &Macro{
		Name: "__WCHAR_WIDTH__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "32", Loc: loc}}
		},
	}
	mt.macros["__WINT_MAX__"] = &Macro{
		Name: "__WINT_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "2147483647", Loc: loc}}
		},
	}
	mt.macros["__WINT_WIDTH__"] = &Macro{
		Name: "__WINT_WIDTH__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "32", Loc: loc}}
		},
	}
	mt.macros["__INTMAX_MAX__"] = &Macro{
		Name: "__INTMAX_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "9223372036854775807L", Loc: loc}}
		},
	}
	mt.macros["__INTMAX_WIDTH__"] = &Macro{
		Name: "__INTMAX_WIDTH__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "64", Loc: loc}}
		},
	}
	mt.macros["__SIZE_MAX__"] = &Macro{
		Name: "__SIZE_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "18446744073709551615UL", Loc: loc}}
		},
	}
	mt.macros["__SIZE_WIDTH__"] = &Macro{
		Name: "__SIZE_WIDTH__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "64", Loc: loc}}
		},
	}
	mt.macros["__PTRDIFF_MAX__"] = &Macro{
		Name: "__PTRDIFF_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "9223372036854775807L", Loc: loc}}
		},
	}
	mt.macros["__PTRDIFF_WIDTH__"] = &Macro{
		Name: "__PTRDIFF_WIDTH__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "64", Loc: loc}}
		},
	}
	mt.macros["__INTPTR_MAX__"] = &Macro{
		Name: "__INTPTR_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "9223372036854775807L", Loc: loc}}
		},
	}
	mt.macros["__INTPTR_WIDTH__"] = &Macro{
		Name: "__INTPTR_WIDTH__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "64", Loc: loc}}
		},
	}
	mt.macros["__UINTPTR_MAX__"] = &Macro{
		Name: "__UINTPTR_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "18446744073709551615UL", Loc: loc}}
		},
	}

	// INT8/16/32/64 types
	mt.macros["__INT8_MAX__"] = &Macro{
		Name: "__INT8_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "127", Loc: loc}}
		},
	}
	mt.macros["__INT16_MAX__"] = &Macro{
		Name: "__INT16_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "32767", Loc: loc}}
		},
	}
	mt.macros["__INT32_MAX__"] = &Macro{
		Name: "__INT32_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "2147483647", Loc: loc}}
		},
	}
	mt.macros["__INT64_MAX__"] = &Macro{
		Name: "__INT64_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "9223372036854775807LL", Loc: loc}}
		},
	}
	mt.macros["__UINT8_MAX__"] = &Macro{
		Name: "__UINT8_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "255", Loc: loc}}
		},
	}
	mt.macros["__UINT16_MAX__"] = &Macro{
		Name: "__UINT16_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "65535", Loc: loc}}
		},
	}
	mt.macros["__UINT32_MAX__"] = &Macro{
		Name: "__UINT32_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "4294967295U", Loc: loc}}
		},
	}
	mt.macros["__UINT64_MAX__"] = &Macro{
		Name: "__UINT64_MAX__",
		Kind: MacroBuiltin,
		BuiltinFunc: func(loc SourceLoc) []Token {
			return []Token{{Type: PP_NUMBER, Text: "18446744073709551615ULL", Loc: loc}}
		},
	}

	// Note: We don't define __INTN_MAX, __INTN_MIN, __UINTN_MAX, __UINTN_C, __INTN_C
	// because they are defined differently by different system headers (clang vs gcc).
	// The headers will define them when needed.
}

// Define adds or replaces a macro in the table.
// Returns an error if the macro is being redefined with a different definition.
func (mt *MacroTable) Define(m *Macro) error {
	existing, ok := mt.macros[m.Name]
	if ok && existing.Kind != MacroBuiltin {
		// Check if redefinition is identical (per C standard)
		if !mt.macrosEqual(existing, m) {
			return fmt.Errorf("macro '%s' redefined with different definition", m.Name)
		}
	}
	mt.macros[m.Name] = m
	return nil
}

// DefineObject creates and defines an object-like macro.
func (mt *MacroTable) DefineObject(name string, replacement []Token, loc SourceLoc) error {
	return mt.Define(&Macro{
		Name:        name,
		Kind:        MacroObject,
		Replacement: replacement,
		Loc:         loc,
	})
}

// DefineFunction creates and defines a function-like macro.
func (mt *MacroTable) DefineFunction(name string, params []string, isVariadic bool, replacement []Token, loc SourceLoc) error {
	return mt.Define(&Macro{
		Name:        name,
		Kind:        MacroFunction,
		Params:      params,
		IsVariadic:  isVariadic,
		Replacement: replacement,
		Loc:         loc,
	})
}

// DefineSimple creates and defines a simple object-like macro with text value.
// Useful for -D command line definitions.
func (mt *MacroTable) DefineSimple(name, value string, loc SourceLoc) error {
	var replacement []Token
	if value != "" {
		// Tokenize the value
		lex := NewLexer(value, "<cmdline>")
		for {
			tok := lex.NextToken()
			if tok.Type == PP_EOF || tok.Type == PP_NEWLINE {
				break
			}
			if tok.Type != PP_WHITESPACE {
				replacement = append(replacement, tok)
			}
		}
	}
	return mt.DefineObject(name, replacement, loc)
}

// DefineFromDirective creates a macro from a parsed #define directive.
func (mt *MacroTable) DefineFromDirective(dir *Directive) error {
	if dir.Type != DIR_DEFINE {
		return fmt.Errorf("not a #define directive")
	}

	if dir.MacroParams != nil {
		return mt.DefineFunction(dir.MacroName, dir.MacroParams, dir.IsVariadic, dir.MacroBody, dir.Loc)
	}
	return mt.DefineObject(dir.MacroName, dir.MacroBody, dir.Loc)
}

// Undefine removes a macro from the table.
func (mt *MacroTable) Undefine(name string) {
	// Can't undefine certain built-ins
	if name == "__FILE__" || name == "__LINE__" {
		return
	}
	delete(mt.macros, name)
}

// Lookup returns the macro with the given name, or nil if not found.
func (mt *MacroTable) Lookup(name string) *Macro {
	return mt.macros[name]
}

// IsDefined returns true if a macro with the given name is defined.
func (mt *MacroTable) IsDefined(name string) bool {
	_, ok := mt.macros[name]
	return ok
}

// Clone creates a copy of the macro table.
func (mt *MacroTable) Clone() *MacroTable {
	newMt := &MacroTable{
		macros:   make(map[string]*Macro),
		compDate: mt.compDate,
		compTime: mt.compTime,
	}
	for name, m := range mt.macros {
		newMt.macros[name] = m
	}
	return newMt
}

// Names returns a list of all defined macro names.
func (mt *MacroTable) Names() []string {
	names := make([]string, 0, len(mt.macros))
	for name := range mt.macros {
		names = append(names, name)
	}
	return names
}

// macrosEqual checks if two macros have identical definitions.
func (mt *MacroTable) macrosEqual(a, b *Macro) bool {
	if a.Kind != b.Kind {
		return false
	}
	if a.IsVariadic != b.IsVariadic {
		return false
	}
	if len(a.Params) != len(b.Params) {
		return false
	}
	for i := range a.Params {
		if a.Params[i] != b.Params[i] {
			return false
		}
	}
	if len(a.Replacement) != len(b.Replacement) {
		return false
	}
	for i := range a.Replacement {
		if !tokensEqual(a.Replacement[i], b.Replacement[i]) {
			return false
		}
	}
	return true
}

// tokensEqual checks if two tokens are equal (ignoring location).
func tokensEqual(a, b Token) bool {
	return a.Type == b.Type && a.Text == b.Text
}

// GetFileToken returns the __FILE__ expansion for the given location.
func (mt *MacroTable) GetFileToken(loc SourceLoc) []Token {
	return []Token{{Type: PP_STRING, Text: fmt.Sprintf("\"%s\"", loc.File), Loc: loc}}
}

// GetLineToken returns the __LINE__ expansion for the given location.
func (mt *MacroTable) GetLineToken(loc SourceLoc) []Token {
	return []Token{{Type: PP_NUMBER, Text: fmt.Sprintf("%d", loc.Line), Loc: loc}}
}

// ApplyCmdlineDefines processes -D and -U command line options.
// Format: "NAME" or "NAME=VALUE" for defines, "NAME" for undefines.
func (mt *MacroTable) ApplyCmdlineDefines(defines, undefines []string) error {
	cmdlineLoc := SourceLoc{File: "<command-line>", Line: 1, Column: 1}

	// Process defines
	for _, def := range defines {
		name, value := parseCmdlineDefine(def)
		if err := mt.DefineSimple(name, value, cmdlineLoc); err != nil {
			return fmt.Errorf("processing -D %s: %w", def, err)
		}
	}

	// Process undefines
	for _, name := range undefines {
		mt.Undefine(name)
	}

	return nil
}

// parseCmdlineDefine parses a -D option value into name and value parts.
func parseCmdlineDefine(s string) (name, value string) {
	for i, c := range s {
		if c == '=' {
			return s[:i], s[i+1:]
		}
	}
	// No '=' means define with value "1"
	return s, "1"
}

// IsFunctionMacro returns true if the named macro is function-like.
func (mt *MacroTable) IsFunctionMacro(name string) bool {
	m := mt.Lookup(name)
	return m != nil && m.Kind == MacroFunction
}

// IsObjectMacro returns true if the named macro is object-like.
func (mt *MacroTable) IsObjectMacro(name string) bool {
	m := mt.Lookup(name)
	return m != nil && m.Kind == MacroObject
}

// String returns a debug string representation of a macro.
func (m *Macro) String() string {
	switch m.Kind {
	case MacroBuiltin:
		return fmt.Sprintf("builtin(%s)", m.Name)
	case MacroObject:
		return fmt.Sprintf("#define %s %s", m.Name, TokensToString(m.Replacement))
	case MacroFunction:
		params := ""
		for i, p := range m.Params {
			if i > 0 {
				params += ", "
			}
			params += p
		}
		if m.IsVariadic {
			if params != "" {
				params += ", "
			}
			params += "..."
		}
		return fmt.Sprintf("#define %s(%s) %s", m.Name, params, TokensToString(m.Replacement))
	}
	return fmt.Sprintf("unknown macro %s", m.Name)
}
