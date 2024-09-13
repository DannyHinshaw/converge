package gonverge

import (
	"fmt"
	"go/format"
	"strings"
)

// goFile represents the contents of a Go source file,
// including its package name, imports, and code.
//
// This struct is used to aggregate multiple Go files into
// a single file, maintaining proper syntax and formatting.
type goFile struct {
	// pkgName is the name of the package
	// that the file belongs to.
	pkgName string

	// imports is a set of all imports for the file.
	imports map[string]struct{}

	// code is the literal code for the file.
	code strings.Builder
}

// newGoFile returns a new goFile instance.
func newGoFile() *goFile {
	return &goFile{
		imports: make(map[string]struct{}),
	}
}

// addImport adds the given import to the set of imports,
// ensuring no duplicate imports are added. This is important
// when merging multiple Go files that may have overlapping dependencies.
func (f *goFile) addImport(importLine string) {
	f.imports[importLine] = struct{}{}
}

// appendCode adds a line of Go code to the current file. Each
// line is appended with a newline character to maintain proper syntax.
func (f *goFile) appendCode(code string) {
	f.code.WriteString(code)
	f.code.WriteString("\n")
}

// merge merges the given goFile into the result
// by adding the imports and appending the code.
func (f *goFile) merge(gf *goFile) {
	if f.pkgName == "" {
		f.pkgName = gf.pkgName
	}

	for imp := range gf.imports {
		f.addImport(imp)
	}

	f.code.WriteString(gf.code.String())
}

// buildImports returns a string of
// all imports for the given package.
func (f *goFile) buildImports() string {
	if len(f.imports) == 0 {
		return ""
	}

	var builder strings.Builder
	if len(f.imports) == 1 {
		for imp := range f.imports {
			builder.WriteString("import ")
			builder.WriteString(imp)
			builder.WriteString("\n")
		}
	} else {
		builder.WriteString("import (\n")
		for imp := range f.imports {
			builder.WriteString("\t")
			builder.WriteString(imp)
			builder.WriteString("\n")
		}
		builder.WriteString(")\n")
	}

	return builder.String()
}

// FormatCode formats the code in the goFile and returns the result.
func (f *goFile) FormatCode() ([]byte, error) {
	// Use a strings.Builder to build
	// the newly converged Go file.
	var builder strings.Builder

	// Write the package name.
	builder.WriteString("package ")
	builder.WriteString(f.pkgName)
	builder.WriteString("\n\n")

	// Write the imports.
	if len(f.imports) > 0 {
		imports := f.buildImports()
		builder.WriteString(imports)
	}

	// Write the code.
	builder.WriteString(f.code.String())

	// Use go/format to format the code in standard gofmt style.
	// Note(@danny): We should also allow the user to specify
	// using gofumpt or other formatters.
	b, err := format.Source([]byte(builder.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to format code: %w", err)
	}

	return b, nil
}
