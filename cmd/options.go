package cmd

import "io"

// Option is a function that configures a Converge.
type Option func(*Converge)

// WithWriter sets the writer to use for the output.
func WithWriter(w io.Writer) Option {
	return func(c *Converge) {
		c.writer = w
	}
}

// WithDstFile sets the destination file to use for the output.
func WithDstFile(dst string) Option {
	return func(c *Converge) {
		c.dst = dst
	}
}
