package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/oskari/react-analyzer/internal/cli"
)

const Version = "0.1.0"

func main() {
	// Define flags
	showHelp := flag.Bool("help", false, "")
	flag.BoolVar(showHelp, "h", false, "")

	showVersion := flag.Bool("version", false, "")
	flag.BoolVar(showVersion, "v", false, "")

	verbose := flag.Bool("verbose", false, "")
	flag.BoolVar(verbose, "V", false, "")

	quiet := flag.Bool("quiet", false, "")
	flag.BoolVar(quiet, "q", false, "")

	noColor := flag.Bool("no-color", false, "")

	// Custom usage (will show our help text)
	flag.Usage = func() {
		cli.PrintHelp(os.Stdout)
	}

	flag.Parse()

	// Handle special flags
	if *showHelp {
		cli.PrintHelp(os.Stdout)
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("react-analyzer %s\n", Version)
		os.Exit(0)
	}

	// Validate file argument
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: missing file argument")
		fmt.Fprintln(os.Stderr, "\nUsage: react-analyzer [options] <file>")
		fmt.Fprintln(os.Stderr, "Run 'react-analyzer --help' for more information")
		os.Exit(2)
	}

	filePath := flag.Arg(0)

	// Build options
	opts := &cli.Options{
		Verbose: *verbose,
		Quiet:   *quiet,
		NoColor: *noColor,
	}

	// Run analysis and exit with appropriate code
	exitCode := cli.Run(filePath, opts)
	os.Exit(exitCode)
}
