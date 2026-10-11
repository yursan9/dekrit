package output

import (
	"io"
	"os"
)

// Options configures a Progress or Spinner.
type Options struct {
	Writer io.Writer
	Prefix string
}

// Option customizes Options.
type Option func(*Options)

// WithWriter sets the writer the widget renders to. A nil writer is ignored.
func WithWriter(w io.Writer) Option {
	return func(o *Options) {
		if w != nil {
			o.Writer = w
		}
	}
}

// WithPrefix sets text printed before the widget on each line.
func WithPrefix(prefix string) Option {
	return func(o *Options) { o.Prefix = prefix }
}

func resolveOptions(opts []Option) Options {
	o := Options{Writer: os.Stdout}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
