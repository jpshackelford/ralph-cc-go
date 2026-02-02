// directive.go implements preprocessing directive parsing.
package cpp

import (
	"fmt"
	"strings"
)

// DirectiveType represents the type of a preprocessing directive.
type DirectiveType int

const (
	DIR_INCLUDE DirectiveType = iota
	DIR_DEFINE
	DIR_UNDEF
	DIR_IF
	DIR_IFDEF
	DIR_IFNDEF
	DIR_ELIF
	DIR_ELSE
	DIR_ENDIF
	DIR_LINE
	DIR_ERROR
	DIR_WARNING
	DIR_PRAGMA
	DIR_LINEMARKER // GCC line marker: # number "filename" [flags]
	DIR_EMPTY      // empty directive (just #)
)

func (d DirectiveType) String() string {
	switch d {
	case DIR_INCLUDE:
		return "include"
	case DIR_DEFINE:
		return "define"
	case DIR_UNDEF:
		return "undef"
	case DIR_IF:
		return "if"
	case DIR_IFDEF:
		return "ifdef"
	case DIR_IFNDEF:
		return "ifndef"
	case DIR_ELIF:
		return "elif"
	case DIR_ELSE:
		return "else"
	case DIR_ENDIF:
		return "endif"
	case DIR_LINE:
		return "line"
	case DIR_ERROR:
		return "error"
	case DIR_WARNING:
		return "warning"
	case DIR_PRAGMA:
		return "pragma"
	case DIR_LINEMARKER:
		return "linemarker"
	case DIR_EMPTY:
		return "empty"
	default:
		return "unknown"
	}
}

// Directive represents a parsed preprocessing directive.
type Directive struct {
	Type DirectiveType
	Loc  SourceLoc

	// For DIR_INCLUDE
	HeaderName   string // the header name including < > or " "
	IsSystemIncl bool   // true for <...>, false for "..."

	// For DIR_DEFINE
	MacroName   string   // the macro name
	MacroParams []string // nil for object-like, list for function-like
	IsVariadic  bool     // true if last param is ...
	MacroBody   []Token  // replacement tokens

	// For DIR_UNDEF, DIR_IFDEF, DIR_IFNDEF
	Identifier string

	// For DIR_IF, DIR_ELIF
	Expression []Token // conditional expression tokens

	// For DIR_LINE
	LineNum  int
	FileName string // may be empty

	// For DIR_LINEMARKER (GCC extension)
	LinemarkerFlags []int // 1=start of file, 2=return to file, 3=system header, 4=extern "C"

	// For DIR_ERROR, DIR_WARNING
	Message string

	// For DIR_PRAGMA
	PragmaTokens []Token
}

// DirectiveParser parses preprocessing directives from a token stream.
type DirectiveParser struct {
	tokens []Token
	pos    int
}

// NewDirectiveParser creates a new directive parser.
func NewDirectiveParser(tokens []Token) *DirectiveParser {
	return &DirectiveParser{tokens: tokens, pos: 0}
}

// ParseDirective parses a single directive starting after the # token.
// Returns nil if not at a directive.
func (p *DirectiveParser) ParseDirective(loc SourceLoc) (*Directive, error) {
	p.skipWhitespace()

	// Handle empty directive (just # followed by newline)
	if p.atEnd() || p.peek().Type == PP_NEWLINE {
		return &Directive{Type: DIR_EMPTY, Loc: loc}, nil
	}

	// Check for GCC line marker: # number "filename" [flags]
	if p.peek().Type == PP_NUMBER {
		return p.parseLinemarker(loc)
	}

	// Must be an identifier (directive name)
	if p.peek().Type != PP_IDENTIFIER {
		return nil, fmt.Errorf("%s:%d: expected directive name, got %s",
			loc.File, loc.Line, p.peek().Type)
	}

	name := p.peek().Text
	p.advance()

	switch name {
	case "include":
		return p.parseInclude(loc)
	case "define":
		return p.parseDefine(loc)
	case "undef":
		return p.parseUndef(loc)
	case "if":
		return p.parseIf(loc)
	case "ifdef":
		return p.parseIfdef(loc)
	case "ifndef":
		return p.parseIfndef(loc)
	case "elif":
		return p.parseElif(loc)
	case "else":
		return p.parseElse(loc)
	case "endif":
		return p.parseEndif(loc)
	case "line":
		return p.parseLine(loc)
	case "error":
		return p.parseError(loc)
	case "warning":
		return p.parseWarning(loc)
	case "pragma":
		return p.parsePragma(loc)
	default:
		return nil, fmt.Errorf("%s:%d: unknown directive #%s",
			loc.File, loc.Line, name)
	}
}

