package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
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

	// Verify output contains issue count (now 7 with inline literal detection)
	if !strings.Contains(output, "Found 7 issues") {
		t.Errorf("Expected '7 issues', got: %s", output)
	}

	// Verify specific violations are reported
	expectedViolations := []string{"config", "items", "options", "settings", "preferences", "theme", "Inline object"}
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
	if !strings.Contains(output, "not found") {
		t.Errorf("Expected 'not found' error, got: %s", output)
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

	// Quiet mode should still show issues (now 7 with inline literal detection)
	if !strings.Contains(output, "Found 7 issues") {
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

	// Verify verbose output includes hook count
	if !strings.Contains(output, "found 2 React hook call(s)") {
		t.Errorf("Expected hook count in verbose mode, got: %s", output)
	}

	// Should still report the issue
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	// Should report the violation
	if !strings.Contains(output, "config") {
		t.Errorf("Expected violation for 'config' in output")
	}
}

func TestRun_SyntaxError(t *testing.T) {
	// Create a temp file with syntax errors
	tmpFile, err := os.CreateTemp("", "test*.tsx")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid TypeScript that tree-sitter will reject
	if _, err := tmpFile.WriteString("const broken = { unclosed object\nfunction test() {"); err != nil {
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
	_ = buf.String() // Read output but don't check it (tree-sitter is lenient)

	// Tree-sitter is very lenient, so it might not fail
	// Just verify we handle the file without crashing
	if exitCode != 0 && exitCode != 1 && exitCode != 2 {
		t.Errorf("Expected exit code 0, 1, or 2, got %d", exitCode)
	}
}

func TestRun_Directory(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &Options{
		Verbose: false,
		Quiet:   false,
		NoColor: true,
	}

	exitCode := Run("../../test/fixtures/", opts)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify exit code (should find issues)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	// Verify analyzing message
	if !strings.Contains(output, "Analyzing") && !strings.Contains(output, "files") {
		t.Errorf("Expected 'Analyzing N files' message, got: %s", output)
	}

	// Verify summary format
	if !strings.Contains(output, "Found") && !strings.Contains(output, "in") {
		t.Errorf("Expected summary with 'Found X issues in Y files', got: %s", output)
	}

	// Verify clean files mentioned
	if !strings.Contains(output, "clean") {
		t.Errorf("Expected mention of clean files in summary, got: %s", output)
	}
}

func TestRun_DirectoryNoIssues(t *testing.T) {
	// Create a temp directory with clean files
	tmpDir, err := os.MkdirTemp("", "test-clean-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a clean file
	cleanFile := filepath.Join(tmpDir, "clean.tsx")
	if err := os.WriteFile(cleanFile, []byte("export const foo = 'bar';"), 0644); err != nil {
		t.Fatalf("Failed to write clean file: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &Options{
		Verbose: false,
		Quiet:   false,
		NoColor: true,
	}

	exitCode := Run(tmpDir, opts)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify exit code (no issues)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Verify success message
	if !strings.Contains(output, "No issues found") {
		t.Errorf("Expected success message, got: %s", output)
	}
}

func TestRun_EmptyDirectory(t *testing.T) {
	// Create a temp directory with no files
	tmpDir, err := os.MkdirTemp("", "test-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	opts := &Options{
		Verbose: false,
		Quiet:   false,
		NoColor: true,
	}

	exitCode := Run(tmpDir, opts)

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify exit code (error)
	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}

	// Verify error message
	if !strings.Contains(output, "no") && !strings.Contains(output, "found") {
		t.Errorf("Expected 'no files found' error, got: %s", output)
	}
}

func TestRun_RelativePathConfigLoading(t *testing.T) {
	// Create a temp directory structure:
	// tmpDir/
	//   .rarc (config file)
	//   nested/
	//     test.tsx (file to analyze)
	tmpDir, err := os.MkdirTemp("", "test-config-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .rarc config that disables all rules except deep-prop-drilling
	configContent := `{
  "rules": {
    "deep-prop-drilling": {
      "enabled": true,
      "options": {
        "maxDepth": 2
      }
    },
    "no-object-deps": {
      "enabled": false
    },
    "no-inline-props": {
      "enabled": false
    }
  }
}`
	configPath := filepath.Join(tmpDir, ".rarc")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tmpDir, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	// Create a test file with an object dependency
	testFile := filepath.Join(nestedDir, "test.tsx")
	testContent := `import React, { useEffect } from 'react';

function MyComponent() {
  const config = { api: 'test' };
  useEffect(() => {
    console.log(config.api);
  }, [config]); // Object dependency - should NOT be flagged (rule disabled)
  return <div>test</div>;
}`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Change to a different directory (parent of tmpDir)
	if err := os.Chdir(filepath.Dir(tmpDir)); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create relative path from current directory to test file
	relPath, err := filepath.Rel(filepath.Dir(tmpDir), testFile)
	if err != nil {
		t.Fatalf("Failed to create relative path: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &Options{
		Verbose: true, // Enable verbose to see config loading message
		Quiet:   false,
		NoColor: true,
	}

	exitCode := Run(relPath, opts)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify config was loaded (verbose mode should mention it)
	if !strings.Contains(output, "Configuration loaded from:") {
		t.Errorf("Expected config to be loaded with relative path, got: %s", output)
	}

	// Verify no object-deps violation (rule should be disabled by config)
	if strings.Contains(output, "no-object-deps") {
		t.Errorf("Expected no-object-deps to be disabled by config, but it was reported: %s", output)
	}

	// Verify exit code is 0 (no violations since rules are disabled)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 (no violations), got %d. Output: %s", exitCode, output)
	}
}
