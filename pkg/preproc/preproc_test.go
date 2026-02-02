package preproc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNeedsPreprocessing(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"test.c", true},
		{"test.C", true},
		{"test.i", false},
		{"test.I", false},
		{"test.p", false},
		{"test.P", false},
		{"path/to/file.c", true},
		{"path/to/file.i", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := NeedsPreprocessing(tt.filename)
			if got != tt.want {
				t.Errorf("NeedsPreprocessing(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestPreprocessString(t *testing.T) {
	// Simple test without includes
	source := `int main() { return 42; }`
	result, err := PreprocessString(source, "test.c", nil)
	if err != nil {
		t.Fatalf("PreprocessString failed: %v", err)
	}

	// The result should contain the original code (possibly with #line directives)
	if !strings.Contains(result, "int main()") {
		t.Errorf("preprocessed output should contain 'int main()', got:\n%s", result)
	}
}

func TestPreprocessWithDefine(t *testing.T) {
	source := `
#ifdef TEST_MACRO
int test_defined = 1;
#else
int test_defined = 0;
#endif
`
	opts := &Options{
		Defines: map[string]string{"TEST_MACRO": ""},
	}
	result, err := PreprocessString(source, "test.c", opts)
	if err != nil {
		t.Fatalf("PreprocessString failed: %v", err)
	}

	if !strings.Contains(result, "test_defined = 1") {
		t.Errorf("expected test_defined = 1 (macro defined), got:\n%s", result)
	}
}

func TestPreprocessWithIncludePath(t *testing.T) {
	// Create a temporary directory with a header file
	tmpDir, err := os.MkdirTemp("", "ralph-preproc-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create include directory
	includeDir := filepath.Join(tmpDir, "include")
	if err := os.Mkdir(includeDir, 0755); err != nil {
		t.Fatalf("failed to create include dir: %v", err)
	}

	// Create a header file
	headerContent := `#define MY_VALUE 42`
	headerPath := filepath.Join(includeDir, "myheader.h")
	if err := os.WriteFile(headerPath, []byte(headerContent), 0644); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}

	// Create source file that includes the header
	sourceContent := `#include "myheader.h"
int x = MY_VALUE;
`
	sourcePath := filepath.Join(tmpDir, "test.c")
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	opts := &Options{
		IncludePaths: []string{includeDir},
	}

	result, err := Preprocess(sourcePath, opts)
	if err != nil {
		t.Fatalf("Preprocess failed: %v", err)
	}

	// The macro should be expanded
	if !strings.Contains(result, "int x = 42") {
		t.Errorf("expected 'int x = 42' (macro expanded), got:\n%s", result)
	}
}

func TestPreprocessWithUndefine(t *testing.T) {
	source := `
#ifdef __STDC__
int stdc_defined = 1;
#else
int stdc_defined = 0;
#endif
`
	// Undefine the standard C macro
	opts := &Options{
		Undefines: []string{"__STDC__"},
	}
	result, err := PreprocessString(source, "test.c", opts)
	if err != nil {
		t.Fatalf("PreprocessString failed: %v", err)
	}

	// With __STDC__ undefined, should get stdc_defined = 0
	if !strings.Contains(result, "stdc_defined = 0") {
		t.Errorf("expected stdc_defined = 0 (macro undefined), got:\n%s", result)
	}
}

func TestFindPreprocessor(t *testing.T) {
	cpp := findPreprocessor()
	if cpp == "" {
		t.Skip("no C preprocessor found on this system")
	}
	t.Logf("Found preprocessor: %s", cpp)
}
