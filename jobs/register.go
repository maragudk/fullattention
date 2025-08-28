package jobs

import (
	"log/slog"

	"maragu.dev/glue/jobs"
)

type RegisterOpts struct {
	Log *slog.Logger
}

// Register all available jobs with the given dependencies.
func Register(r *jobs.Runner, opts RegisterOpts) {
	if opts.Log == nil {
		opts.Log = slog.New(slog.DiscardHandler)
	}
}
