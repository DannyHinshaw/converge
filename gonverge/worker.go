package gonverge

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// fileProducer walks a directory and sends all file paths
// to the given channel for the consumer to process.
type fileProducer struct {
	// excludes is a list of files that should
	// be excluded from processing in the merge.
	excludes []string

	// fpCh is the channel to send file paths to.
	fpCh chan<- string

	// errCh is the channel to send errors to.
	errCh chan<- error
}

// newFileProducer returns a new fileProducer.
func newFileProducer(excludes []string, fpCh chan<- string, errCh chan<- error) *fileProducer {
	return &fileProducer{
		fpCh:     fpCh,
		errCh:    errCh,
		excludes: excludes,
	}
}

// produce walks the given directory and sends all file paths
// to the fpCh channel for the consumer to process.
func (fp *fileProducer) produce(dir string) {
	defer close(fp.fpCh)
	err := fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if !validFile(info, fp.excludes) {
			return nil
		}

		fullPath := filepath.Join(dir, path)
		fp.fpCh <- fullPath

		return nil
	})
	if err != nil {
		fp.errCh <- err
	}
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
func newFileConsumer(fpCh <-chan string, resCh chan<- *goFile, errCh chan error) *fileConsumer {
	return &fileConsumer{
		fpCh:  fpCh,
		resCh: resCh,
		errCh: errCh,
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
			proc := newFileProcessor(fp)
			res, err := proc.process()
			if err != nil {
				fc.errCh <- err
				return
			}
			fc.resCh <- res
		}
	}
}

// validFile checks that the file is a valid *non-test* Go file.
func validFile(fileInfo os.FileInfo, excludes []string) bool {
	name := fileInfo.Name()
	if !strings.HasSuffix(name, ".go") {
		return false
	}
	if strings.HasSuffix(name, "_test.go") {
		return false
	}

	// Iterate over excludes and check if the file
	// should be excluded from processing.
	for _, exclude := range excludes {
		if name == exclude {
			return false
		}
	}

	return true
}
