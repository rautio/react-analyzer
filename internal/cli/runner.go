package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/config"
	"github.com/rautio/react-analyzer/internal/graph"
	"github.com/rautio/react-analyzer/internal/parser"
	"github.com/rautio/react-analyzer/internal/rules"
)

// validExtensions are the file extensions we analyze
var validExtensions = []string{".tsx", ".jsx", ".ts", ".js"}

// Options contains CLI configuration
type Options struct {
	Verbose      bool
	Quiet        bool
	NoColor      bool
	Workers      int  // Number of parallel workers (0 = auto-detect CPUs, 1 = sequential)
	JSON         bool // Output results as JSON
	IncludeGraph bool // Include dependency graph in JSON output (requires JSON mode)
	Mermaid      bool // Output Mermaid flowchart diagram
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

// JSONOutput represents the complete JSON output structure
type JSONOutput struct {
	Issues []rules.Issue `json:"issues"`
	Stats  JSONStats     `json:"stats"`
	Graph  *graph.Graph  `json:"graph,omitempty"` // Only included if --graph flag is set
}

// JSONStats represents statistics in JSON format
type JSONStats struct {
	FilesAnalyzed   int     `json:"filesAnalyzed"`
	FilesWithIssues int     `json:"filesWithIssues"`
	FilesClean      int     `json:"filesClean"`
	TotalIssues     int     `json:"totalIssues"`
	DurationMs      float64 `json:"durationMs"`
}

// Run executes the analysis and returns exit code
func Run(path string, opts *Options) int {
	// Start timing
	startTime := time.Now()

	// Convert to absolute path for consistent config/module resolution
	absPath, err := filepath.Abs(path)
	if err != nil {
		printError(fmt.Errorf("failed to resolve path: %v", err), opts.NoColor)
		return 2
	}
	path = absPath

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

	// Determine base directory for config and module resolution
	baseDir := path
	if !info.IsDir() {
		baseDir = filepath.Dir(path)
	}

	// Load configuration from .reactanalyzerrc.json (if exists)
	cfg, configPath, err := config.LoadWithPath(baseDir)
	if err != nil {
		// Config loading error is non-fatal, use defaults
		if opts.Verbose && !opts.JSON {
			fmt.Fprintf(os.Stderr, "Warning: could not load config (%v), using defaults\n", err)
		}
		cfg = config.DefaultConfig()
	} else if opts.Verbose && !opts.JSON {
		if configPath != "" {
			fmt.Printf("Configuration loaded from: %s\n", configPath)
		} else {
			fmt.Println("Configuration: using defaults (no config file found)")
		}
	}

	// Create rule registry with configuration
	registry := rules.NewRegistry(cfg)

	// Determine project root for module resolver
	// If a config file was found, use its directory as the project root
	// This ensures alias resolution works even when analyzing single files
	projectRoot := baseDir
	if configPath != "" {
		projectRoot = filepath.Dir(configPath)
	}

	// Create module resolver for cross-file analysis
	resolver, err := analyzer.NewModuleResolver(projectRoot)
	if err != nil {
		printError(fmt.Errorf("failed to initialize module resolver: %v", err), opts.NoColor)
		return 2
	}
	defer resolver.Close()

	// Verbose: show path aliases loaded
	if opts.Verbose && !opts.JSON {
		aliases, aliasConfigPath := resolver.GetPathAliasesWithSource()
		if len(aliases) > 0 {
			fmt.Printf("Path aliases loaded from: %s\n", aliasConfigPath)
			fmt.Printf("  Found %d alias(es):\n", len(aliases))
			for prefix, target := range aliases {
				fmt.Printf("    %s -> %s\n", prefix, target)
			}
			fmt.Println()
		} else {
			fmt.Printf("No path aliases found (searched from %s up to root)\n\n", baseDir)
		}
	}

	// Print analysis start for directories (not in JSON mode)
	if len(filesToAnalyze) > 1 && !opts.Quiet && !opts.JSON {
		fmt.Printf("Analyzing %d files...\n\n", len(filesToAnalyze))
	}

	if opts.Verbose && !opts.JSON {
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

	// Analyze all files using worker pool (handles both sequential and parallel)
	fileResults, totalParseDuration, totalAnalyzeDuration := analyzeFiles(
		filesToAnalyze,
		registry,
		resolver,
		opts,
	)

	// Process results
	var allIssues []rules.Issue
	for _, result := range fileResults {
		if result.Error != nil {
			// Print error but continue with other files (not in JSON mode)
			if !opts.JSON {
				fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", result.FilePath, result.Error)
			}
			continue
		}

		stats.FilesAnalyzed++
		if len(result.Issues) > 0 {
			stats.FilesWithIssues++
			allIssues = append(allIssues, result.Issues...)
		}
	}

	// Run graph-based rules (if any modules were parsed)
	var depGraph *graph.Graph // Store graph for potential JSON output
	if stats.FilesAnalyzed > 0 {
		if opts.Verbose && !opts.JSON {
			fmt.Println("\nBuilding dependency graph...")
			modules := resolver.GetModules()
			fmt.Printf("  Parsed modules: %d\n", len(modules))
			for path := range modules {
				fmt.Printf("    - %s\n", path)
			}
		}
		graphStart := time.Now()

		// Build the graph from all parsed modules
		builder := graph.NewBuilder(resolver)
		var err error
		depGraph, err = builder.Build()
		if err != nil {
			if !opts.JSON {
				fmt.Fprintf(os.Stderr, "Warning: failed to build dependency graph: %v\n", err)
			}
		} else {
			// Run graph-based rules
			graphIssues := registry.RunGraph(depGraph)
			if len(graphIssues) > 0 {
				allIssues = append(allIssues, graphIssues...)
				// Update stats for graph-based issues
				graphFilesAffected := make(map[string]bool)
				for _, issue := range graphIssues {
					graphFilesAffected[issue.FilePath] = true
				}
				stats.FilesWithIssues += len(graphFilesAffected)
			}

			if opts.Verbose && !opts.JSON {
				fmt.Printf("Graph analysis completed in %s\n", formatDuration(time.Since(graphStart)))
				fmt.Printf("  Components: %d\n", len(depGraph.ComponentNodes))
				fmt.Printf("  State nodes: %d\n", len(depGraph.StateNodes))
				fmt.Printf("  Edges: %d\n", len(depGraph.Edges))

				// Show component details
				if len(depGraph.ComponentNodes) > 0 {
					fmt.Println("\n  Component details:")
					for id, comp := range depGraph.ComponentNodes {
						children := ""
						if len(comp.Children) > 0 {
							children = fmt.Sprintf(", %d children", len(comp.Children))
						}
						fmt.Printf("    %s (file: %s%s)\n", comp.Name, comp.Location.FilePath, children)
						_ = id // Use id to avoid unused variable warning
					}
				}

				// Show prop passing edges
				propEdges := 0
				for _, edge := range depGraph.Edges {
					if edge.Type == "passes" {
						propEdges++
					}
				}
				if propEdges > 0 {
					fmt.Printf("\n  Prop passing edges: %d\n", propEdges)
				}
			}
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

	// Output results in Mermaid format if requested
	if opts.Mermaid {
		return outputMermaid(depGraph, opts)
	}

	// Output results in JSON format if requested
	if opts.JSON {
		return outputJSON(allIssues, stats, depGraph, opts)
	}

	// Display results in text format
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
	// Parse file using resolver (this caches the module for graph building)
	parseStart := time.Now()
	module, err := resolver.GetModule(filePath)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to parse file: %v", err)
	}
	parseDuration := time.Since(parseStart)

	// Verbose output for single file analysis
	if opts.Verbose {
		hookCount := countHooksFromAST(module.AST)
		if hookCount > 0 {
			fmt.Printf("%s: found %d React hook call(s)\n", filePath, hookCount)
		}
	}

	// Run all registered rules with timing
	analyzeStart := time.Now()
	issues := registry.RunAll(module.AST, resolver)
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
func countHooksFromAST(ast *parser.AST) int {
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

// outputJSON outputs analysis results in JSON format
func outputJSON(issues []rules.Issue, stats *AnalysisStats, depGraph *graph.Graph, opts *Options) int {
	// Ensure issues is an empty array instead of nil for consistent JSON
	if issues == nil {
		issues = []rules.Issue{}
	}

	// Build JSON output structure
	output := JSONOutput{
		Issues: issues,
		Stats: JSONStats{
			FilesAnalyzed:   stats.FilesAnalyzed,
			FilesWithIssues: stats.FilesWithIssues,
			FilesClean:      stats.FilesClean,
			TotalIssues:     stats.TotalIssues,
			DurationMs:      float64(stats.Duration.Microseconds()) / 1000.0,
		},
	}

	// Include graph if requested and available
	if opts.IncludeGraph && depGraph != nil {
		output.Graph = depGraph
	}

	// Marshal to JSON with indentation
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal JSON output: %v\n", err)
		return 2
	}

	// Print JSON to stdout
	fmt.Println(string(jsonBytes))

	// Return exit code based on issues found
	if len(issues) > 0 {
		return 1
	}
	return 0
}

// outputMermaid outputs the dependency graph in Mermaid flowchart format
func outputMermaid(depGraph *graph.Graph, opts *Options) int {
	if depGraph == nil {
		fmt.Fprintf(os.Stderr, "Error: no graph available for Mermaid output\n")
		return 2
	}

	// Generate Mermaid diagram
	mermaidDiagram := depGraph.ToMermaid()

	// Print to stdout (and only Mermaid, nothing else)
	fmt.Print(mermaidDiagram)

	return 0
}
