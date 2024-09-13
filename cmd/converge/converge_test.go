package converge_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dannyhinshaw/converge/cmd/converge"
	"github.com/dannyhinshaw/converge/internal/gonverge"
)

func TestConverge_Run(t *testing.T) {
	reExclude := regexp.MustCompile("exclude.go")

	tests := map[string]struct {
		setup func() (*converge.Command, func())
		err   bool
	}{
		"BasicConvergence": {
			setup: func() (*converge.Command, func()) {
				srcDir, cleanupSrc := createTempDirWithFiles(t, map[string]string{
					"file1.go": "package main\nfunc main() {}",
					// Add more files as needed
				})

				fc := gonverge.NewGoFileConverger()
				opt := converge.WithWriter(bytes.NewBuffer(nil))
				cmdRunner := converge.NewCommand(fc, srcDir, opt)

				return cmdRunner, cleanupSrc
			},
			err: false,
		},
		"InvalidSourceDirectory": {
			setup: func() (*converge.Command, func()) {
				fc := gonverge.NewGoFileConverger()
				opt := converge.WithWriter(bytes.NewBuffer(nil))
				cmdRunner := converge.NewCommand(fc, "/invalid/dir", opt)

				return cmdRunner, func() {}
			},
			err: true,
		},
		"OutputToFile": {
			setup: func() (*converge.Command, func()) {
				srcDir, cleanupSrc := createTempDirWithFiles(t, map[string]string{
					"file1.go": "package main\nfunc main() {}",
				})
				outFile, cleanupOut := createTempFile(t)

				opt := converge.WithDstFile(outFile.Name())
				fc := gonverge.NewGoFileConverger()
				cmdRunner := converge.NewCommand(fc, srcDir, opt)

				return cmdRunner, func() {
					cleanupSrc()
					cleanupOut()
				}
			},
			err: false,
		},
		"ExclusionListEffectiveness": {
			setup: func() (*converge.Command, func()) {
				srcDir, cleanupSrc := createTempDirWithFiles(t, map[string]string{
					"file1.go":   "package main\nfunc main() {}",
					"exclude.go": "// This file should be excluded",
				})

				opt := gonverge.WithExcludes([]regexp.Regexp{
					*reExclude,
				})

				fc := gonverge.NewGoFileConverger(opt)
				cmdRunner := converge.NewCommand(fc, srcDir)

				return cmdRunner, cleanupSrc
			},
			err: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			converger, cleanup := tc.setup()
			defer cleanup()

			err := converger.Run(context.Background())
			if tc.err && err == nil {
				t.Errorf("Command.Run() error = %v, wantErr %v", err, tc.err)
			}
		})
	}
}

func TestConverge_ContextCancellation(t *testing.T) {
	r := require.New(t)

	srcDir, cleanupSrc := createTempDirWithFiles(t, map[string]string{
		"file1.go": "package main\nfunc main() {}",
	})
	defer cleanupSrc()

	fc := gonverge.NewGoFileConverger()
	opt := converge.WithWriter(bytes.NewBuffer(nil))
	cmdRunner := converge.NewCommand(fc, srcDir, opt)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Try to run the command after cancellation.
	// We expect an error due to context cancellation.
	err := cmdRunner.Run(ctx)
	r.ErrorIs(err, context.Canceled)
}

// createTempFile creates a single temp file, returning the file pointer and a cleanup function.
func createTempFile(t *testing.T) (*os.File, func()) {
	t.Helper()
	file, err := os.CreateTemp("", "converge_test_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return file, func() { _ = os.Remove(file.Name()) }
}

// createTempDirWithFiles creates a temp directory with the given files.
func createTempDirWithFiles(t *testing.T, files map[string]string) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "converge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	for filename, content := range files {
		fp := filepath.Join(dir, filename)
		if err = os.WriteFile(fp, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}

	return dir, func() { _ = os.RemoveAll(dir) }
}
