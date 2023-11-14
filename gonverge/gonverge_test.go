package gonverge_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	gonverge2 "github.com/dannyhinshaw/converge/gonverge"
)

func TestGoFileConverger_ConvergeFiles(t *testing.T) {
	// Define test cases
	tests := map[string]struct {
		setup    func() (string, func())
		expected string
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
		// Additional test cases here...
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir, cleanup := tc.setup()
			defer cleanup()

			opts := []gonverge2.Option{
				gonverge2.WithMaxWorkers(1),
			}

			converger := gonverge2.NewGoFileConverger(opts...)

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
