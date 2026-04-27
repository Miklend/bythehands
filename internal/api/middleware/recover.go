package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"talkabout/internal/api/handlers"
)

func Recoverer(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered", slog.Any("panic", rec), slog.String("path", r.URL.Path), slog.String("stack", string(debug.Stack())))
					handlers.WriteError(w, http.StatusInternalServerError, "internal", "internal error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
