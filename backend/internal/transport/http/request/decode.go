// Package request decodes and validates inbound HTTP data at the API boundary.
// Frontend validation is never trusted (§107).
package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
)

// DecodeJSON reads exactly one JSON object into dst. Unknown fields are
// rejected so that contract drift surfaces immediately instead of being
// silently ignored.
func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return apperr.New(apperr.CodeValidation).WithMessage("A request body is required.")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return decodeError(err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return apperr.New(apperr.CodeValidation).WithMessage("The request body must contain a single JSON object.")
	}
	return nil
}

func decodeError(err error) error {
	var maxBytes *http.MaxBytesError
	if errors.As(err, &maxBytes) {
		return apperr.New(apperr.CodeValidation).WithMessage("The request body is too large.")
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return apperr.New(apperr.CodeValidation).
			WithMessage("The request body contains a field of the wrong type.").
			WithDetail("fields", map[string]string{typeErr.Field: "expected " + typeErr.Type.String()})
	}

	if errors.Is(err, io.EOF) {
		return apperr.New(apperr.CodeValidation).WithMessage("A request body is required.")
	}
	return apperr.New(apperr.CodeValidation).WithMessage("The request body is not valid JSON.").WithCause(err)
}
