// Package converge provides the command structure for the "converge" CLI tool.
// This command uses a FileConverger interface to merge multiple files from
// a source directory into a single file. The design allows for future support
// of different file types by implementing additional FileConverger variants.
package converge

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// FileConverger is a type that can converge multiple files into one.
type FileConverger interface {
	// ConvergeFiles converges all files in the given directory and
	// package into one and writes the result to the given output.
	ConvergeFiles(ctx context.Context, dir string, w io.Writer) error
}

// Command holds the configuration and dependencies for the "converge" command.
// If a destination file (dst) is specified, it takes precedence over the writer.
// Otherwise, output defaults to os.Stdout or the provided writer.
type Command struct {
	// dir is the directory to read files from.
	dir string

	// dst is the destination file for the output,
	// if one was provided.
	dst string

	// fc is the file converger to use.
	fc FileConverger

	// writer is the destination for the output.
	writer io.Writer
}

// NewCommand returns a new Command with standard defaults.
func NewCommand(fc FileConverger, dir string, opts ...Option) *Command {
	c := Command{
		fc:     fc,
		dir:    dir,
		dst:    "",
		writer: os.Stdout,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return &c
}

// Option is a function that configures a Command.
type Option func(*Command)

// WithWriter sets the writer to use for the output.
func WithWriter(w io.Writer) Option {
	return func(c *Command) {
		c.writer = w
	}
}

// WithDstFile sets the destination file to use for the output.
func WithDstFile(dst string) Option {
	return func(c *Command) {
		c.dst = dst
	}
}

// Run runs the converge command.
func (c *Command) Run(ctx context.Context) error {
	if err := c.build(); err != nil {
		return fmt.Errorf("failed to build converge command: %w", err)
	}
	if err := c.validate(); err != nil {
		return fmt.Errorf("failed to validate converge command: %w", err)
	}
	if err := c.fc.ConvergeFiles(ctx, c.dir, c.writer); err != nil {
		return fmt.Errorf("failed to converge files: %w", err)
	}
	return nil
}

// build prepares the command for execution by converting paths to absolute paths,
// setting up the writer, and ensuring the output destination is valid.
//
// It must be run before validate since validate depends on these paths.
func (c *Command) build() error {
	var err error
	if c.dir, err = filepath.Abs(c.dir); err != nil {
		return fmt.Errorf("failed to get absolute path to source directory %s: %w", c.dir, err)
	}

	// No dest file or writer supplied,
	// just default to os.Stdout.
	if c.writer == nil {
		c.writer = os.Stdout
	}
	if c.dst == "" {
		return nil
	}

	// Destination file supplied, so we'll need the
	// absolute path to it for validation and writing.
	if c.dst, err = filepath.Abs(c.dst); err != nil {
		return fmt.Errorf("failed to get absolute path to destination file %s: %w", c.dst, err)
	}
	if c.writer, err = os.Create(c.dst); err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", c.dst, err)
	}

	return nil
}

// validate performs an initial check to make sure that the src and dst
// arguments are for objects that actually exist on the users system before kicking
// off the full-blown converge operation.
func (c *Command) validate() error {
	var (
		merr  error
		wg    sync.WaitGroup
		errCh = make(chan error, 1)
	)

	validators := []func(){
		func() {
			defer wg.Done()
			if err := validateSrcDir(c.dir); err != nil {
				errCh <- err
			}
		},
		func() {
			defer wg.Done()
			if err := validateDstFile(c.dst); err != nil {
				errCh <- err
			}
		},
	}

	for _, fn := range validators {
		wg.Add(1)
		go fn()
	}

	wg.Wait()
	close(errCh)

	// Drain the error channel and join
	// all errors into one.
	for err := range errCh {
		merr = errors.Join(merr, err)
	}
	return merr
}

// validateSrcDir checks that the source directory exists, is a directory,
// and that the user has permission to read from it.
func validateSrcDir(src string) error {
	switch srcInfo, err := os.Stat(src); {
	case err != nil && !os.IsNotExist(err):
		return fmt.Errorf("failed to access source %s: %w", src, err)
	case err == nil && !srcInfo.IsDir():
		return fmt.Errorf("source %s is not a directory", src)
	default:
		return nil
	}
}

// validateDstFile ensures that the destination file is not a directory and
// checks for write permissions. If the file doesn't exist, no error is returned.
func validateDstFile(dst string) error {
	switch dstInfo, err := os.Stat(dst); {
	case err != nil && !os.IsNotExist(err):
		return fmt.Errorf("failed to access destination file %s: %w", dst, err)
	case err == nil && dstInfo.IsDir():
		return fmt.Errorf("destination file %s is a directory", dst)
	default:
		return nil
	}
}