func (p *DirectiveParser) parseInclude(loc SourceLoc) (*Directive, error) {
	p.skipWhitespace()

	if p.atEnd() || p.peek().Type == PP_NEWLINE {
		return nil, fmt.Errorf("%s:%d: #include expects a file name", loc.File, loc.Line)
	}

	dir := &Directive{Type: DIR_INCLUDE, Loc: loc}

	tok := p.peek()
	if tok.Type == PP_HEADER_NAME {
		dir.HeaderName = tok.Text
		dir.IsSystemIncl = strings.HasPrefix(tok.Text, "<")
		p.advance()
	} else if tok.Type == PP_STRING {
		dir.HeaderName = tok.Text
		dir.IsSystemIncl = false
		p.advance()
	} else if tok.Type == PP_PUNCTUATOR && tok.Text == "<" {
		// Handle <header> when lexer returns it as punctuators
		var header strings.Builder
		header.WriteByte('<')
		p.advance()
		for !p.atEnd() && p.peek().Type != PP_NEWLINE {
			if p.peek().Type == PP_PUNCTUATOR && p.peek().Text == ">" {
				header.WriteByte('>')
				p.advance()
				break
			}
			header.WriteString(p.peek().Text)
			p.advance()
		}
		dir.HeaderName = header.String()
		dir.IsSystemIncl = true
	} else {
		// Could be a macro that expands to header name
		// For now, collect remaining tokens as expression
		dir.Expression = p.collectToNewline()
	}

	return dir, nil
}

func (p *DirectiveParser) parseDefine(loc SourceLoc) (*Directive, error) {
	p.skipWhitespace()

	if p.atEnd() || p.peek().Type == PP_NEWLINE {
		return nil, fmt.Errorf("%s:%d: #define expects an identifier", loc.File, loc.Line)
	}

	if p.peek().Type != PP_IDENTIFIER {
		return nil, fmt.Errorf("%s:%d: #define expects an identifier, got %s",
			loc.File, loc.Line, p.peek().Type)
	}

	dir := &Directive{Type: DIR_DEFINE, Loc: loc}
	dir.MacroName = p.peek().Text
	p.advance()

	// Check for function-like macro (identifier immediately followed by '(')
	// No whitespace between macro name and '(' for function-like
	if !p.atEnd() && p.peek().Type == PP_PUNCTUATOR && p.peek().Text == "(" {
		p.advance() // consume '('
		dir.MacroParams = []string{}

		// Parse parameters
		for !p.atEnd() {
			p.skipWhitespace()
			if p.peek().Type == PP_PUNCTUATOR && p.peek().Text == ")" {
				p.advance()
				break
			}

			// Check for variadic ...
			if p.peek().Type == PP_PUNCTUATOR && p.peek().Text == "..." {
				dir.IsVariadic = true
				p.advance()
				p.skipWhitespace()
				if p.peek().Type != PP_PUNCTUATOR || p.peek().Text != ")" {
					return nil, fmt.Errorf("%s:%d: '...' must be last parameter",
						loc.File, loc.Line)
				}
				p.advance()
				break
			}

			if p.peek().Type != PP_IDENTIFIER {
				return nil, fmt.Errorf("%s:%d: expected parameter name, got %s",
					loc.File, loc.Line, p.peek().Type)
			}

			paramName := p.peek().Text
			// Check for variadic identifier (identifier followed by ...)
			p.advance()
			p.skipWhitespace()
			if p.peek().Type == PP_PUNCTUATOR && p.peek().Text == "..." {
				dir.MacroParams = append(dir.MacroParams, paramName)
				dir.IsVariadic = true
				p.advance()
				p.skipWhitespace()
				if p.peek().Type != PP_PUNCTUATOR || p.peek().Text != ")" {
					return nil, fmt.Errorf("%s:%d: '...' must be last parameter",
						loc.File, loc.Line)
				}
				p.advance()
				break
			}

			dir.MacroParams = append(dir.MacroParams, paramName)

			p.skipWhitespace()
			if p.peek().Type == PP_PUNCTUATOR && p.peek().Text == "," {
				p.advance()
			}
		}
	}

	// Skip optional whitespace before body
	p.skipWhitespace()

	// Collect replacement tokens
	dir.MacroBody = p.collectToNewline()

	return dir, nil
}

