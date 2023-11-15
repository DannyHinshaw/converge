package gonverge

// Option is a functional option for the GoFileConverger.
type Option func(*GoFileConverger)

// WithMaxWorkers sets the maximum amount of workers to use.
func WithMaxWorkers(maxWorkers int) Option {
	return func(gfc *GoFileConverger) {
		gfc.MaxWorkers = maxWorkers
	}
}

// WithExcludes sets the list of files to exclude from merging.
func WithExcludes(excludes []string) Option {
	return func(gfc *GoFileConverger) {
		gfc.Excludes = excludes
	}
}
