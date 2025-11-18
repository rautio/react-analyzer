package cli

import (
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/rautio/react-analyzer/internal/analyzer"
	"github.com/rautio/react-analyzer/internal/rules"
)

// FileJob represents a file to analyze
type FileJob struct {
	FilePath string
	Index    int // Preserve order for deterministic output
}

// FileResult contains analysis results for a file
type FileResult struct {
	FilePath        string
	Issues          []rules.Issue
	ParseDuration   time.Duration
	AnalyzeDuration time.Duration
	Error           error
	Index           int // For sorting results
}

// analyzeFiles processes files using a worker pool.
// This unified function handles both sequential (workers=1) and parallel (workers>1) cases.
func analyzeFiles(
	filePaths []string,
	registry *rules.Registry,
	resolver *analyzer.ModuleResolver,
	opts *Options,
) ([]*FileResult, time.Duration, time.Duration) {

	// Determine worker count
	numWorkers := opts.Workers
	if numWorkers <= 0 {
		// Auto-detect: use number of CPUs
		numWorkers = runtime.NumCPU()
	}
	// Cap at number of files (no point having more workers than files)
	if numWorkers > len(filePaths) {
		numWorkers = len(filePaths)
	}

	// Create channels
	jobs := make(chan FileJob, len(filePaths))
	results := make(chan *FileResult, len(filePaths))

	// Start worker pool
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go worker(jobs, results, registry, resolver, opts, &wg)
	}

	// Send all jobs
	for i, path := range filePaths {
		jobs <- FileJob{FilePath: path, Index: i}
	}
	close(jobs)

	// Wait for completion in background
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	fileResults := make([]*FileResult, 0, len(filePaths))
	var totalParseDuration time.Duration
	var totalAnalyzeDuration time.Duration

	for result := range results {
		fileResults = append(fileResults, result)
		if result.Error == nil {
			totalParseDuration += result.ParseDuration
			totalAnalyzeDuration += result.AnalyzeDuration
		}
	}

	// Sort by original index for deterministic output
	sort.Slice(fileResults, func(i, j int) bool {
		return fileResults[i].Index < fileResults[j].Index
	})

	return fileResults, totalParseDuration, totalAnalyzeDuration
}

// worker processes jobs from the jobs channel
func worker(
	jobs <-chan FileJob,
	results chan<- *FileResult,
	registry *rules.Registry,
	resolver *analyzer.ModuleResolver,
	opts *Options,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for job := range jobs {
		issues, parseDuration, analyzeDuration, err := analyzeFile(
			job.FilePath,
			registry,
			resolver,
			opts,
		)

		results <- &FileResult{
			FilePath:        job.FilePath,
			Issues:          issues,
			ParseDuration:   parseDuration,
			AnalyzeDuration: analyzeDuration,
			Error:           err,
			Index:           job.Index,
		}
	}
}
