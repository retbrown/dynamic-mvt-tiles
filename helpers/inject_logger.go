package helpers

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type chilogger struct {
	logZ *zap.Logger
}

// NewZapMiddleware returns a new Zap Middleware handler.
func NewZapMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return chilogger{
		logZ: logger,
	}.middleware
}

func (c chilogger) middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/health-check" || r.RequestURI == "/metrics" {
			next.ServeHTTP(w, r.WithContext(r.Context()))
		} else {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			logger := c.logZ

			next.ServeHTTP(ww, r.Clone(ctxzap.ToContext(r.Context(), logger)))
		}
	}
	return http.HandlerFunc(fn)
}
