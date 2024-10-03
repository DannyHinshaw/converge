// Package cmd provides the main entry point for the converge CLI tool.
// It sets up the necessary command-line flags, handles user input, and
// triggers the converge operation with the given options.
// The package also manages logging, error handling, and execution flow.
package cmd

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"

	"github.com/dannyhinshaw/converge/cmd/converge"
	"github.com/dannyhinshaw/converge/internal/gonverge"
	"github.com/dannyhinshaw/converge/internal/olog"
)

// dopeASCII is just a dope ASCII art string.
const dopeASCII = `
┏┏┓┏┓┓┏┏┓┏┓┏┓┏┓
┗┗┛┛┗┗┛┗ ┛ ┗┫┗
            ┛

`

// defaultTimeout is the default amount of time before
// cancelling the converge operation.
const defaultTimeout = 15 * time.Second

// usageTemplate is a utility function for replacing the default usage
// template with any custom usage template in a central location.
func usageTemplate(c cobra.Command) string {
	return dopeASCII + c.UsageTemplate()
}

// NewRoot creates the root command for the converge CLI tool.
// It sets up the command-line flags, initializes logging, and
// runs the converge operation with the provided options.
func NewRoot(version string) *cobra.Command {
	var rootCmd cmd
	c := cobra.Command{
		Version: version,
		Use:     "converge [flags]",
		Short:   "Merge multiple Go source files into a single file",
		Long: `
Converge merges multiple Go source files from a directory into a single file.

By default, the tool does not process directories recursively. You can specify 
the source directory with the --dir flag and an output file using --output. If 
no output file is provided, the result will be printed to stdout. You can exclude
files by providing regular expressions with the --exclude flag.

The result is formatted according to Go's standard "gofmt" style.
`,
		Args: cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, _ []string) {
			ctx, cancel := context.WithTimeout(cmd.Context(), rootCmd.timeout)
			defer cancel()

			// Default to only logging errors.
			lvl := olog.LevelError
			if rootCmd.verbose {
				lvl = olog.LevelDebug
			}

			lg := olog.NewLogger(lvl).
				WithName("converge")

			lg.Info("Starting converge operation...")
			lg.Debug("Verbose logging enabled.")

			rootCmd.lg = lg.WithName("rootCmd")
			if err := rootCmd.run(ctx); err != nil {
				lg.Error("failed to run command:", err)
				return
			}

			// Only print success message if an outfile was provided.
			// This is to prevent the success message from being printed
			// when the converged code output is written to stdout.
			if rootCmd.outfile != "" {
				lg.Info("Converge operation completed successfully.")
			}
		},
	}

	bindFlags(&c, &rootCmd)
	c.SetUsageTemplate(
		usageTemplate(c),
	)

	return &c
}

// bindFlags handles binding the command-line flags to the root command.
func bindFlags(c *cobra.Command, rootCmd *cmd) {
	fs := c.Flags()

	fs.StringVarP(&rootCmd.dir,
		"dir", "d", ".",
		"The directory containing Go files to merge",
	)
	fs.StringVarP(&rootCmd.outfile,
		"output", "o", "",
		"File to write the merged Go code (default: stdout)",
	)
	fs.StringSliceVarP(&rootCmd.exclude,
		"exclude", "e", nil,
		"Regular expressions for filenames to exclude from merging",
	)
	fs.DurationVarP(&rootCmd.timeout,
		"timeout", "t", defaultTimeout,
		"Maximum duration before canceling the operation (e.g., '5s', '1m')",
	)
	fs.BoolVarP(&rootCmd.verbose,
		"verbose", "v", false,
		"Enable verbose logging for debugging purposes",
	)
	// Note(@danny): In the future add a flag that allows users
	// to configure words to replace in the converged file.
	// Also, add ability to remove duplicate imports, types,
	// functions (etc), from the converged file.
}

// cmd holds the command-line options and utilities
// for running the converge command.
type cmd struct {
	// lg is the logger for the command.
	lg olog.LevelLogger

	// dir is the source directory containing
	// Go source files to be converged.
	dir string

	// outfile is the path to the output file where the
	// converged content will be written; defaults to
	// stdout if not specified.
	outfile string

	// exclude is a list of regex patterns to be used for
	// excluding files from converge if they match.
	exclude []string

	// timeout is the maximum time (in seconds) before
	// cancelling the converge operation.
	timeout time.Duration

	// verbose enables verbose logging
	// for debugging purposes.
	verbose bool
}

// run executes the converge command.
func (c *cmd) run(ctx context.Context) error {
	c.lg.Debug("Starting converge command")

	// Create the converger that will handle
	// the low level processing of the files.
	converger, err := createConverger(c.lg.WithName("converger"), c.exclude)
	if err != nil {
		return fmt.Errorf("failed to create converger: %w", err)
	}

	// Create the command that will run the converger
	// and write the output to the specified file.
	convergeCmd := createCommand(converger, c.dir, c.outfile)
	if err = convergeCmd.Run(ctx); err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	c.lg.Debug("Converge command completed successfully.")

	if c.outfile != "" {
		c.lg.Infof("Successfully merged '%s' into '%s'.", c.dir, c.outfile)
	}

	return nil
}

// createCommand creates a new converge.Command with the given options.
func createCommand(converger converge.FileConverger, dir, outFile string) *converge.Command {
	var cmdOpts []converge.Option
	if outFile != "" {
		cmdOpts = append(cmdOpts, converge.WithDstFile(outFile))
	}
	return converge.NewCommand(converger, dir, cmdOpts...)
}

// createConverger creates a new gonverge.GoFileConverger by handling
// which options to set and passed into the converger.
func createConverger(lg olog.LevelLogger, ex []string) (*gonverge.GoFileConverger, error) {
	var gonvOpts []gonverge.Option
	if lg != nil {
		gonvOpts = append(gonvOpts, gonverge.WithLogger(
			lg.WithName("gonverge"),
		))
	}

	var excludes []regexp.Regexp
	for _, e := range ex {
		re, err := regexp.Compile(e)
		if err != nil {
			return nil, fmt.Errorf("failed to compile regex: %w", err)
		}
		excludes = append(excludes, *re)
	}
	if len(excludes) > 0 {
		gonvOpts = append(gonvOpts, gonverge.WithExcludes(excludes))
	}

	return gonverge.NewGoFileConverger(gonvOpts...), nil
}
