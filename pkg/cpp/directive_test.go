package cpp

import (
	"testing"
)

func TestDirectiveTypeString(t *testing.T) {
	tests := []struct {
		dt   DirectiveType
		want string
	}{
		{DIR_INCLUDE, "include"},
		{DIR_DEFINE, "define"},
		{DIR_UNDEF, "undef"},
		{DIR_IF, "if"},
		{DIR_IFDEF, "ifdef"},
		{DIR_IFNDEF, "ifndef"},
		{DIR_ELIF, "elif"},
		{DIR_ELSE, "else"},
		{DIR_ENDIF, "endif"},
		{DIR_LINE, "line"},
		{DIR_ERROR, "error"},
		{DIR_WARNING, "warning"},
		{DIR_PRAGMA, "pragma"},
		{DIR_LINEMARKER, "linemarker"},
		{DIR_EMPTY, "empty"},
		{DirectiveType(999), "unknown"},
	}
	for _, tc := range tests {
		if got := tc.dt.String(); got != tc.want {
			t.Errorf("DirectiveType(%d).String() = %q, want %q", tc.dt, got, tc.want)
		}
	}
}

func parseDirective(t *testing.T, input string) *Directive {
	t.Helper()
	lexer := NewLexer(input, "test.c")
	var tokens []Token
	// Skip the # token
	tok := lexer.NextToken()
	if tok.Type != PP_HASH {
		t.Fatalf("expected HASH, got %v", tok.Type)
	}
	loc := tok.Loc
	// Collect remaining tokens until newline/EOF
	for {
		tok := lexer.NextToken()
		if tok.Type == PP_EOF || tok.Type == PP_NEWLINE {
			break
		}
		tokens = append(tokens, tok)
	}
	dir, err := ParseDirectiveFromTokens(tokens, loc)
	if err != nil {
		t.Fatalf("ParseDirective error: %v", err)
	}
	return dir
}

func TestParseIncludeAngleBracket(t *testing.T) {
	dir := parseDirective(t, `#include <stdio.h>`)
	if dir.Type != DIR_INCLUDE {
		t.Errorf("got type %v, want DIR_INCLUDE", dir.Type)
	}
	if dir.HeaderName != "<stdio.h>" {
		t.Errorf("got header %q, want %q", dir.HeaderName, "<stdio.h>")
	}
	if !dir.IsSystemIncl {
		t.Errorf("expected IsSystemIncl to be true")
	}
}

func TestParseIncludeQuoted(t *testing.T) {
	dir := parseDirective(t, `#include "myfile.h"`)
	if dir.Type != DIR_INCLUDE {
		t.Errorf("got type %v, want DIR_INCLUDE", dir.Type)
	}
	if dir.HeaderName != `"myfile.h"` {
		t.Errorf("got header %q, want %q", dir.HeaderName, `"myfile.h"`)
	}
	if dir.IsSystemIncl {
		t.Errorf("expected IsSystemIncl to be false")
	}
}

func TestParseDefineObject(t *testing.T) {
	dir := parseDirective(t, `#define FOO 42`)
	if dir.Type != DIR_DEFINE {
		t.Errorf("got type %v, want DIR_DEFINE", dir.Type)
	}
	if dir.MacroName != "FOO" {
		t.Errorf("got name %q, want %q", dir.MacroName, "FOO")
	}
	if dir.MacroParams != nil {
		t.Errorf("expected nil params for object-like macro")
	}
	if len(dir.MacroBody) != 1 {
		t.Errorf("got %d body tokens, want 1", len(dir.MacroBody))
	}
	if dir.MacroBody[0].Text != "42" {
		t.Errorf("got body %q, want %q", dir.MacroBody[0].Text, "42")
	}
}

func TestParseDefineEmpty(t *testing.T) {
	dir := parseDirective(t, `#define EMPTY`)
	if dir.Type != DIR_DEFINE {
		t.Errorf("got type %v, want DIR_DEFINE", dir.Type)
	}
	if dir.MacroName != "EMPTY" {
		t.Errorf("got name %q, want %q", dir.MacroName, "EMPTY")
	}
	if len(dir.MacroBody) != 0 {
		t.Errorf("got %d body tokens, want 0", len(dir.MacroBody))
	}
}

func TestParseDefineFunctionLike(t *testing.T) {
	dir := parseDirective(t, `#define MAX(a, b) ((a) > (b) ? (a) : (b))`)
	if dir.Type != DIR_DEFINE {
		t.Errorf("got type %v, want DIR_DEFINE", dir.Type)
	}
	if dir.MacroName != "MAX" {
		t.Errorf("got name %q, want %q", dir.MacroName, "MAX")
	}
	if len(dir.MacroParams) != 2 {
		t.Errorf("got %d params, want 2", len(dir.MacroParams))
	}
	if dir.MacroParams[0] != "a" || dir.MacroParams[1] != "b" {
		t.Errorf("got params %v, want [a, b]", dir.MacroParams)
	}
	if dir.IsVariadic {
		t.Errorf("expected non-variadic")
	}
}

