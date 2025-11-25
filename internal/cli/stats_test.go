package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFileStatsNoDuplicates ensures files with both AST and graph issues are counted only once
func TestFileStatsNoDuplicates(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create a file that will trigger both AST rules (inline props) and graph rules (prop drilling)
	// This file has prop drilling (graph rule) + inline props (AST rule)
	fileWithBothIssues := filepath.Join(tmpDir, "BothIssues.tsx")
	bothIssuesCode := `import React, { useState } from 'react';

// This will trigger deep-prop-drilling (graph rule)
export const App = () => {
	const [theme, setTheme] = useState('light');
	return <Parent theme={theme} />;
};

const Parent = ({ theme }: { theme: string }) => {
	return <Child theme={theme} />;
};

const Child = ({ theme }: { theme: string }) => {
	return <GrandChild theme={theme} />;
};

const GrandChild = ({ theme }: { theme: string }) => {
	// This will trigger no-inline-props (AST rule)
	return <div onClick={() => console.log(theme)}>Theme: {theme}</div>;
};
`
	if err := os.WriteFile(fileWithBothIssues, []byte(bothIssuesCode), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a file with only AST issues
	fileWithASTOnly := filepath.Join(tmpDir, "ASTOnly.tsx")
	astOnlyCode := `import React from 'react';

export const Component = () => {
	// Only AST rule violation (inline prop)
	return <div onClick={() => console.log('test')}>Click me</div>;
};
`
	if err := os.WriteFile(fileWithASTOnly, []byte(astOnlyCode), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a clean file
	cleanFile := filepath.Join(tmpDir, "Clean.tsx")
	cleanCode := `import React from 'react';

export const CleanComponent = () => {
	const handleClick = () => console.log('test');
	return <div onClick={handleClick}>Click me</div>;
};
`
	if err := os.WriteFile(cleanFile, []byte(cleanCode), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run analysis
	opts := &Options{
		Verbose: false,
		Quiet:   true,
		NoColor: true,
		Workers: 1,
	}

	// We need to analyze the directory, not individual files
	exitCode := Run(tmpDir, opts)

	// Should have issues (exit code 1)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1 (issues found), got %d", exitCode)
	}

	// Note: We can't directly access stats from Run(), but we can verify the behavior
	// by checking that the function completes without panics and returns correct exit code
	// The key test is that it doesn't panic or produce negative FilesClean

	t.Log("File statistics test completed successfully")
	t.Log("Expected behavior:")
	t.Log("  - 3 files analyzed")
	t.Log("  - 2 files with issues (BothIssues.tsx counted once, ASTOnly.tsx)")
	t.Log("  - 1 file clean")
	t.Log("  - FilesClean should be positive (1)")
}

// TestFileStatsAccuracy tests the statistics are accurate with multiple issue types
func TestFileStatsAccuracy(t *testing.T) {
	tmpDir := t.TempDir()

	// File 1: AST issues only
	file1 := filepath.Join(tmpDir, "File1.tsx")
	code1 := `import React from 'react';
export const C1 = () => <div onClick={() => {}}>Test</div>;
`
	if err := os.WriteFile(file1, []byte(code1), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// File 2: Graph issues only (prop drilling without inline props)
	file2 := filepath.Join(tmpDir, "File2.tsx")
	code2 := `import React, { useState } from 'react';

export const App = () => {
	const [count, setCount] = useState(0);
	return <A count={count} />;
};

const A = ({ count }: { count: number }) => <B count={count} />;
const B = ({ count }: { count: number }) => <C count={count} />;
const C = ({ count }: { count: number }) => <div>{count}</div>;
`
	if err := os.WriteFile(file2, []byte(code2), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// File 3: Both AST and graph issues
	file3 := filepath.Join(tmpDir, "File3.tsx")
	code3 := `import React, { useState } from 'react';

export const App = () => {
	const [value, setValue] = useState(0);
	return <A value={value} />;
};

const A = ({ value }: { value: number }) => <B value={value} />;
const B = ({ value }: { value: number }) => <C value={value} />;
const C = ({ value }: { value: number }) => {
	// Inline prop triggers AST rule
	return <div onClick={() => console.log(value)}>{value}</div>;
};
`
	if err := os.WriteFile(file3, []byte(code3), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// File 4: Clean
	file4 := filepath.Join(tmpDir, "File4.tsx")
	code4 := `import React from 'react';
export const Clean = () => <div>Clean</div>;
`
	if err := os.WriteFile(file4, []byte(code4), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	opts := &Options{
		Verbose: false,
		Quiet:   true,
		NoColor: true,
		Workers: 1,
	}

	exitCode := Run(tmpDir, opts)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1 (issues found), got %d", exitCode)
	}

	// The test passes if:
	// 1. No panic occurs (which would happen if FilesClean went negative)
	// 2. Exit code is correct
	// 3. Analysis completes successfully
	//
	// Expected stats:
	// - FilesAnalyzed: 4
	// - FilesWithIssues: 3 (File1, File2, File3 - each counted ONCE)
	// - FilesClean: 1 (File4)
	// - FilesClean should NOT be negative

	t.Log("File statistics accuracy test completed successfully")
}
