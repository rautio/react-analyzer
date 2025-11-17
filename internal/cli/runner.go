package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oskari/react-analyzer/internal/parser"
	"github.com/oskari/react-analyzer/internal/rules"
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

	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		printError(fmt.Errorf("failed to initialize parser: %v", err), opts.NoColor)
		return 2
	}
	defer p.Close()

	// Parse file
	ast, err := p.ParseFile(filePath, content)
	if err != nil {
		printError(fmt.Errorf("failed to parse file: %v", err), opts.NoColor)
		return 2
	}
	defer ast.Close()

	if opts.Verbose {
		fmt.Printf("React Analyzer v0.1.0\n")
		fmt.Printf("Analyzing: %s\n", filePath)
		fmt.Printf("File size: %d bytes\n", len(content))

		// Count hooks found
		hookCount := countHooks(ast)
		if hookCount > 0 {
			fmt.Printf("Found %d React hook call(s)\n", hookCount)
		}

		fmt.Printf("\nRules enabled:\n")
		fmt.Printf("  ✓ no-object-deps      Prevent inline objects in hook dependencies\n\n")
	}

	// Run rule analysis
	rule := &rules.NoObjectDeps{}
	issues := rule.Check(ast)

	// Display results
	if len(issues) == 0 {
		if !opts.Quiet {
			printSuccess(fmt.Sprintf("No issues found in %s", filePath), opts.NoColor)
		}
		return 0
	}

	// Print issues
	printIssues(issues, opts)

	// Return exit code 1 when issues are found
	return 1
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

// printSuccess formats and prints a success message
func printSuccess(message string, noColor bool) {
	if noColor {
		fmt.Printf("✓ %s\n", message)
	} else {
		// Green color for success
		fmt.Printf("\033[32m✓\033[0m %s\n", message)
	}
}

// printIssues formats and prints rule violations
func printIssues(issues []rules.Issue, opts *Options) {
	// Print summary
	issueWord := "issue"
	if len(issues) > 1 {
		issueWord = "issues"
	}

	if opts.NoColor {
		fmt.Printf("\n✖ Found %d %s:\n\n", len(issues), issueWord)
	} else {
		fmt.Printf("\n\033[31m✖ Found %d %s:\033[0m\n\n", len(issues), issueWord)
	}

	// Print each issue
	for _, issue := range issues {
		// Format: filename:line:column - [rule] message
		location := fmt.Sprintf("%s:%d:%d", issue.FilePath, issue.Line, issue.Column+1)

		if opts.NoColor {
			fmt.Printf("%s - [%s] %s\n", location, issue.Rule, issue.Message)
		} else {
			// Gray for location, yellow for rule, white for message
			fmt.Printf("\033[90m%s\033[0m - \033[33m[%s]\033[0m %s\n",
				location, issue.Rule, issue.Message)
		}
	}

	fmt.Println() // Empty line after issues
}

// countHooks counts the number of React hook calls in the AST
func countHooks(ast *parser.AST) int {
	count := 0
	ast.Root.Walk(func(node *parser.Node) bool {
		if node.IsHookCall() {
			count++
		}
		return true
	})
	return count
}