func TestParseDefineFunctionLikeNoParams(t *testing.T) {
	dir := parseDirective(t, `#define GETVAL() getValue()`)
	if dir.Type != DIR_DEFINE {
		t.Errorf("got type %v, want DIR_DEFINE", dir.Type)
	}
	if dir.MacroName != "GETVAL" {
		t.Errorf("got name %q, want %q", dir.MacroName, "GETVAL")
	}
	if len(dir.MacroParams) != 0 {
		t.Errorf("got %d params, want 0", len(dir.MacroParams))
	}
}

func TestParseDefineVariadic(t *testing.T) {
	dir := parseDirective(t, `#define PRINTF(fmt, ...) printf(fmt, __VA_ARGS__)`)
	if dir.Type != DIR_DEFINE {
		t.Errorf("got type %v, want DIR_DEFINE", dir.Type)
	}
	if dir.MacroName != "PRINTF" {
		t.Errorf("got name %q, want %q", dir.MacroName, "PRINTF")
	}
	if len(dir.MacroParams) != 1 {
		t.Errorf("got %d params, want 1", len(dir.MacroParams))
	}
	if dir.MacroParams[0] != "fmt" {
		t.Errorf("got params %v, want [fmt]", dir.MacroParams)
	}
	if !dir.IsVariadic {
		t.Errorf("expected variadic")
	}
}

func TestParseUndef(t *testing.T) {
	dir := parseDirective(t, `#undef FOO`)
	if dir.Type != DIR_UNDEF {
		t.Errorf("got type %v, want DIR_UNDEF", dir.Type)
	}
	if dir.Identifier != "FOO" {
		t.Errorf("got identifier %q, want %q", dir.Identifier, "FOO")
	}
}

func TestParseIf(t *testing.T) {
	dir := parseDirective(t, `#if defined(FOO) && X > 0`)
	if dir.Type != DIR_IF {
		t.Errorf("got type %v, want DIR_IF", dir.Type)
	}
	if len(dir.Expression) == 0 {
		t.Errorf("expected non-empty expression")
	}
}

func TestParseIfdef(t *testing.T) {
	dir := parseDirective(t, `#ifdef __GNUC__`)
	if dir.Type != DIR_IFDEF {
		t.Errorf("got type %v, want DIR_IFDEF", dir.Type)
	}
	if dir.Identifier != "__GNUC__" {
		t.Errorf("got identifier %q, want %q", dir.Identifier, "__GNUC__")
	}
}

func TestParseIfndef(t *testing.T) {
	dir := parseDirective(t, `#ifndef HEADER_H`)
	if dir.Type != DIR_IFNDEF {
		t.Errorf("got type %v, want DIR_IFNDEF", dir.Type)
	}
	if dir.Identifier != "HEADER_H" {
		t.Errorf("got identifier %q, want %q", dir.Identifier, "HEADER_H")
	}
}

func TestParseElif(t *testing.T) {
	dir := parseDirective(t, `#elif X == 2`)
	if dir.Type != DIR_ELIF {
		t.Errorf("got type %v, want DIR_ELIF", dir.Type)
	}
	if len(dir.Expression) == 0 {
		t.Errorf("expected non-empty expression")
	}
}

func TestParseElse(t *testing.T) {
	dir := parseDirective(t, `#else`)
	if dir.Type != DIR_ELSE {
		t.Errorf("got type %v, want DIR_ELSE", dir.Type)
	}
}

func TestParseEndif(t *testing.T) {
	dir := parseDirective(t, `#endif`)
	if dir.Type != DIR_ENDIF {
		t.Errorf("got type %v, want DIR_ENDIF", dir.Type)
	}
}

func TestParseLine(t *testing.T) {
	dir := parseDirective(t, `#line 100`)
	if dir.Type != DIR_LINE {
		t.Errorf("got type %v, want DIR_LINE", dir.Type)
	}
	if dir.LineNum != 100 {
		t.Errorf("got line %d, want %d", dir.LineNum, 100)
	}
	if dir.FileName != "" {
		t.Errorf("got filename %q, want empty", dir.FileName)
	}
}

func TestParseLineWithFile(t *testing.T) {
	dir := parseDirective(t, `#line 50 "other.c"`)
	if dir.Type != DIR_LINE {
		t.Errorf("got type %v, want DIR_LINE", dir.Type)
	}
	if dir.LineNum != 50 {
		t.Errorf("got line %d, want %d", dir.LineNum, 50)
	}
	if dir.FileName != "other.c" {
		t.Errorf("got filename %q, want %q", dir.FileName, "other.c")
	}
}

