package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/infrastructure/observability"
	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

// Recovery converts a panic into the standard error envelope. The stack trace
// is logged and never sent to the client (§156).
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}
			if recovered == http.ErrAbortHandler {
				panic(recovered)
			}

			observability.LoggerFrom(r.Context()).Error("recovered from panic in http handler",
				"panic", recovered,
				"stack", string(debug.Stack()),
			)
			response.Error(w, r, apperr.New(apperr.CodeInternal))
		}()

		next.ServeHTTP(w, r)
	})
}
