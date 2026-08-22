package response

import (
	"net/http"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/infrastructure/observability"
)

// errorEnvelope is the only error shape the API emits (§105).
type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

// internalErrorBody is a pre-rendered envelope used when marshalling itself
// fails, so the client always receives a parseable body.
var internalErrorBody = []byte(`{"error":{"code":"INTERNAL_ERROR","message":"An unexpected error occurred.","details":{}}}`)

// Error renders err as the standard envelope. Internal causes are logged, not
// returned, so provider and database details never reach the frontend
// (§28, §156).
func Error(w http.ResponseWriter, r *http.Request, err error) {
	appErr := apperr.From(err)
	if appErr == nil {
		appErr = apperr.New(apperr.CodeInternal)
	}

	logger := observability.LoggerFrom(r.Context()).With(
		observability.FieldErrorCode, string(appErr.Code),
		observability.FieldOperation, r.Method+" "+r.URL.Path,
	)
	if appErr.Category() == apperr.CategoryInternal {
		logger.Error("request failed", "error", appErr.Error())
	} else {
		logger.Warn("request failed", "error", appErr.Error())
	}

	details := appErr.Details
	if details == nil {
		details = map[string]any{}
	}

	JSON(w, r, appErr.HTTPStatus(), errorEnvelope{
		Error: errorBody{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: details,
		},
	})
}
