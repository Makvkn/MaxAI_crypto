package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/maxaicrypto/backend/internal/infrastructure/observability"
)

// statusRecorder captures the status code so the access log can report it
// without the handler cooperating.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach the underlying writer, which SSE
// streaming needs for flushing (§82).
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// Logging emits one structured access log per request and seeds the context
// logger with the request identifier, so every downstream log line correlates
// (§123).
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestLogger := logger.With(observability.FieldRequestID, RequestIDFrom(r.Context()))
			ctx := observability.WithLogger(r.Context(), requestLogger)

			recorder := &statusRecorder{ResponseWriter: w}
			started := time.Now()

			next.ServeHTTP(recorder, r.WithContext(ctx))

			if recorder.status == 0 {
				recorder.status = http.StatusOK
			}
			requestLogger.Info("http request",
				observability.FieldOperation, r.Method+" "+r.URL.Path,
				observability.FieldStatus, recorder.status,
				observability.FieldDurationMS, time.Since(started).Milliseconds(),
				"bytes", recorder.bytes,
			)
		})
	}
}
