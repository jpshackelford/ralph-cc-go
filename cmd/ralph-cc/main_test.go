package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	if version == "" {
		t.Error("version should not be empty")
	}
}

func TestDebugFlagsExist(t *testing.T) {
	var out, errOut bytes.Buffer
	cmd := newRootCmd(&out, &errOut)

	expectedFlags := []string{"dparse", "dc", "dasm", "dclight", "dcminor", "drtl", "dltl", "dmach"}
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", flagName)
		}
	}
}

func TestDebugFlagsWarnAndExit(t *testing.T) {
	// These flags are still unimplemented
	testCases := []struct {
		flagName string
		wantMsg  string
	}{
		{"dc", "dc"},
		{"dasm", "dasm"},
		{"dclight", "dclight"},
		{"dcminor", "dcminor"},
		{"drtl", "drtl"},
		{"dltl", "dltl"},
		{"dmach", "dmach"},
	}

	for _, tc := range testCases {
		t.Run(tc.flagName, func(t *testing.T) {
			// Reset all flags before each test
			resetDebugFlags()

			var out, errOut bytes.Buffer
			cmd := newRootCmd(&out, &errOut)
			cmd.SetArgs([]string{"--" + tc.flagName, "test.c"})
			err := cmd.Execute()

			// Should return an error
			if err == nil {
				t.Errorf("expected error for flag --%s, got nil", tc.flagName)
			}
			if !errors.Is(err, ErrNotImplemented) {
				t.Errorf("expected ErrNotImplemented, got %v", err)
			}

			output := errOut.String()
			if !strings.Contains(output, tc.wantMsg) {
				t.Errorf("expected output to contain %q, got %q", tc.wantMsg, output)
			}
			if !strings.Contains(output, "not yet implemented") {
				t.Errorf("expected output to contain 'not yet implemented', got %q", output)
			}
		})
	}
}

func TestNoDebugFlagsNoError(t *testing.T) {
	resetDebugFlags()

	var out, errOut bytes.Buffer
	cmd := newRootCmd(&out, &errOut)
	cmd.SetArgs([]string{"test.c"})
	err := cmd.Execute()

	if err != nil {
		t.Errorf("expected no error without debug flags, got %v", err)
	}
}

func TestDParseFlag(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.c")
	content := `int main() { return 0; }`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	resetDebugFlags()

	var out, errOut bytes.Buffer
	cmd := newRootCmd(&out, &errOut)
	cmd.SetArgs([]string{"--dparse", testFile})
	err := cmd.Execute()

	if err != nil {
		t.Errorf("expected no error for -dparse, got %v", err)
	}

	output := out.String()
	// Check that it contains expected AST output
	if !strings.Contains(output, "int main()") {
		t.Errorf("expected output to contain 'int main()', got %q", output)
	}
	if !strings.Contains(output, "return 0") {
		t.Errorf("expected output to contain 'return 0', got %q", output)
	}
}

func TestDParseFlagMultipleFunctions(t *testing.T) {
	// Create a temporary test file with multiple functions
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "multi.c")
	content := `int add(int a, int b) { return a + b; }
int main() { return add(1, 2); }`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	resetDebugFlags()

	var out, errOut bytes.Buffer
	cmd := newRootCmd(&out, &errOut)
	cmd.SetArgs([]string{"--dparse", testFile})
	err := cmd.Execute()

	if err != nil {
		t.Errorf("expected no error for -dparse, got %v", err)
	}

	output := out.String()
	// Check that it contains both functions
	if !strings.Contains(output, "int add(") {
		t.Errorf("expected output to contain 'int add(', got %q", output)
	}
	if !strings.Contains(output, "int main()") {
		t.Errorf("expected output to contain 'int main()', got %q", output)
	}
}

func TestDParseFlagFileNotFound(t *testing.T) {
	resetDebugFlags()

	var out, errOut bytes.Buffer
	cmd := newRootCmd(&out, &errOut)
	cmd.SetArgs([]string{"--dparse", "nonexistent.c"})
	err := cmd.Execute()

	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestDParseCreatesOutputFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.c")
	content := `int main() { return 42; }`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	expectedOutputFile := filepath.Join(tmpDir, "test.parsed.c")

	resetDebugFlags()

	var out, errOut bytes.Buffer
	cmd := newRootCmd(&out, &errOut)
	cmd.SetArgs([]string{"--dparse", testFile})
	err := cmd.Execute()

	if err != nil {
		t.Errorf("expected no error for -dparse, got %v", err)
	}

	// Check that output file was created
	if _, err := os.Stat(expectedOutputFile); os.IsNotExist(err) {
		t.Errorf("expected output file %s to be created", expectedOutputFile)
	}

	// Check output file contents match stdout
	fileContent, err := os.ReadFile(expectedOutputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if out.String() != string(fileContent) {
		t.Errorf("output file content doesn't match stdout\nStdout:\n%s\nFile:\n%s", out.String(), string(fileContent))
	}

	// Verify content looks correct
	if !strings.Contains(string(fileContent), "int main()") {
		t.Errorf("expected output file to contain 'int main()'")
	}
	if !strings.Contains(string(fileContent), "return 42") {
		t.Errorf("expected output file to contain 'return 42'")
	}
}

func TestParsedOutputFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test.c", "test.parsed.c"},
		{"path/to/file.c", "path/to/file.parsed.c"},
		{"/absolute/path.c", "/absolute/path.parsed.c"},
		{"no_extension", "no_extension.parsed.c"},
		{"multiple.dots.c", "multiple.dots.parsed.c"},
	}

	for _, tc := range tests {
		result := parsedOutputFilename(tc.input)
		if result != tc.expected {
			t.Errorf("parsedOutputFilename(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func resetDebugFlags() {
	dParse = false
	dC = false
	dAsm = false
	dClight = false
	dCminor = false
	dRTL = false
	dLTL = false
	dMach = false
}
