package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dannyhinshaw/converge/cmd"
	"github.com/dannyhinshaw/converge/gonverge"
	"github.com/dannyhinshaw/converge/internal/version"
)

// getUsage returns the usage message for the command.
func getUsage() string {
	return `

        ┏┏┓┏┓┓┏┏┓┏┓┏┓┏┓
        ┗┗┛┛┗┗┛┗ ┛ ┗┫┗
               	    ┛

Usage: converge <source-directory> [options]

Converge multiple files in a Go package into one.

Arguments:
	<source-directory>   Path to the directory containing Go source files to be merged.

Options:
	-f, --file           Path to the output file where the merged content will be written;
	                     defaults to stdout if not specified.
	-v                   Enable verbose logging for debugging purposes.
	-h, --help           Show this help message and exit.
	--version            Show version information.
	-t, --timeout        Maximum time (in seconds) before cancelling the merge operation;
	                     if not specified, the command runs until completion.
	-w, --workers        Maximum number of concurrent workers in the worker pool.
	-e, --exclude        Comma-separated list of filenames to exclude from merging.

Examples:
    converge ./src ./merged.go                                Merge all Go files in the 'src' directory into 'merged.go'
    converge -v ./src ./merged.go                             Merge with verbose logging enabled.
    converge -t 60 ./src ./merged.go                          Merge with a timeout of 60 seconds.
    converge -w 4 ./src ./merged.go                           Merge using a maximum of 4 workers.
    converge -e "file1.go,file2.go" ./src ./merged.go         Merge while excluding 'file1.go' and 'file2.go'.

Note:
	The tool does not merge files in subdirectories and ignores test files (_test.go).`
}

func main() {
	var (
		outFile     string
		exclude     string
		workers     int
		timeout     int
		verboseLog  bool
		showVersion bool
	)

	// outFile flag (short and long version)
	flag.StringVar(&outFile, "f", "", "Output file for merged content; defaults to stdout if not specified")
	flag.StringVar(&outFile, "file", "", "")

	// exclude flag (short and long version)
	flag.StringVar(&exclude, "e", "", "Comma-separated list of files to exclude from merging")
	flag.StringVar(&exclude, "exclude", "", "")

	// workers flag (short and long version)
	flag.IntVar(&workers, "w", 0, "Maximum number of workers to use for file processing")
	flag.IntVar(&workers, "workers", 0, "")

	// timeout flag (short and long version)
	flag.IntVar(&timeout, "t", 0, "Maximum time in seconds before cancelling the operation")
	flag.IntVar(&timeout, "timeout", 0, "")

	// Verbose flag (short version only)
	flag.BoolVar(&verboseLog, "v", false, "Enable verbose logging")

	// Version flag (long version only)
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, getUsage())
		flag.PrintDefaults()
	}

	flag.Parse()

	// Handle version flag
	if showVersion {
		fmt.Println("converge version:", version.GetVersion()) //nolint:forbidigo // only want to print version
		return
	}

	// Verbose logging
	if verboseLog {
		log.SetFlags(0)
		log.Println("Verbose logging enabled")
	}

	// Check for at least one positional argument (source directory)
	if flag.NArg() < 1 {
		log.Println("Error: Source directory is required.")
		flag.Usage()
		os.Exit(1)
	}
	source := flag.Arg(0) // First positional argument

	// Create context to handle timeouts
	t := time.Duration(timeout) * time.Second
	ctx, cancel := newCancelContext(context.Background(), t)
	defer cancel()

	// Build options for the converger
	var gonvOpts []gonverge.Option
	if workers > 0 {
		gonvOpts = append(gonvOpts, gonverge.WithMaxWorkers(workers))
	}
	if exclude != "" {
		excludeFiles := strings.Split(exclude, ",")
		gonvOpts = append(gonvOpts, gonverge.WithExcludes(excludeFiles))
	}
	converger := gonverge.NewGoFileConverger(gonvOpts...)

	// Build options for the command
	var cmdOpts []cmd.Option
	if outFile != "" {
		cmdOpts = append(cmdOpts, cmd.WithDstFile(outFile))
	}

	command := cmd.NewCommand(converger, source, cmdOpts...)
	if err := command.Run(ctx); err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	log.Printf("Files from '%s' have been successfully merged into '%s'.\n", source, outFile)
}

// newCancelContext returns a new cancellable context
// with the given timeout if one is specified.
//
// If no timeout is specified, the context will not have a
// timeout, but a cancel function will still be returned.
func newCancelContext(ctx context.Context, t time.Duration) (context.Context, context.CancelFunc) {
	if t > 0 {
		return context.WithTimeout(ctx, t)
	}
	return context.WithCancel(ctx)
}
