package gonverge

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// procState is the state of the file processor.
type procState int

const (
	// procStateCoding represents the file processor state
	// when processing a code block.
	procStateCoding procState = iota

	// procStateImporting represents the file processor state
	// when processing an import block.
	procStateImporting
)

const (
	// tokenPkgDecl is the token for a package declaration line.
	tokenPkgDecl = `package `

	// tokenImport is the token for import declarations.
	tokenImport = `import`

	// tokenImportMono is the token for a single import line.
	tokenImportMono = tokenImport + ` "`

	// tokenImportMulti is the token that starts an import block.
	tokenImportMultiStart = tokenImport + ` (`

	// tokenImportMultiEnd is the token that ends an import block.
	tokenImportMultiFinish = `)`
)

// fileProcessor holds the *os.File representations
// of the command line arguments.
type fileProcessor struct {
	// filePath is the path to the file to process.
	filePath string

	// state determines how the current
	// line should be processed.
	state procState
}

// newFileProcessor returns a new fileProcessor.
func newFileProcessor(filePath string) *fileProcessor {
	return &fileProcessor{
		filePath: filePath,
		state:    procStateCoding,
	}
}

// process handles opening, parsing, and aggregating
// the contents of the file into a goFile.
func (p *fileProcessor) process() (*goFile, error) {
	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	res := newGoFile()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		switch line := scanner.Text(); {
		case strings.HasPrefix(line, tokenPkgDecl):
			res.pkgName = strings.TrimPrefix(line, tokenPkgDecl)
			p.state = procStateCoding

		case strings.HasPrefix(line, tokenImportMultiStart):
			p.state = procStateImporting

		case strings.HasPrefix(line, tokenImportMono):
			res.addImport(strings.TrimPrefix(line, tokenImport))

		case p.importing() && strings.HasSuffix(line, tokenImportMultiFinish):
			p.state = procStateCoding

		case p.importing():
			if line == "" {
				continue
			}
			res.addImport(line)

		case p.coding():
			res.appendCode(line)

		default:
			res.appendCode(line)
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return res, nil
}

// importing returns true if the filePath processor is currently
// processing an import block.
func (p *fileProcessor) importing() bool {
	return p.state == procStateImporting
}

// coding returns true if the filePath processor is currently
// processing a code block.
func (p *fileProcessor) coding() bool {
	return p.state == procStateCoding
}
