package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dannyhinshaw/converge/cmd"
	"github.com/dannyhinshaw/converge/internal/cli"
	"github.com/dannyhinshaw/converge/internal/gonverge"
	"github.com/dannyhinshaw/converge/internal/logger"
)

// versions is set by the linker at build time.
// Defaults to "(dev)" if not set.
var version = "(dev)"

func main() {
	app := cli.NewApp()

	// Handle version flag
	if app.ShowVersion {
		fmt.Println("converge version:", version) //nolint:forbidigo // not debugging
		return
	}

	// Create logger and set verbose logging.
	clog := logger.New("converge", logger.WithVerbose(app.VerboseLog))
	if app.VerboseLog {
		clog.Debug("Verbose logging enabled")
	}

	srcDir, err := app.ParseSrcDir()
	if err != nil {
		clog.Error(err.Error())
		flag.Usage()
		os.Exit(1)
	}

	var (
		converger   = createConverger(clog, app.Workers, app.Packages, app.Exclude)
		command     = createCommand(converger, srcDir, app.OutFile)
		ctx, cancel = newCancelContext(context.Background(), app.Timeout)
	)
	defer cancel()

	if err = command.Run(ctx); err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if app.OutFile != "" {
		log.Printf("Files from '%s' have been successfully merged into '%s'.\n", srcDir, app.OutFile)
	}
}

// createCommand creates a new Converge command with the given options.
func createCommand(converger cmd.FileConverger, src, outFile string) *cmd.Converge {
	var cmdOpts []cmd.Option
	if outFile != "" {
		cmdOpts = append(cmdOpts, cmd.WithDstFile(outFile))
	}

	return cmd.NewCommand(converger, src, cmdOpts...)
}

// createConverger creates a new GoFileConverger with the given options.
func createConverger(ll logger.LevelLogger, workers int, packages, exclude string) *gonverge.GoFileConverger {
	var gonvOpts []gonverge.Option
	if ll != nil {
		gonvOpts = append(gonvOpts, gonverge.WithLogger(ll))
	}

	if workers > 0 {
		gonvOpts = append(gonvOpts, gonverge.WithMaxWorkers(workers))
	}

	if packages != "" {
		var pkgs []string
		for _, pkg := range strings.Split(packages, ",") {
			pkgs = append(pkgs, strings.TrimSpace(pkg))
		}

		gonvOpts = append(gonvOpts, gonverge.WithPackages(pkgs))
	}

	if exclude != "" {
		var excludeFiles []string
		for _, excludeFile := range strings.Split(exclude, ",") {
			excludeFiles = append(excludeFiles, strings.TrimSpace(excludeFile))
		}

		gonvOpts = append(gonvOpts, gonverge.WithExcludes(excludeFiles))
	}

	return gonverge.NewGoFileConverger(gonvOpts...)
}

// newCancelContext returns a new cancellable context
// with the given timeout if one is specified.
//
// If no timeout is specified, the context will not have a
// timeout, but a cancel function will still be returned.
func newCancelContext(ctx context.Context, timeout int) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	}

	return context.WithCancel(ctx)
}
