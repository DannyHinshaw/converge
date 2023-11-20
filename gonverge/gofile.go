package gonverge

import (
	"go/format"
	"strings"
)

// goFile is structure that holds the contents of a Go file,
// and can be used to generate a new Go file.
type goFile struct {
	// pkgName is the name of the package that the file belongs to.
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

// addImport adds the given import to the set of imports.
func (f *goFile) addImport(importLine string) {
	f.imports[importLine] = struct{}{}
}

// appendCode appends the literal code to the code block.
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
	var builder strings.Builder
	builder.WriteString("package ")
	builder.WriteString(f.pkgName)
	builder.WriteString("\n\n")
	if len(f.imports) > 0 {
		builder.WriteString(f.buildImports())
	}
	builder.WriteString(f.code.String())

	// Use go/format to format the code in standard gofmt style.
	return format.Source([]byte(builder.String()))
}
