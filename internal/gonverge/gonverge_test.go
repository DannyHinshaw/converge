package gonverge_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/dannyhinshaw/converge/internal/gonverge"
)

func TestGoFileConverger_ConvergeFiles(t *testing.T) {
	// Define test cases
	tests := map[string]struct {
		setup    func() (string, func())
		expected string
		excludes []string
		err      bool
	}{
		"EmptyDirectory": {
			setup: func() (string, func()) {
				dir := createTempDirWithFiles(t, map[string]string{})
				return dir, func() { os.RemoveAll(dir) }
			},
			expected: "",
			err:      false,
		},
		"SingleFile": {
			setup: func() (string, func()) {
				files := map[string]string{"file.go": "package main\nfunc main() {}"}
				dir := createTempDirWithFiles(t, files)
				return dir, func() { os.RemoveAll(dir) }
			},
			expected: "package main\n\nfunc main() {}\n",
			err:      false,
		},
		"MultipleFiles": {
			setup: func() (string, func()) {
				files := map[string]string{
					"file1.go": "package main\nfunc func1() {}",
					"file2.go": "package main\nfunc func2() {}",
				}
				dir := createTempDirWithFiles(t, files)
				return dir, func() { os.RemoveAll(dir) }
			},
			expected: "package main\n\nfunc func1() {}\nfunc func2() {}\n",
			err:      false,
		},
		"MultipleFilesWithExclusion": {
			setup: func() (string, func()) {
				files := map[string]string{
					"file1.go":   "package main\nfunc func1() {}",
					"file2.go":   "package main\nfunc func2() {}",
					"exclude.go": "package main\nfunc exclude() {}",
				}
				dir := createTempDirWithFiles(t, files)
				return dir, func() { os.RemoveAll(dir) }
			},
			expected: "package main\n\nfunc func1() {}\nfunc func2() {}\n",
			excludes: []string{"exclude.go"},
			err:      false,
		},
		"MultipleFilesWithExclusionButNoFile": {
			setup: func() (string, func()) {
				files := map[string]string{
					"file1.go": "package main\nfunc func1() {}",
					"file2.go": "package main\nfunc func2() {}",
				}
				dir := createTempDirWithFiles(t, files)
				return dir, func() { os.RemoveAll(dir) }
			},
			expected: "package main\n\nfunc func1() {}\nfunc func2() {}\n",
			excludes: []string{"exclude.go"},
			err:      false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir, cleanup := tc.setup()
			defer cleanup()

			opts := []gonverge.Option{
				gonverge.WithMaxWorkers(1),
			}

			if len(tc.excludes) > 0 {
				opts = append(opts, gonverge.WithExcludes(tc.excludes))
			}

			converger := gonverge.NewGoFileConverger(opts...)

			var output bytes.Buffer
			err := converger.ConvergeFiles(context.Background(), dir, &output)
			if tc.err && err == nil {
				t.Fatalf("ConvergeFiles() error = %v, wantErr %v", err, tc.err)
			}

			if got := output.String(); got != tc.expected {
				t.Errorf("ConvergeFiles() got = %v, want %v", got, tc.expected)
			}
		})
	}
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
