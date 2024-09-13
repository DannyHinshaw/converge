// Package gonverge provides a tool for merging multiple Go files into one.
// This is particularly useful for simplifying Go codebases by combining
// related Go source files while preserving proper package structure and imports.
package gonverge

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"sync"

	"github.com/dannyhinshaw/converge/internal/olog"
)

// maxWorkers is the maximum amount of workers to use for processing files.
const maxWorkers = 32

// debugLogger represents a logger that only logs debug messages.
type debugLogger interface {
	// Debugf logs a formatted debug message.
	Debugf(format string, v ...any)

	// Debug logs a debug message.
	Debug(v ...any)

	// WithName returns a new logger with the given name.
	WithName(name string) olog.LevelLogger
}

// GoFileConverger is responsible for merging multiple Go source files
// into a single file. It uses a worker pool to process files in parallel,
// respecting exclusion patterns and logging progress. The result is a single,
// well-formatted Go file.
type GoFileConverger struct {
	// workers determines the maximum amount
	// of file workers that will be created to
	// process all files in the given directory.
	workers int

	// exclude is a map of regular expressions
	// to apply to file names for exclusion.
	exclude map[string]regexp.Regexp

	// lg is the logger to use for logging.
	lg debugLogger

	// fpCh is the channel to send file paths to.
	// fpCh is buffered so consumers can finish
	// processing their files after the producer
	// has closed the channel.
	fpCh chan string

	// resCh is the channel that delivers processed files.
	resCh chan *goFile

	// errCh is the channel that concurrent functions
	// can send errors to for
	errCh chan error
}

// NewGoFileConverger creates a new GoFileConverger with sensible defaults,
// including a no-op logger. To enable logging, use the WithLogger option.
func NewGoFileConverger(opts ...Option) *GoFileConverger {
	// Upper limit of 32 workers.
	workers := runtime.NumCPU()
	if workers > maxWorkers {
		workers = maxWorkers
	}

	gfc := GoFileConverger{
		workers: workers,
		exclude: make(map[string]regexp.Regexp),
		fpCh:    make(chan string, workers),
		resCh:   make(chan *goFile),
		errCh:   make(chan error),
		lg:      olog.NewNoopLogger(),
	}

	for _, opt := range opts {
		opt(&gfc)
	}

	return &gfc
}

// Option is a functional option for the GoFileConverger.
type Option func(*GoFileConverger)

// WithLogger sets the logger for the GoFileConverger.
func WithLogger(lg debugLogger) Option {
	return func(gfc *GoFileConverger) {
		gfc.lg = lg
	}
}

// WithExcludes allows the caller to specify a list of regular
// expressions that define which files should be excluded from
// the merging process. This is useful for excluding test files
// or specific files in a directory.
func WithExcludes(excludes []regexp.Regexp) Option {
	return func(gfc *GoFileConverger) {
		for _, e := range excludes {
			gfc.exclude[e.String()] = e
		}
	}
}

// WithMaxWorkers sets the maximum amount of workers to use and
// adjusts the file producer channel accordingly.
func WithMaxWorkers(maxWorkers int) Option {
	return func(gfc *GoFileConverger) {
		gfc.workers = maxWorkers
		gfc.fpCh = make(chan string, maxWorkers)
	}
}

// ConvergeFiles converges all Go files in the given directory and
// package into one and writes the result to the given output.
func (c *GoFileConverger) ConvergeFiles(ctx context.Context, dir string, w io.Writer) error {
	var (
		producerWG sync.WaitGroup
		consumerWG sync.WaitGroup
	)

	lg := c.lg.WithName("ConvergeFiles")

	// Start consumer worker pool
	lg.Debugf("Starting %d consumer workers", c.workers)
	for range c.workers {
		consumerWG.Add(1)
		go func() {
			defer consumerWG.Done()
			consumer := newFileConsumer(c.fpCh, c.resCh, c.errCh)
			consumer.consume(ctx)
		}()
	}

	// Setup and start producer
	lg.Debugf("Producing files in directory: %s", dir)
	producerWG.Add(1)
	go func() {
		defer producerWG.Done()
		defer close(c.fpCh) // Close only after producer is done

		producer := newFileProducer(c.lg, c.exclude, c.fpCh, c.errCh)

		c.lg.Debugf("Starting file producer for directory: %s", dir)
		producer.produce(dir)
	}()

	// Wait for the producer and consumers
	// to finish before closing channels
	go func() {
		producerWG.Wait()
		consumerWG.Wait()
		close(c.resCh)
		close(c.errCh)
	}()

	// Build the Go file from the results.
	outFile, err := c.buildFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to buildFile file converger: %w", err)
	}

	// Build and format the output.
	outBytes, err := outFile.FormatCode()
	if err != nil {
		return fmt.Errorf("failed to format code: %w", err)
	}

	// Write the output.
	_, err = w.Write(outBytes)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// buildFile handles running the converger and returning the result or an error.
func (c *GoFileConverger) buildFile(ctx context.Context) (*goFile, error) {
	gf := newGoFile()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err, ok := <-c.errCh:
			if !ok {
				return gf, nil
			}
			return nil, err
		case f, ok := <-c.resCh:
			if !ok {
				return gf, nil
			}
			gf.merge(f)
		}
	}
}
