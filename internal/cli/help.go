package cli

import (
	"fmt"
	"io"
)

// PrintHelp displays the help message
func PrintHelp(w io.Writer) {
	help := `react-analyzer v0.1.0

Static analysis tool for React performance issues.

USAGE:
    react-analyzer [OPTIONS] <path>

ARGUMENTS:
    <path>    File or directory to analyze (.tsx, .jsx, .ts, .js)

OPTIONS:
    -h, --help         Show this help message
    -v, --version      Show version number
    -V, --verbose      Show detailed output
    -q, --quiet        Only show errors
        --no-color     Disable colored output

EXAMPLES:
    react-analyzer src/App.tsx              # Analyze single file
    react-analyzer src/                     # Analyze directory
    react-analyzer --verbose .              # Analyze entire project

RULES:
    no-object-deps     Prevent inline objects in hook dependencies
                       Inline objects/arrays in useEffect, useMemo, or
                       useCallback dependencies cause infinite re-renders.

EXIT CODES:
    0    No issues found
    1    Issues found
    2    Analysis error

For more information, visit: https://github.com/your-org/react-analyzer
`
	fmt.Fprint(w, help)
}
