package gonverge

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/dannyhinshaw/converge/internal/logger"
)

// GoFileConverger is a struct that converges multiple Go files into one.
type GoFileConverger struct {
	// MaxWorkers determines the maximum amount
	// of file workers that will be created to
	// process all files in the given directory.
	MaxWorkers int

	// Packages is a list of packages to include in the output.
	// If empty, converger will default to top-level NON-TEST
	// package in the given directory.
	Packages []string

	// Excludes is a list of files to exclude from merging.
	Excludes []string

	// Logger is the logger to use for logging.
	Logger logger.LevelLogger
}

// NewGoFileConverger creates a new GoFileConverger with sensible defaults.
func NewGoFileConverger(opts ...Option) *GoFileConverger {
	gfc := GoFileConverger{
		MaxWorkers: runtime.NumCPU(),
		Logger:     &logger.NoopLogger{},
	}

	for _, opt := range opts {
		opt(&gfc)
	}

	return &gfc
}

// ConvergeFiles converges all Go files in the given directory and
// package into one and writes the result to the given output.
func (gfc *GoFileConverger) ConvergeFiles(ctx context.Context, src string, w io.Writer) error {
	// fpCh is buffered so consumers can finish
	// processing their files after the producer
	// has closed the channel.
	fpCh := make(chan string, gfc.MaxWorkers)
	resCh := make(chan *goFile)
	errCh := make(chan error)
	var wg sync.WaitGroup

	// Start consumer worker pool
	for i := 0; i < gfc.MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			consumer := newFileConsumer(fpCh, resCh, errCh)
			consumer.consume(ctx)
		}()
	}

	// Setup and start producer
	producer := newFileProducer(gfc.Logger, gfc.Excludes, gfc.Packages, fpCh, errCh)
	go producer.produce(src)

	// Wait for all consumers to finish
	// then close the results and errors channels.
	// This will cause the buildFile function to return.
	go func() {
		wg.Wait()
		close(resCh)
	}()

	// Build the Go file from the results.
	outFile, err := gfc.buildFile(ctx, errCh, resCh)
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
func (gfc *GoFileConverger) buildFile(ctx context.Context, errCh <-chan error, resCh <-chan *goFile) (*goFile, error) {
	gf := newGoFile()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err, ok := <-errCh:
			if !ok {
				return gf, nil
			}
			return nil, err
		case f, ok := <-resCh:
			if !ok {
				return gf, nil
			}
			gf.merge(f)
		}
	}
}
