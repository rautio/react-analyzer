package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oskari/react-analyzer/internal/analyzer"
	"github.com/oskari/react-analyzer/internal/parser"
	"github.com/oskari/react-analyzer/internal/rules"
)

// validExtensions are the file extensions we analyze
var validExtensions = []string{".tsx", ".jsx", ".ts", ".js"}

// Options contains CLI configuration
type Options struct {
	Verbose bool
	Quiet   bool
	NoColor bool
}

// Run executes the analysis and returns exit code
func Run(path string, opts *Options) int {
	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		printError(fmt.Errorf("path not found: %s", path), opts.NoColor)
		return 2
	}
	if err != nil {
		printError(fmt.Errorf("cannot access path: %s", path), opts.NoColor)
		return 2
	}

	// Determine if it's a file or directory
	var filesToAnalyze []string
	if info.IsDir() {
		// Find all relevant files in directory
		files, err := findFiles(path)
		if err != nil {
			printError(fmt.Errorf("failed to scan directory: %v", err), opts.NoColor)
			return 2
		}
		if len(files) == 0 {
			printError(fmt.Errorf("no .tsx, .jsx, .ts, or .js files found in %s", path), opts.NoColor)
			return 2
		}
		filesToAnalyze = files
	} else {
		// Single file - validate extension
		if err := validateFileExtension(path); err != nil {
			printError(err, opts.NoColor)
			return 2
		}
		filesToAnalyze = []string{path}
	}

	// Create rule registry
	registry := rules.NewRegistry()

	// Create module resolver for cross-file analysis
	// Use the provided path as the base directory, or parent directory if it's a file
	baseDir := path
	if !info.IsDir() {
		baseDir = filepath.Dir(path)
	}
	resolver, err := analyzer.NewModuleResolver(baseDir)
	if err != nil {
		printError(fmt.Errorf("failed to initialize module resolver: %v", err), opts.NoColor)
		return 2
	}
	defer resolver.Close()

	// Print analysis start for directories
	if len(filesToAnalyze) > 1 && !opts.Quiet {
		fmt.Printf("Analyzing %d files...\n\n", len(filesToAnalyze))
	}

	if opts.Verbose {
		fmt.Printf("Rules enabled: %d\n", registry.Count())
		for _, rule := range registry.GetRules() {
			fmt.Printf("  - %s\n", rule.Name())
		}
		fmt.Println()
	}

	// Analyze all files
	var allIssues []rules.Issue
	filesWithIssues := 0
	filesAnalyzed := 0

	for _, filePath := range filesToAnalyze {
		issues, err := analyzeFile(filePath, registry, resolver, opts)
		if err != nil {
			// Print error but continue with other files
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", filePath, err)
			continue
		}

		filesAnalyzed++
		if len(issues) > 0 {
			filesWithIssues++
			allIssues = append(allIssues, issues...)
		}
	}

	// Display results
	if len(allIssues) == 0 {
		if !opts.Quiet {
			if len(filesToAnalyze) == 1 {
				printSuccess(fmt.Sprintf("No issues found in %s", filesToAnalyze[0]), opts.NoColor)
			} else {
				printSuccess(fmt.Sprintf("No issues found in %d files", filesAnalyzed), opts.NoColor)
			}
		}
		return 0
	}

	// Print issues grouped by file
	printIssuesGrouped(allIssues, filesAnalyzed, filesWithIssues, opts)

	// Return exit code 1 when issues are found
	return 1
}

// analyzeFile analyzes a single file and returns issues
func analyzeFile(filePath string, registry *rules.Registry, resolver *analyzer.ModuleResolver, opts *Options) ([]rules.Issue, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %v", err)
	}

	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize parser: %v", err)
	}
	defer p.Close()

	// Parse file
	ast, err := p.ParseFile(filePath, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %v", err)
	}
	defer ast.Close()

	// Verbose output for single file analysis
	if opts.Verbose && len(content) > 0 {
		hookCount := countHooks(ast)
		if hookCount > 0 {
			fmt.Printf("%s: found %d React hook call(s)\n", filePath, hookCount)
		}
	}

	// Run all registered rules
	issues := registry.RunAll(ast, resolver)

	return issues, nil
}

// findFiles recursively finds all relevant files in a directory
func findFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and node_modules
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file has valid extension
		ext := strings.ToLower(filepath.Ext(path))
		for _, validExt := range validExtensions {
			if ext == validExt {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}

// validateFileExtension checks if a file has a valid extension
func validateFileExtension(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	for _, validExt := range validExtensions {
		if ext == validExt {
			return nil
		}
	}

	return fmt.Errorf("unsupported file type: %s\nSupported extensions: .tsx, .jsx, .ts, .js", ext)
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

// printIssuesGrouped formats and prints rule violations grouped by file
func printIssuesGrouped(issues []rules.Issue, totalFiles, filesWithIssues int, opts *Options) {
	// Group issues by file
	issuesByFile := make(map[string][]rules.Issue)
	for _, issue := range issues {
		issuesByFile[issue.FilePath] = append(issuesByFile[issue.FilePath], issue)
	}

	// Print each file's issues
	for filePath, fileIssues := range issuesByFile {
		if opts.NoColor {
			fmt.Printf("\n%s\n", filePath)
		} else {
			fmt.Printf("\n\033[1m%s\033[0m\n", filePath) // Bold
		}

		for _, issue := range fileIssues {
			// Format: line:column - [rule] message
			location := fmt.Sprintf("  %d:%d", issue.Line, issue.Column+1)

			if opts.NoColor {
				fmt.Printf("%s - [%s] %s\n", location, issue.Rule, issue.Message)
			} else {
				// Gray for location, yellow for rule, white for message
				fmt.Printf("\033[90m%s\033[0m - \033[33m[%s]\033[0m %s\n",
					location, issue.Rule, issue.Message)
			}
		}
	}

	// Print summary
	issueWord := "issue"
	if len(issues) > 1 {
		issueWord = "issues"
	}

	fileWord := "file"
	if filesWithIssues > 1 {
		fileWord = "files"
	}

	cleanFiles := totalFiles - filesWithIssues
	summary := fmt.Sprintf("\n✖ Found %d %s in %d %s", len(issues), issueWord, filesWithIssues, fileWord)
	if cleanFiles > 0 {
		cleanWord := "file"
		if cleanFiles > 1 {
			cleanWord = "files"
		}
		summary += fmt.Sprintf(" (%d %s clean)", cleanFiles, cleanWord)
	}

	if opts.NoColor {
		fmt.Println(summary)
	} else {
		fmt.Printf("\033[31m%s\033[0m\n", summary)
	}
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