func (p *DirectiveParser) parseUndef(loc SourceLoc) (*Directive, error) {
	p.skipWhitespace()

	if p.atEnd() || p.peek().Type == PP_NEWLINE {
		return nil, fmt.Errorf("%s:%d: #undef expects an identifier", loc.File, loc.Line)
	}

	if p.peek().Type != PP_IDENTIFIER {
		return nil, fmt.Errorf("%s:%d: #undef expects an identifier, got %s",
			loc.File, loc.Line, p.peek().Type)
	}

	dir := &Directive{Type: DIR_UNDEF, Loc: loc, Identifier: p.peek().Text}
	p.advance()
	return dir, nil
}

func (p *DirectiveParser) parseIf(loc SourceLoc) (*Directive, error) {
	p.skipWhitespace()
	dir := &Directive{Type: DIR_IF, Loc: loc}
	dir.Expression = p.collectToNewline()
	if len(dir.Expression) == 0 {
		return nil, fmt.Errorf("%s:%d: #if expects an expression", loc.File, loc.Line)
	}
	return dir, nil
}

func (p *DirectiveParser) parseIfdef(loc SourceLoc) (*Directive, error) {
	p.skipWhitespace()

	if p.atEnd() || p.peek().Type == PP_NEWLINE {
		return nil, fmt.Errorf("%s:%d: #ifdef expects an identifier", loc.File, loc.Line)
	}

	if p.peek().Type != PP_IDENTIFIER {
		return nil, fmt.Errorf("%s:%d: #ifdef expects an identifier, got %s",
			loc.File, loc.Line, p.peek().Type)
	}

	dir := &Directive{Type: DIR_IFDEF, Loc: loc, Identifier: p.peek().Text}
	p.advance()
	return dir, nil
}

func (p *DirectiveParser) parseIfndef(loc SourceLoc) (*Directive, error) {
	p.skipWhitespace()

	if p.atEnd() || p.peek().Type == PP_NEWLINE {
		return nil, fmt.Errorf("%s:%d: #ifndef expects an identifier", loc.File, loc.Line)
	}

	if p.peek().Type != PP_IDENTIFIER {
		return nil, fmt.Errorf("%s:%d: #ifndef expects an identifier, got %s",
			loc.File, loc.Line, p.peek().Type)
	}

	dir := &Directive{Type: DIR_IFNDEF, Loc: loc, Identifier: p.peek().Text}
	p.advance()
	return dir, nil
}

func (p *DirectiveParser) parseElif(loc SourceLoc) (*Directive, error) {
	p.skipWhitespace()
	dir := &Directive{Type: DIR_ELIF, Loc: loc}
	dir.Expression = p.collectToNewline()
	if len(dir.Expression) == 0 {
		return nil, fmt.Errorf("%s:%d: #elif expects an expression", loc.File, loc.Line)
	}
	return dir, nil
}

func (p *DirectiveParser) parseElse(loc SourceLoc) (*Directive, error) {
	return &Directive{Type: DIR_ELSE, Loc: loc}, nil
}

func (p *DirectiveParser) parseEndif(loc SourceLoc) (*Directive, error) {
	return &Directive{Type: DIR_ENDIF, Loc: loc}, nil
}

