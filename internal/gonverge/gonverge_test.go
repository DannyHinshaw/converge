package gonverge_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dannyhinshaw/converge/internal/gonverge"
)

func TestGoFileConverger_ConvergeFiles(t *testing.T) {
	a := assert.New(t)

	excludeRe := regexp.MustCompile("exclude.go")

	tests := map[string]struct {
		files    map[string]string
		excludes []regexp.Regexp
		expected string
		err      bool
	}{
		"EmptyDirectory": {
			files:    nil,
			expected: "",
		},
		"SingleFile": {
			files: map[string]string{
				"file.go": "package main\nfunc main() {}",
			},
			expected: "package main\n\nfunc main() {}\n",
		},
		"MultipleFiles": {
			files: map[string]string{
				"file1.go": "package main\nfunc func1() {}",
				"file2.go": "package main\nfunc func2() {}",
			},
			expected: "package main\n\nfunc func1() {}\nfunc func2() {}\n",
		},
		"MultipleFilesWithExclusion": {
			files: map[string]string{
				"file1.go":   "package main\nfunc func1() {}",
				"file2.go":   "package main\nfunc func2() {}",
				"exclude.go": "package main\nfunc exclude() {}",
			},
			expected: "package main\n\nfunc func1() {}\nfunc func2() {}\n",
			excludes: []regexp.Regexp{*excludeRe},
		},
		"MultipleFilesWithExclusionButNoFile": {
			files: map[string]string{
				"file1.go": "package main\nfunc func1() {}",
				"file2.go": "package main\nfunc func2() {}",
			},
			expected: "package main\n\nfunc func1() {}\nfunc func2() {}\n",
			excludes: []regexp.Regexp{*excludeRe},
		},
		"MultipleFilesWithNonGoFiles": {
			files: map[string]string{
				"file1.go": "package main\nfunc func1() {}",
				"file.txt": "This is a text file",
				"file2.go": "package main\nfunc func2() {}",
			},
			expected: "package main\n\nfunc func1() {}\nfunc func2() {}\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := createTempDirWithFiles(t, tc.files)
			defer func() {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatalf("Failed to remove temp dir: %v", err)
				}
			}()

			// Use only one worker to make the test deterministic.
			opts := []gonverge.Option{
				gonverge.WithMaxWorkers(1),
			}
			if len(tc.excludes) > 0 {
				opts = append(opts, gonverge.WithExcludes(tc.excludes))
			}
			converger := gonverge.NewGoFileConverger(
				opts...,
			)

			var output bytes.Buffer

			err := converger.ConvergeFiles(context.Background(), dir, &output)
			a.False(tc.err && err == nil)
			a.Equal(tc.expected, output.String())
		})
	}
}

func TestGoFileConverger_DirectoryNotFound(t *testing.T) {
	a := assert.New(t)

	dir := "/non-existent-directory"
	converger := gonverge.NewGoFileConverger()

	var output bytes.Buffer
	err := converger.ConvergeFiles(context.Background(), dir, &output)
	a.Error(err)
}

// createTempDirWithFiles creates a temporary directory with the given files for testing.
func createTempDirWithFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "gonverge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	for filename, content := range files {
		fp := filepath.Join(dir, filename)
		if err = os.WriteFile(fp, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}

	return dir
}
