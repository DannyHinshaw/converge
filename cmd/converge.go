package cmd

import (
	"context"
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
	ConvergeFiles(ctx context.Context, src string, w io.Writer) error
}

// Converge holds all the dependencies for the converge command
// and is used to validate, build, and run the command.
type Converge struct {
	mu sync.Mutex

	// fc is the file converger to use.
	fc FileConverger

	// src is the directory to read files from.
	src string

	// dst is the destination file for the output,
	// if one was provided.
	dst string

	// writer is the destination for the output.
	writer io.Writer
}

// NewCommand returns a new Converge Converge.
func NewCommand(fc FileConverger, src string, opts ...Option) *Converge {
	c := Converge{
		fc:  fc,
		src: src,
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
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	if c.src, err = filepath.Abs(c.src); err != nil {
		return fmt.Errorf("failed to get absolute path to source directory %s: %w", c.src, err)
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
func (c *Converge) validate() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	errCh := make(chan error)

	go func() {
		defer close(errCh)
		if err := validateSrcDir(c.src); err != nil {
			errCh <- err
			return
		}
		if err := validateDstFile(c.dst); err != nil {
			errCh <- err
		}
	}()

	if err, ok := <-errCh; ok {
		return err
	}

	return nil
}

// validateSrcDir makes sure that the source directory exists, is a directory,
// and that we have permission to read from it.
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

// validateDstFile makes sure that if the destination file already exists;
// it is not a directory, and we have permission to write to it.
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

// Option is a function that configures a Converge.
type Option func(*Converge)

// WithWriter sets the writer to use for the output.
func WithWriter(w io.Writer) Option {
	return func(c *Converge) {
		c.writer = w
	}
}

// WithDstFile sets the destination file to use for the output.
func WithDstFile(dst string) Option {
	return func(c *Converge) {
		c.dst = dst
	}
}
