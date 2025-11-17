package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/parser"
	"github.com/rautio/react-analyzer/internal/rules"
)

// validExtensions are the file extensions we analyze
var validExtensions = []string{".tsx", ".jsx", ".ts", ".js"}

// Options contains CLI configuration
type Options struct {
	Verbose bool
	Quiet   bool
	NoColor bool
}

// AnalysisStats holds metrics about the analysis run
type AnalysisStats struct {
	FilesAnalyzed   int
	FilesWithIssues int
	FilesClean      int
	TotalIssues     int
	Duration        time.Duration

	// Verbose-only stats
	ParseDuration   time.Duration
	AnalyzeDuration time.Duration
	RuleStats       map[string]*RuleStats
}

// RuleStats holds per-rule execution metrics
type RuleStats struct {
	Name        string
	IssuesFound int
	Duration    time.Duration
}

// Run executes the analysis and returns exit code
func Run(path string, opts *Options) int {
	// Start timing
	startTime := time.Now()

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

	// Initialize stats
	stats := &AnalysisStats{
		RuleStats: make(map[string]*RuleStats),
	}

	// Analyze all files
	var allIssues []rules.Issue
	var totalParseDuration time.Duration
	var totalAnalyzeDuration time.Duration

	for _, filePath := range filesToAnalyze {
		issues, parseDuration, analyzeDuration, err := analyzeFile(filePath, registry, resolver, opts)
		if err != nil {
			// Print error but continue with other files
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", filePath, err)
			continue
		}

		totalParseDuration += parseDuration
		totalAnalyzeDuration += analyzeDuration

		stats.FilesAnalyzed++
		if len(issues) > 0 {
			stats.FilesWithIssues++
			allIssues = append(allIssues, issues...)
		}
	}

	// Collect final stats
	stats.FilesClean = stats.FilesAnalyzed - stats.FilesWithIssues
	stats.TotalIssues = len(allIssues)
	stats.Duration = time.Since(startTime)
	stats.ParseDuration = totalParseDuration
	stats.AnalyzeDuration = totalAnalyzeDuration

	// Collect per-rule stats (for verbose mode)
	if opts.Verbose {
		collectRuleStats(allIssues, stats)
	}

	// Display results
	if len(allIssues) == 0 {
		if !opts.Quiet {
			if len(filesToAnalyze) == 1 {
				printSuccess(fmt.Sprintf("No issues found in %s", filesToAnalyze[0]), opts.NoColor)
			} else {
				printSuccess(fmt.Sprintf("No issues found in %d files", stats.FilesAnalyzed), opts.NoColor)
			}
		}
		// Print timing even when no issues (unless quiet)
		if !opts.Quiet {
			printTiming(stats, opts)
		}
		return 0
	}

	// Print issues grouped by file
	printIssuesGrouped(allIssues, stats, opts)

	// Return exit code 1 when issues are found
	return 1
}

// analyzeFile analyzes a single file and returns issues with timing metrics
func analyzeFile(filePath string, registry *rules.Registry, resolver *analyzer.ModuleResolver, opts *Options) ([]rules.Issue, time.Duration, time.Duration, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("cannot read file: %v", err)
	}

	// Create parser
	p, err := parser.NewParser()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to initialize parser: %v", err)
	}
	defer p.Close()

	// Parse file with timing
	parseStart := time.Now()
	ast, err := p.ParseFile(filePath, content)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to parse file: %v", err)
	}
	defer ast.Close()
	parseDuration := time.Since(parseStart)

	// Verbose output for single file analysis
	if opts.Verbose && len(content) > 0 {
		hookCount := countHooks(ast)
		if hookCount > 0 {
			fmt.Printf("%s: found %d React hook call(s)\n", filePath, hookCount)
		}
	}

	// Run all registered rules with timing
	analyzeStart := time.Now()
	issues := registry.RunAll(ast, resolver)
	analyzeDuration := time.Since(analyzeStart)

	return issues, parseDuration, analyzeDuration, nil
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
func printIssuesGrouped(issues []rules.Issue, stats *AnalysisStats, opts *Options) {
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
	if stats.FilesWithIssues > 1 {
		fileWord = "files"
	}

	summary := fmt.Sprintf("\n✖ Found %d %s in %d %s", len(issues), issueWord, stats.FilesWithIssues, fileWord)
	if stats.FilesClean > 0 {
		cleanWord := "file"
		if stats.FilesClean > 1 {
			cleanWord = "files"
		}
		summary += fmt.Sprintf(" (%d %s clean)", stats.FilesClean, cleanWord)
	}

	if opts.NoColor {
		fmt.Println(summary)
	} else {
		fmt.Printf("\033[31m%s\033[0m\n", summary)
	}

	// Print timing (default mode)
	if !opts.Quiet {
		printTiming(stats, opts)
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

// printTiming prints basic timing information (default mode)
func printTiming(stats *AnalysisStats, opts *Options) {
	fileWord := "file"
	if stats.FilesAnalyzed != 1 {
		fileWord = "files"
	}

	fmt.Printf("Analyzed %d %s in %s\n",
		stats.FilesAnalyzed,
		fileWord,
		formatDuration(stats.Duration))

	// Verbose: Show detailed breakdown
	if opts.Verbose {
		printDetailedStats(stats, opts)
	}
}

// printDetailedStats prints detailed statistics (verbose mode)
func printDetailedStats(stats *AnalysisStats, opts *Options) {
	fmt.Println("\nPerformance Summary:")
	fmt.Printf("  Time elapsed: %s (parse: %s, analyze: %s)\n",
		formatDuration(stats.Duration),
		formatDuration(stats.ParseDuration),
		formatDuration(stats.AnalyzeDuration))

	if stats.FilesAnalyzed > 0 {
		filesPerSec := float64(stats.FilesAnalyzed) / stats.Duration.Seconds()
		fmt.Printf("  Throughput: %.0f files/sec\n", filesPerSec)
	}

	// Show per-rule stats if available
	if len(stats.RuleStats) > 0 {
		fmt.Println("\nRules executed:")
		for _, ruleStats := range stats.RuleStats {
			issueWord := "issue"
			if ruleStats.IssuesFound != 1 {
				issueWord = "issues"
			}
			fmt.Printf("  %s: %d %s\n",
				ruleStats.Name,
				ruleStats.IssuesFound,
				issueWord)
		}
	}
}

// collectRuleStats collects per-rule statistics from issues
func collectRuleStats(issues []rules.Issue, stats *AnalysisStats) {
	// Count issues per rule
	ruleCounts := make(map[string]int)
	for _, issue := range issues {
		ruleCounts[issue.Rule]++
	}

	// Create RuleStats entries
	for ruleName, count := range ruleCounts {
		stats.RuleStats[ruleName] = &RuleStats{
			Name:        ruleName,
			IssuesFound: count,
		}
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.0fμs", float64(d.Microseconds()))
	} else if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	} else {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}