func TestParseLinemarker(t *testing.T) {
	// GCC line marker format: # linenum "filename" flags
	lexer := NewLexer(`# 1 "test.c" 1 3`, "test.c")
	var tokens []Token
	tok := lexer.NextToken()
	if tok.Type != PP_HASH {
		t.Fatalf("expected HASH, got %v", tok.Type)
	}
	loc := tok.Loc
	for {
		tok := lexer.NextToken()
		if tok.Type == PP_EOF || tok.Type == PP_NEWLINE {
			break
		}
		tokens = append(tokens, tok)
	}
	dir, err := ParseDirectiveFromTokens(tokens, loc)
	if err != nil {
		t.Fatalf("ParseDirective error: %v", err)
	}
	if dir.Type != DIR_LINEMARKER {
		t.Errorf("got type %v, want DIR_LINEMARKER", dir.Type)
	}
	if dir.LineNum != 1 {
		t.Errorf("got line %d, want %d", dir.LineNum, 1)
	}
	if dir.FileName != "test.c" {
		t.Errorf("got filename %q, want %q", dir.FileName, "test.c")
	}
	if len(dir.LinemarkerFlags) != 2 || dir.LinemarkerFlags[0] != 1 || dir.LinemarkerFlags[1] != 3 {
		t.Errorf("got flags %v, want [1, 3]", dir.LinemarkerFlags)
	}
}

func TestParseError(t *testing.T) {
	dir := parseDirective(t, `#error This is an error message`)
	if dir.Type != DIR_ERROR {
		t.Errorf("got type %v, want DIR_ERROR", dir.Type)
	}
	if dir.Message != "This is an error message" {
		t.Errorf("got message %q, want %q", dir.Message, "This is an error message")
	}
}

func TestParseWarning(t *testing.T) {
	dir := parseDirective(t, `#warning Deprecated usage`)
	if dir.Type != DIR_WARNING {
		t.Errorf("got type %v, want DIR_WARNING", dir.Type)
	}
	if dir.Message != "Deprecated usage" {
		t.Errorf("got message %q, want %q", dir.Message, "Deprecated usage")
	}
}

func TestParsePragma(t *testing.T) {
	dir := parseDirective(t, `#pragma once`)
	if dir.Type != DIR_PRAGMA {
		t.Errorf("got type %v, want DIR_PRAGMA", dir.Type)
	}
	if len(dir.PragmaTokens) != 1 || dir.PragmaTokens[0].Text != "once" {
		t.Errorf("got pragma tokens %v", dir.PragmaTokens)
	}
}

func TestParsePragmaGCC(t *testing.T) {
	dir := parseDirective(t, `#pragma GCC diagnostic push`)
	if dir.Type != DIR_PRAGMA {
		t.Errorf("got type %v, want DIR_PRAGMA", dir.Type)
	}
	// Should have GCC, diagnostic, push (with whitespace between)
	nonWS := []string{}
	for _, tok := range dir.PragmaTokens {
		if tok.Type != PP_WHITESPACE {
			nonWS = append(nonWS, tok.Text)
		}
	}
	if len(nonWS) != 3 || nonWS[0] != "GCC" || nonWS[1] != "diagnostic" || nonWS[2] != "push" {
		t.Errorf("got pragma tokens %v", nonWS)
	}
}

func TestParseEmptyDirective(t *testing.T) {
	lexer := NewLexer("#\n", "test.c")
	tok := lexer.NextToken()
	if tok.Type != PP_HASH {
		t.Fatalf("expected HASH, got %v", tok.Type)
	}
	loc := tok.Loc
	var tokens []Token
	for {
		tok := lexer.NextToken()
		if tok.Type == PP_EOF || tok.Type == PP_NEWLINE {
			break
		}
		tokens = append(tokens, tok)
	}
	dir, err := ParseDirectiveFromTokens(tokens, loc)
	if err != nil {
		t.Fatalf("ParseDirective error: %v", err)
	}
	if dir.Type != DIR_EMPTY {
		t.Errorf("got type %v, want DIR_EMPTY", dir.Type)
	}
}

func TestParseDirectiveErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unknown directive", "#foo"},
		{"include missing file", "#include"},
		{"define missing name", "#define"},
		{"undef missing name", "#undef"},
		{"ifdef missing name", "#ifdef"},
		{"ifndef missing name", "#ifndef"},
		{"if missing expression", "#if"},
		{"elif missing expression", "#elif"},
		{"line missing number", "#line"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(tc.input, "test.c")
			var tokens []Token
			tok := lexer.NextToken()
			if tok.Type != PP_HASH {
				t.Fatalf("expected HASH, got %v", tok.Type)
			}
			loc := tok.Loc
			for {
				tok := lexer.NextToken()
				if tok.Type == PP_EOF || tok.Type == PP_NEWLINE {
					break
				}
				tokens = append(tokens, tok)
			}
			_, err := ParseDirectiveFromTokens(tokens, loc)
			if err == nil {
				t.Errorf("expected error for %q", tc.input)
			}
		})
	}
}

func TestParseIntNumber(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"42", 42},
		{"0", 0},
		{"123", 123},
		{"abc", 0}, // invalid input
	}
	for _, tc := range tests {
		if got := parseIntNumber(tc.input); got != tc.want {
			t.Errorf("parseIntNumber(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestUnquoteString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`"path/to/file.h"`, "path/to/file.h"},
		{`""`, ""},
		{`noQuotes`, "noQuotes"},
	}
	for _, tc := range tests {
		if got := unquoteString(tc.input); got != tc.want {
			t.Errorf("unquoteString(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
