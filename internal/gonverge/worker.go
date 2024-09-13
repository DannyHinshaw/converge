package gonverge

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// fileProducer walks a directory and sends all file paths
// to the given channel for the consumer to process.
type fileProducer struct {
	// excludes is a map of regular expressions
	// to apply to file names for exclusion.
	excludes map[string]regexp.Regexp

	// lg is the lg to use for logging.
	lg debugLogger

	// fpCh is the channel to send file paths to.
	fpCh chan<- string

	// errCh is the channel to send errors to.
	errCh chan<- error
}

// newFileProducer handles the creation of a new fileProducer.
func newFileProducer(lg debugLogger, ex map[string]regexp.Regexp, fc chan<- string, ec chan<- error) *fileProducer {
	return &fileProducer{
		lg:       lg,
		fpCh:     fc,
		errCh:    ec,
		excludes: ex,
	}
}

// produce walks the given directory and sends all file paths
// to the fpCh channel for the consumer to process.
func (fp *fileProducer) produce(dir string) {
	lg := fp.lg.WithName("walkDir")
	lg.Debug("Producing files in directory:", dir)

	if err := fp.walkDir(dir); err != nil {
		fp.errCh <- fmt.Errorf("error walking directory: %w", err)
	}
}

// walkDir walks the given directory and sends all file paths
// to the fpCh channel for the consumer to process.
func (fp *fileProducer) walkDir(dir string) error {
	lg := fp.lg.WithName("walkDir")
	lg.Debug("Walking directory:", dir)

	return fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error { //nolint:wrapcheck // Low level error doesn't need wrapped any further.
		if err != nil {
			return fmt.Errorf("error walking directory: %w", err)
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("error getting file info: %w", err)
		}

		fullPath := filepath.Join(dir, path)
		if !fp.validFile(info.Name(), fullPath) {
			lg.Debug("file path is not valid:", fullPath)
			return nil
		}

		lg.Debug("file path is valid:", fullPath)
		fp.fpCh <- fullPath

		return nil
	})
}

// validFile checks that the file is a valid *non-test* Go file.
// If the pkgSet is empty, it will default to the top-level
// Go files that are *not* test files.
//
// However, if the pkgSet is not empty, it will include all
// files that have a package name that is in the set.
// This behavior essentially allows for the user to specify
// the package names they want to include, including test files
// with the package name in the set.
func (fp *fileProducer) validFile(name, fullPath string) bool {
	lg := fp.lg.WithName("validFile")
	lg.Debugf("Validating package %s at: %s", name, fullPath)

	if !strings.HasSuffix(name, ".go") {
		return false
	}

	// Check if the file should be excluded from processing.
	for _, re := range fp.excludes {
		if re.MatchString(name) {
			lg.Debug("File excluded from processing:", name)
			return false
		}
	}

	return true
}

// fileConsumer reads file paths from the given channel,
// processes them, and then sends back either the processed
// result or an error (if one occurred).
type fileConsumer struct {
	// fpCh is the channel to read file paths from.
	fpCh <-chan string

	// resCh is the channel to send processed files to.
	resCh chan<- *goFile

	// errCh is the channel to send errors to.
	errCh chan error
}

// newFileConsumer returns a new fileConsumer.
func newFileConsumer(fc <-chan string, rc chan<- *goFile, ec chan error) *fileConsumer {
	return &fileConsumer{
		fpCh:  fc,
		resCh: rc,
		errCh: ec,
	}
}

// consume consumes file paths from the given channel,
// processes them, and then sends back either the processed
// result or an error (if one occurred).
//
// It will stop processing if an error occurs or if the
// context is cancelled, since this is an all or nothing
// command (can't *half* converge files).
func (fc *fileConsumer) consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-fc.errCh:
			return
		case fp, ok := <-fc.fpCh:
			if !ok {
				return
			}
			res, err := fc.processFile(fp)
			if err != nil {
				fc.errCh <- err
				return
			}
			fc.resCh <- res
		}
	}
}

// processFile processes the given file path and returns the
// processed result or an error if one occurred.
func (fc *fileConsumer) processFile(fp string) (*goFile, error) {
	proc := newFileProcessor(fp)
	if proc == nil {
		return nil, fmt.Errorf("failed to create fileProcessor for file: %s", fp)
	}

	res, err := proc.process()
	if err != nil {
		return nil, fmt.Errorf("error processing file: %w", err)
	}

	return res, nil
}
