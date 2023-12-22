package cli

import (
	"flag"
	"fmt"
)

// App is the CLI application struct.
type App struct {
	// SrcDir is the path to the source directory
	// containing Go source files to be converged.
	SrcDir string

	// OutFile is the path to the output file where the
	// converged content will be written; defaults to
	// stdout if not specified.
	OutFile string

	// Packages is a list of specific Packages to include
	// in the converged file.
	Packages string

	// Exclude is a comma-separated list of regex patterns
	// that will be used to Exclude files from converging.
	Exclude string

	// Workers is the maximum number of concurrent Workers
	// in the worker pool.
	Workers int

	// Timeout is the maximum time (in seconds) before
	// cancelling the converge operation.
	Timeout int

	// VerboseLog enables verbose logging
	// for debugging purposes.
	VerboseLog bool

	// ShowVersion shows version information and exits.
	ShowVersion bool
}

// NewApp creates a new CLI application struct with the arguments parsed.
func NewApp() App {
	var a App

	// OutFile flag (short and long version)
	flag.StringVar(&a.OutFile, "f", "", "Output file for merged content; defaults to stdout if not specified")
	flag.StringVar(&a.OutFile, "file", "", "")

	// Packages flag (short and long version)
	flag.StringVar(&a.Packages, "p", "", "Comma-separated list of packages to include")
	flag.StringVar(&a.Packages, "pkg", "", "")

	// Exclude flag (short and long version)
	flag.StringVar(&a.Exclude, "e", "", "Comma-separated list of regex patterns to exclude")
	flag.StringVar(&a.Exclude, "exclude", "", "")

	// Workers flag (short and long version)
	flag.IntVar(&a.Workers, "w", 0, "Maximum number of workers to use for file processing")
	flag.IntVar(&a.Workers, "workers", 0, "")

	// Timeout flag (short and long version)
	flag.IntVar(&a.Timeout, "t", 0, "Maximum time in seconds before cancelling the operation")
	flag.IntVar(&a.Timeout, "timeout", 0, "")

	// Verbose flag (short version only)
	flag.BoolVar(&a.VerboseLog, "v", false, "Enable verbose logging")

	// Version flag (long version only)
	flag.BoolVar(&a.ShowVersion, "version", false, "Show version information and exit")

	// Custom usage message
	flag.Usage = func() {
		fmt.Println(a.Usage()) //nolint:forbidigo // not debugging
		flag.PrintDefaults()
	}
	flag.Parse()

	return a
}

// ParseSrcDir parses the source directory from the positional arguments.
func (a App) ParseSrcDir() (string, error) {
	if flag.NArg() < 1 {
		return "", fmt.Errorf("source directory is required")
	}

	return flag.Arg(0), nil
}

// Usage returns the usage help message.
func (a App) Usage() string {
	return `

┏┏┓┏┓┓┏┏┓┏┓┏┓┏┓
┗┗┛┛┗┗┛┗ ┛ ┗┫┗
            ┛

Usage: converge <source-directory> [options]

The converge tool provides ways to 'converge' multiple Go source files into a single file.
By default it does not converge files in subdirectories and ignores test files (_test.go).

Arguments:
	<source-dir>         Path to the directory containing Go source files to be converged.

Options:
    -f, --file           Path to the output file where the converged content will be written;
                         defaults to stdout if not specified.
    -p, --pkg            List of specific packages to include in the converged file.
                         Note that if you converge multiple packages the converged file will
                         not be compilable.
	-t, --timeout        Maximum time (in seconds) before cancelling the converge operation;
                         if not specified, the command runs until completion.
	-w, --workers        Maximum number of concurrent workers in the worker pool.
	-e, --exclude        Comma-separated list of regex patterns to exclude from converging.
	-v                   Enable verbose logging for debugging purposes.
	-h, --help           Show this help message and exit.
	--version            Show version information.

Examples:
    converge . -o converged.go                     All Go files in current dir into 'converged.go'
    converge . -p included_test,included           All Go files with package name included_test or included. 
    converge . -v                                  Run with verbose logging enabled.
    converge . -t 60                               Run with a timeout of 60 seconds.
    converge . -w 4                                Run using a maximum of 4 workers.
    converge . -e "file1.go,pattern(.*).go"        Run while excluding 'file1.go' and 'file2.go'.
`
}
