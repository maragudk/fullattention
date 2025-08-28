package http

import (
	"log/slog"

	"maragu.dev/glue/http"

	"app/sqlite"
)

func InjectHTTPRouter(log *slog.Logger, db *sqlite.Database) func(*Router) {
	return func(r *Router) {
		r.Group(func(r *http.Router) {
			Home(r, log)
		})
	}
}
