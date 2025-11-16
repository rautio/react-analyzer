package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Options contains CLI configuration
type Options struct {
	Verbose bool
	Quiet   bool
	NoColor bool
}

// Run executes the analysis and returns exit code
func Run(filePath string, opts *Options) int {
	// Validate file exists
	if err := validateFile(filePath); err != nil {
		printError(err, opts.NoColor)
		return 2
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		printError(fmt.Errorf("cannot read file: %s\n%v", filePath, err), opts.NoColor)
		return 2
	}

	// TODO: Parse file and run analysis
	// For now, just return success
	_ = content // Will be used when we add the parser

	if opts.Verbose {
		fmt.Printf("React Analyzer v0.1.0\n")
		fmt.Printf("Analyzing: %s\n", filePath)
		fmt.Printf("File size: %d bytes\n", len(content))
		fmt.Printf("\nRules enabled:\n")
		fmt.Printf("  ✓ no-object-deps      Prevent inline objects in hook dependencies\n\n")
	}

	// Mock result for now
	if !opts.Quiet {
		fmt.Printf("✓ No issues found in %s\n", filePath)
	}

	return 0
}

// validateFile checks if the file exists and has a valid extension
func validateFile(filePath string) error {
	// Check if file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}
	if err != nil {
		return fmt.Errorf("cannot access file: %s", filePath)
	}

	// Check if it's a file (not directory)
	if info.IsDir() {
		return fmt.Errorf("expected a file, got directory: %s", filePath)
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(filePath))
	validExts := []string{".tsx", ".jsx", ".ts", ".js"}

	valid := false
	for _, validExt := range validExts {
		if ext == validExt {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("unsupported file type: %s\nSupported extensions: .tsx, .jsx, .ts, .js", ext)
	}

	return nil
}

// printError formats and prints an error message
func printError(err error, noColor bool) {
	if noColor {
		fmt.Fprintf(os.Stderr, "✖ Error: %v\n", err)
	} else {
		// Red color for errors
		fmt.Fprintf(os.Stderr, "\033[31m✖ Error:\033[0m %v\n", err)
	}
}