func (p *DirectiveParser) parseLine(loc SourceLoc) (*Directive, error) {
	p.skipWhitespace()

	if p.atEnd() || p.peek().Type == PP_NEWLINE {
		return nil, fmt.Errorf("%s:%d: #line expects a line number", loc.File, loc.Line)
	}

	if p.peek().Type != PP_NUMBER {
		return nil, fmt.Errorf("%s:%d: #line expects a line number, got %s",
			loc.File, loc.Line, p.peek().Type)
	}

	dir := &Directive{Type: DIR_LINE, Loc: loc}
	dir.LineNum = parseIntNumber(p.peek().Text)
	p.advance()

	p.skipWhitespace()

	// Optional filename
	if !p.atEnd() && p.peek().Type == PP_STRING {
		dir.FileName = unquoteString(p.peek().Text)
		p.advance()
	}

	return dir, nil
}

func (p *DirectiveParser) parseLinemarker(loc SourceLoc) (*Directive, error) {
	// GCC line marker format: # linenum "filename" flags
	dir := &Directive{Type: DIR_LINEMARKER, Loc: loc}

	dir.LineNum = parseIntNumber(p.peek().Text)
	p.advance()

	p.skipWhitespace()

	// Optional filename
	if !p.atEnd() && p.peek().Type == PP_STRING {
		dir.FileName = unquoteString(p.peek().Text)
		p.advance()

		// Optional flags (1, 2, 3, 4)
		p.skipWhitespace()
		for !p.atEnd() && p.peek().Type == PP_NUMBER {
			flag := parseIntNumber(p.peek().Text)
			dir.LinemarkerFlags = append(dir.LinemarkerFlags, flag)
			p.advance()
			p.skipWhitespace()
		}
	}

	return dir, nil
}

func (p *DirectiveParser) parseError(loc SourceLoc) (*Directive, error) {
	dir := &Directive{Type: DIR_ERROR, Loc: loc}
	p.skipWhitespace()
	// Collect everything to newline as the message
	var msg strings.Builder
	for !p.atEnd() && p.peek().Type != PP_NEWLINE {
		msg.WriteString(p.peek().Text)
		p.advance()
	}
	dir.Message = strings.TrimSpace(msg.String())
	return dir, nil
}

func (p *DirectiveParser) parseWarning(loc SourceLoc) (*Directive, error) {
	dir := &Directive{Type: DIR_WARNING, Loc: loc}
	p.skipWhitespace()
	// Collect everything to newline as the message
	var msg strings.Builder
	for !p.atEnd() && p.peek().Type != PP_NEWLINE {
		msg.WriteString(p.peek().Text)
		p.advance()
	}
	dir.Message = strings.TrimSpace(msg.String())
	return dir, nil
}

func (p *DirectiveParser) parsePragma(loc SourceLoc) (*Directive, error) {
	dir := &Directive{Type: DIR_PRAGMA, Loc: loc}
	p.skipWhitespace()
	dir.PragmaTokens = p.collectToNewline()
	return dir, nil
}

// Helper methods

func (p *DirectiveParser) atEnd() bool {
	return p.pos >= len(p.tokens)
}

func (p *DirectiveParser) peek() Token {
	if p.atEnd() {
		return Token{Type: PP_EOF}
	}
	return p.tokens[p.pos]
}

func (p *DirectiveParser) advance() {
	if !p.atEnd() {
		p.pos++
	}
}

func (p *DirectiveParser) skipWhitespace() {
	for !p.atEnd() && p.peek().Type == PP_WHITESPACE {
		p.advance()
	}
}

func (p *DirectiveParser) collectToNewline() []Token {
	var tokens []Token
	for !p.atEnd() && p.peek().Type != PP_NEWLINE {
		tokens = append(tokens, p.peek())
		p.advance()
	}
	// Trim trailing whitespace
	for len(tokens) > 0 && tokens[len(tokens)-1].Type == PP_WHITESPACE {
		tokens = tokens[:len(tokens)-1]
	}
	return tokens
}

// parseIntNumber parses a simple integer from a string.
func parseIntNumber(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// unquoteString removes surrounding quotes from a string literal.
func unquoteString(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// ParseDirectiveFromTokens is a convenience function that parses a directive
// from a slice of tokens (after the # has been seen).
func ParseDirectiveFromTokens(tokens []Token, loc SourceLoc) (*Directive, error) {
	parser := NewDirectiveParser(tokens)
	return parser.ParseDirective(loc)
}
