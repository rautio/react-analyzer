package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRun_ValidFile(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &Options{
		Verbose: false,
		Quiet:   false,
		NoColor: true, // Disable colors for easier testing
	}

	exitCode := Run("../../test/fixtures/valid-object-deps.tsx", opts)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify output contains success message
	if !strings.Contains(output, "No issues found") {
		t.Errorf("Expected success message, got: %s", output)
	}
}

func TestRun_FileWithViolations(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &Options{
		Verbose: false,
		Quiet:   false,
		NoColor: true,
	}

	exitCode := Run("../../test/fixtures/object-in-render.tsx", opts)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify exit code
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	// Verify output contains issue count
	if !strings.Contains(output, "Found 6 issues") {
		t.Errorf("Expected '6 issues', got: %s", output)
	}

	// Verify specific violations are reported
	expectedViolations := []string{"config", "items", "options", "settings", "preferences", "theme"}
	for _, varName := range expectedViolations {
		if !strings.Contains(output, varName) {
			t.Errorf("Expected violation for '%s' in output", varName)
		}
	}

	// Verify rule name appears
	if !strings.Contains(output, "no-object-deps") {
		t.Errorf("Expected rule name 'no-object-deps' in output")
	}
}

func TestRun_FileNotFound(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	opts := &Options{
		Verbose: false,
		Quiet:   false,
		NoColor: true,
	}

	exitCode := Run("nonexistent.tsx", opts)

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify exit code
	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}

	// Verify error message
	if !strings.Contains(output, "file not found") {
		t.Errorf("Expected 'file not found' error, got: %s", output)
	}
}

func TestRun_InvalidExtension(t *testing.T) {
	// Create a temp file with wrong extension
	tmpFile, err := os.CreateTemp("", "test*.py")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	opts := &Options{
		Verbose: false,
		Quiet:   false,
		NoColor: true,
	}

	exitCode := Run(tmpFile.Name(), opts)

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify exit code
	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}

	// Verify error message
	if !strings.Contains(output, "unsupported file type") {
		t.Errorf("Expected 'unsupported file type' error, got: %s", output)
	}
}

func TestRun_QuietMode(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &Options{
		Verbose: false,
		Quiet:   true, // Quiet mode
		NoColor: true,
	}

	exitCode := Run("../../test/fixtures/valid-object-deps.tsx", opts)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify no output in quiet mode when successful
	if strings.TrimSpace(output) != "" {
		t.Errorf("Expected no output in quiet mode, got: %s", output)
	}
}

func TestRun_QuietModeWithIssues(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &Options{
		Verbose: false,
		Quiet:   true, // Quiet mode
		NoColor: true,
	}

	exitCode := Run("../../test/fixtures/object-in-render.tsx", opts)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify exit code
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	// Quiet mode should still show issues
	if !strings.Contains(output, "Found 6 issues") {
		t.Errorf("Expected issues to be shown even in quiet mode, got: %s", output)
	}
}

func TestRun_VerboseMode(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &Options{
		Verbose: true, // Verbose mode
		Quiet:   false,
		NoColor: true,
	}

	exitCode := Run("../../test/fixtures/with-hooks.tsx", opts)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify verbose output includes analysis details
	if !strings.Contains(output, "React Analyzer") {
		t.Errorf("Expected version info in verbose mode")
	}

	if !strings.Contains(output, "Analyzing:") {
		t.Errorf("Expected 'Analyzing:' in verbose mode")
	}

	if !strings.Contains(output, "File size:") {
		t.Errorf("Expected file size in verbose mode")
	}

	if !strings.Contains(output, "Found 2 React hook call(s)") {
		t.Errorf("Expected hook count in verbose mode")
	}

	if !strings.Contains(output, "Rules enabled:") {
		t.Errorf("Expected rules list in verbose mode")
	}

	// Should still report the issue
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
}

func TestRun_SyntaxError(t *testing.T) {
	// Create a temp file with syntax errors
	tmpFile, err := os.CreateTemp("", "test*.tsx")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid TypeScript
	if _, err := tmpFile.WriteString("const broken = {"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	opts := &Options{
		Verbose: false,
		Quiet:   false,
		NoColor: true,
	}

	exitCode := Run(tmpFile.Name(), opts)

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify exit code
	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}

	// Verify error message mentions parsing
	if !strings.Contains(output, "parse") && !strings.Contains(output, "syntax") {
		t.Errorf("Expected parse/syntax error, got: %s", output)
	}
}
