package cmd_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/dannyhinshaw/converge/cmd"
	"github.com/dannyhinshaw/converge/internal/gonverge"
)

func TestConverge_Run(t *testing.T) {
	tests := map[string]struct {
		setup func() (*cmd.Converge, func())
		err   bool
	}{
		"BasicConvergence": {
			setup: func() (*cmd.Converge, func()) {
				srcDir, cleanupSrc := createTempDirWithFiles(t, map[string]string{
					"file1.go": "package main\nfunc main() {}",
					// Add more files as needed
				})

				fc := gonverge.NewGoFileConverger()
				opt := cmd.WithWriter(bytes.NewBuffer(nil))
				cmdRunner := cmd.NewCommand(fc, srcDir, opt)

				return cmdRunner, cleanupSrc
			},
			err: false,
		},
		"InvalidSourceDirectory": {
			setup: func() (*cmd.Converge, func()) {
				fc := gonverge.NewGoFileConverger()
				opt := cmd.WithWriter(bytes.NewBuffer(nil))
				cmdRunner := cmd.NewCommand(fc, "/invalid/dir", opt)

				return cmdRunner, func() {}
			},
			err: true,
		},
		"OutputToFile": {
			setup: func() (*cmd.Converge, func()) {
				srcDir, cleanupSrc := createTempDirWithFiles(t, map[string]string{
					"file1.go": "package main\nfunc main() {}",
				})
				outFile, cleanupOut := createTempFile(t)

				opt := cmd.WithDstFile(outFile.Name())
				fc := gonverge.NewGoFileConverger()
				cmdRunner := cmd.NewCommand(fc, srcDir, opt)

				return cmdRunner, func() {
					cleanupSrc()
					cleanupOut()
				}
			},
			err: false,
		},
		"ExclusionListEffectiveness": {
			setup: func() (*cmd.Converge, func()) {
				srcDir, cleanupSrc := createTempDirWithFiles(t, map[string]string{
					"file1.go":   "package main\nfunc main() {}",
					"exclude.go": "// This file should be excluded",
				})

				opt := gonverge.WithExcludes([]string{
					"exclude.go",
				})

				fc := gonverge.NewGoFileConverger(opt)
				cmdRunner := cmd.NewCommand(fc, srcDir)

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
				t.Errorf("Converge.Run() error = %v, wantErr %v", err, tc.err)
			}
		})
	}
}

func TestConverge_ContextCancellation(t *testing.T) {
	srcDir, cleanupSrc := createTempDirWithFiles(t, map[string]string{
		"file1.go": "package main\nfunc main() {}",
	})
	defer cleanupSrc()

	fc := gonverge.NewGoFileConverger()
	opt := cmd.WithWriter(bytes.NewBuffer(nil))
	cmdRunner := cmd.NewCommand(fc, srcDir, opt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	// Run the command in a separate goroutine.
	go func() {
		defer wg.Done()
		_ = cmdRunner.Run(ctx)
	}()

	// Cancel the context after a brief delay to simulate interruption.
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for the command to complete.
	wg.Wait()

	// Try to run the command again after cancellation.
	// We expect an error due to context cancellation.
	if err := cmdRunner.Run(ctx); err == nil {
		t.Errorf("Expected error due to context cancellation, but got nil")
	}
}

// createTempFile creates a single temp file, returning the file pointer and a cleanup function.
func createTempFile(t *testing.T) (*os.File, func()) {
	t.Helper()
	file, err := os.CreateTemp("", "converge_test_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return file, func() { os.Remove(file.Name()) }
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

	return dir, func() { os.RemoveAll(dir) }
}
