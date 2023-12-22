package gonverge

import "github.com/dannyhinshaw/converge/internal/logger"

// Option is a functional option for the GoFileConverger.
type Option func(*GoFileConverger)

// WithLogger sets the logger for the GoFileConverger.
func WithLogger(logger logger.LevelLogger) Option {
	return func(gfc *GoFileConverger) {
		gfc.Logger = logger
	}
}

// WithPackages sets the list of packages to include in the output.
func WithPackages(packages []string) Option {
	return func(gfc *GoFileConverger) {
		gfc.Packages = packages
	}
}

// WithExcludes sets the list of files to exclude from merging.
func WithExcludes(excludes []string) Option {
	return func(gfc *GoFileConverger) {
		gfc.Excludes = excludes
	}
}

// WithMaxWorkers sets the maximum amount of workers to use.
func WithMaxWorkers(maxWorkers int) Option {
	return func(gfc *GoFileConverger) {
		gfc.MaxWorkers = maxWorkers
	}
}
