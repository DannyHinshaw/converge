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
)

var usageMessage = `Usage: converge [options] <source-directory> <destination-file>

Converges multiple Go source files into a single file. It scans the specified source directory for Go files, 
merges their contents, and writes the result to the destination file.

Arguments:
  <source-directory>    Path to the directory containing Go source files to be merged.
  <destination-file>    Path to the file where the merged output will be written.

Options:
  -v, --verbose         Enable verbose logging for debugging purposes.
  -h, --help            Show this help message and exit.
  --version             Show version information.
  -t, --timeout         Maximum time (in seconds) before cancelling the merge operation. 
                        If not specified, the command runs until completion.
  -w, --workers         Maximum number of concurrent workers in the worker pool.
  -e, --exclude         Comma-separated list of filenames to exclude from merging.

Examples:
  converge ./src ./merged.go                                Merge all Go files in the 'src' directory into 'merged.go'.
  converge --verbose ./src ./merged.go                      Merge with verbose logging enabled.
  converge -t 60 ./src ./merged.go                          Merge with a timeout of 60 seconds.
  converge -w 4 ./src ./merged.go                           Merge using a maximum of 4 workers.
  converge -e "file1.go,file2.go" ./src ./merged.go         Merge while excluding 'file1.go' and 'file2.go'.

Note:
  The tool does not merge files in subdirectories and ignores test files (_test.go).`

func main() {
	var (
		source  string
		output  string
		exclude string
		workers int
		timeout int
		verbose bool
	)

	flag.StringVar(&source, "src", ".", "Source directory to scan for Go files")
	flag.StringVar(&output, "out", "", "Output file for merged content; defaults to stdout if not specified")
	flag.StringVar(&exclude, "exclude", "", "Comma-separated list of files to exclude from merging")
	flag.IntVar(&timeout, "timeout", 0, "Maximum time in seconds before cancelling the operation")
	flag.IntVar(&workers, "workers", 0, "Maximum number of workers to use for file processing")
	flag.BoolVar(&verbose, "v", false, "Enable verbose logging")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usageMessage)
		flag.PrintDefaults()
	}
	flag.Parse()

	// Verbose logging
	if verbose {
		log.SetFlags(0)
		log.Println("Verbose logging enabled")
	}

	// TODO: Nuke this?
	// Handle invalid arguments
	//if flag.NArg() > 0 {
	//	flag.Usage()
	//	log.Println("HEREEE????")
	//	os.Exit(1)
	//}

	// Create context to handle timeouts
	t := time.Duration(timeout) * time.Second
	ctx, cancel := newCancelContext(context.Background(), t)
	defer cancel()

	// Build options for the converger
	var opts []gonverge.Option
	if workers > 0 {
		opts = append(opts, gonverge.WithMaxWorkers(workers))
	}
	if exclude != "" {
		excludeFiles := strings.Split(exclude, ",")
		opts = append(opts, gonverge.WithExcludeList(excludeFiles))
	}

	// Perform the converge operation
	converger := gonverge.NewGoFileConverger(opts...)
	command := cmd.NewCommand(converger, source, output)
	if err := command.Run(ctx); err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Printf("Files from '%s' have been successfully merged into '%s'.\n", source, output)
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
