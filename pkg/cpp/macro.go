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
