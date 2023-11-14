package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileConverger is a type that can converge multiple files into one.
type FileConverger interface {
	// ConvergeFiles converges all files in the given directory and
	// package into one and writes the result to the given output.
	ConvergeFiles(ctx context.Context, src string, w io.Writer) error
}

// Converge holds all the dependencies for the converge command
// and is used to validate, build, and run the command.
type Converge struct {
	// fc is the file converger to use.
	fc FileConverger

	// src is the directory to read files from.
	src string

	// dst is the destination file for the output,
	// if one was provided.
	dst string

	// writer is the destination for the output.
	writer io.Writer

	// errCh is the channel used to synchronize
	// the exit and return of concurrent validation
	// functions (if an error occurs).
	errCh chan error
}

// NewCommand returns a new Converge Converge.
func NewCommand(fc FileConverger, src, dst string, opts ...Option) *Converge {
	c := Converge{
		fc:  fc,
		src: src,
		dst: dst,
	}
	for _, opt := range opts {
		opt(&c)
	}

	return &c
}

// Run runs the converge command.
func (c *Converge) Run(ctx context.Context) error {
	if err := c.build(); err != nil {
		return fmt.Errorf("failed to build i/o resources: %w", err)
	}
	if err := c.validate(); err != nil {
		return fmt.Errorf("failed to validate i/o resources: %w", err)
	}

	return c.fc.ConvergeFiles(ctx, c.src, c.writer)
}

// build handles getting the full path to the source directory and destination file (if supplied).
// If the destination file is not supplied, it will validate the writer provided is not nil.
// If the writer is nil, it will default to os.Stdout.
//
// In the case where both the destination file and writer are supplied, the destination file
// will take precedence and the writer will be ignored.
//
// It's important that build is ran before validate, because validate depends on the full path
// to the source directory and destination file (if supplied).
func (c *Converge) build() error {
	var err error
	if c.src, err = filepath.Abs(c.src); err != nil {
		return fmt.Errorf("failed to get absolute path to source directory %s: %w", c.src, err)
	}

	// No dest file or writer supplied,
	// just default to os.Stdout.
	if c.dst == "" {
		if c.writer == nil {
			c.writer = os.Stdout
			return nil
		}
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
func (c *Converge) validate() error {
	vs := []struct {
		fn  func(string)
		arg string
	}{
		{fn: c.validateSrcDir, arg: c.src},
		{fn: c.validateDstFile, arg: c.dst},
	}

	// Set up a buffered channel to receive errors,
	// the buffer size is the number of validation
	// functions we have so none of them block.
	c.errCh = make(chan error, len(vs))

	// Iterate over the validation functions and
	// kick each off in their own goroutine.
	for _, v := range vs {
		go v.fn(v.arg)
	}

	// Wait for errors to come in...
	for i := 0; i < len(vs); i++ {
		err := <-c.errCh
		if err != nil {
			return err
		}
	}

	return nil
}

// validateSrcDir makes sure that the source directory exists, is a directory,
// and that we have permission to read from it.
func (c *Converge) validateSrcDir(src string) {
	switch srcInfo, err := os.Stat(src); {
	case err != nil && !os.IsNotExist(err):
		c.errCh <- fmt.Errorf("failed to access source %s: %w", c.src, err)
	case err == nil && !srcInfo.IsDir():
		c.errCh <- fmt.Errorf("source %s is not a directory", c.src)
	default:
		c.errCh <- nil
	}
}

// validateDstFile makes sure that if the destination file already exists;
// it is not a directory, and we have permission to write to it.
func (c *Converge) validateDstFile(dst string) {
	switch dstInfo, err := os.Stat(dst); {
	case err != nil && !os.IsNotExist(err):
		c.errCh <- fmt.Errorf("failed to access destination file %s: %w", c.dst, err)
	case err == nil && dstInfo.IsDir():
		c.errCh <- fmt.Errorf("destination file %s is a directory", c.dst)
	default:
		c.errCh <- nil
	}
}

// Option is a function that configures a Converge.
type Option func(*Converge)

// WithWriter sets the writer to use for the output.
func WithWriter(w io.Writer) Option {
	return func(c *Converge) {
		c.writer = w
	}
}
