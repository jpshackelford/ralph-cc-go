package cpp

import (
	"testing"
)

func TestTokenPasteRescan(t *testing.T) {
	// Test that token pasting creates identifiers that are then expanded as macros
	tests := []struct {
		name     string
		macros   []macroSpec
		input    string
		expected string
	}{
		{
			name: "paste creates macro name that should expand",
			macros: []macroSpec{
				{name: "XY", params: nil, body: "42"},
				{name: "PASTE", params: []string{"a", "b"}, body: "a##b"},
			},
			input:    "PASTE(X, Y)",
			expected: "42",
		},
		{
			name: "three-way paste creates macro name",
			macros: []macroSpec{
				{name: "ABC", params: nil, body: "100"},
				{name: "JOIN3", params: []string{"x", "y", "z"}, body: "x##y##z"},
			},
			input:    "JOIN3(A, B, C)",
			expected: "100",
		},
		{
			name: "stdint-like pattern",
			macros: []macroSpec{
				{name: "INT32_MAX", params: nil, body: "2147483647"},
				{name: "__stdint_join3", params: []string{"a", "b", "c"}, body: "a##b##c"},
				{name: "__INTN_MAX", params: []string{"n"}, body: "__stdint_join3(INT, n, _MAX)"},
			},
			input:    "__INTN_MAX(32)",
			expected: "2147483647",
		},
		{
			// Regression test: spaces around ## should be handled correctly
			name: "paste with spaces around ## operator",
			macros: []macroSpec{
				{name: "INT32_MAX", params: nil, body: "2147483647"},
				{name: "JOIN_SPACE", params: []string{"a", "b", "c"}, body: "a ## b ## c"},
			},
			input:    "JOIN_SPACE(INT, 32, _MAX)",
			expected: "2147483647",
		},
		{
			// Regression test: stdint pattern with spaces (like in real stdint.h)
			name: "stdint pattern with spaces",
			macros: []macroSpec{
				{name: "INT32_MAX", params: nil, body: "2147483647"},
				{name: "__stdint_join3", params: []string{"a", "b", "c"}, body: "a ## b ## c"},
				{name: "__INTN_MAX", params: []string{"n"}, body: "__stdint_join3( INT, n, _MAX)"},
			},
			input:    "__INTN_MAX(32)",
			expected: "2147483647",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := NewMacroTable()
			for _, m := range tt.macros {
				bodyTokens := tokenize(m.body)
				if m.params == nil {
					if err := mt.DefineObject(m.name, bodyTokens, SourceLoc{File: "test", Line: 1}); err != nil {
						t.Fatalf("DefineObject error: %v", err)
					}
				} else {
					if err := mt.DefineFunction(m.name, m.params, m.variadic, bodyTokens, SourceLoc{File: "test", Line: 1}); err != nil {
						t.Fatalf("DefineFunction error: %v", err)
					}
				}
			}

			e := NewExpander(mt)
			result, err := e.ExpandString(tt.input)
			if err != nil {
				t.Fatalf("ExpandString error: %v", err)
			}

			result = normalizeWhitespace(result)
			expected := normalizeWhitespace(tt.expected)
			if result != expected {
				t.Errorf("got %q, want %q", result, expected)
			}
		})
	}
}
