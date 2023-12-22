package gonverge

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dannyhinshaw/converge/internal/logger"
)

// fileProducer walks a directory and sends all file paths
// to the given channel for the consumer to process.
type fileProducer struct {
	// excludes is a map of files to exclude from processing.

	excludes map[string]regexp.Regexp

	// packages is a set of packages to include
	// in the output. If empty, converger will
	// default to top-level NON-TEST package in
	// the given directory.
	packages map[string]struct{}

	// log is the logger to use for logging.
	log logger.LevelLogger

	// fpCh is the channel to send file paths to.
	fpCh chan<- string

	// errCh is the channel to send errors to.
	errCh chan<- error
}

// newFileProducer returns a new fileProducer.
func newFileProducer(ll logger.LevelLogger, excludes []string,
	packages []string, fpCh chan<- string, errCh chan<- error) *fileProducer {
	ll = ll.WithGroup("file_producer")

	packageSet := make(map[string]struct{}, len(packages))
	for _, pkg := range packages {
		packageSet[pkg] = struct{}{}
	}

	excludeSet := make(map[string]regexp.Regexp)
	for _, exclude := range excludes {
		if exclude == "" {
			continue
		}
		if _, ok := excludeSet[exclude]; ok {
			continue
		}

		re, err := regexp.Compile(exclude)
		if err != nil {
			errCh <- fmt.Errorf("error compiling exclude regex: %w", err)
			continue
		}
		if re == nil {
			errCh <- fmt.Errorf("error compiling exclude regex: regex is nil")
			continue
		}
		excludeSet[exclude] = *re
	}

	return &fileProducer{
		log:      ll,
		fpCh:     fpCh,
		errCh:    errCh,
		excludes: excludeSet,
		packages: packageSet,
	}
}

// produce walks the given directory and sends all file paths
// to the fpCh channel for the consumer to process.
func (fp *fileProducer) produce(dir string) {
	defer close(fp.fpCh)

	err := fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error {
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
			fp.log.Info("file is not valid", "fullPath", fullPath)
			return nil
		}

		fp.log.Info("file is valid", "fullPath", fullPath)
		fp.fpCh <- fullPath

		return nil
	})
	if err != nil {
		fp.errCh <- err
	}
}

// validFile checks that the file is a valid *non-test* Go file.
func (fp *fileProducer) validFile(name, fullPath string) bool {
	checkPkgs := len(fp.packages) > 0

	// Check if the file should be excluded from processing.
	for _, re := range fp.excludes {
		if re.MatchString(name) {
			fp.log.Info("file %s excluded from processing", name)
			return false
		}
	}

	// Packages specified overrides all other checks.
	if checkPkgs {
		if fp.validPackage(fullPath) {
			fp.log.Info("file is valid package", "name", name)
			return true
		}
		return false
	}

	// Only process Go files
	if !strings.HasSuffix(name, ".go") {
		return false
	}

	// Ignore test files by default.
	// NOTE: If you need to converge test files, you can
	// do so by specifying the package name in the packages
	// option. This will override the default behavior.
	if strings.HasSuffix(name, "_test.go") {
		return false
	}

	return true
}

// validPackage checks that the file is a valid Go file and that
// the package name is in the set of packages to include.
func (fp *fileProducer) validPackage(fullPath string) bool {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fullPath, nil, parser.PackageClauseOnly)
	if err != nil {
		fp.log.Info("error parsing file %s: %v", fullPath, err)
		return false
	}

	if node == nil || node.Name == nil {
		fp.log.Info("error parsing file %s: %v", fullPath, err)
		return false
	}

	name := node.Name.Name
	if _, ok := fp.packages[name]; !ok {
		fp.log.Info("package not in packages set", "name", name, "set", fp.packages)
		return false
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
				fc.errCh <- fmt.Errorf("error processing file: %w", err)
				return
			}
			fc.resCh <- res
		}
	}
}
