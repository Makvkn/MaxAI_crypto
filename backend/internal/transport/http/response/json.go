// Package response renders HTTP responses. Every body is snake_case JSON, and
// every failure uses the single documented error envelope (§105).
package response

import (
	"encoding/json"
	"net/http"

	"github.com/maxaicrypto/backend/internal/infrastructure/observability"
)

// JSON writes v with the given status code.
func JSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		observability.LoggerFrom(r.Context()).Error("failed to marshal response body", "error", err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(internalErrorBody)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		observability.LoggerFrom(r.Context()).Warn("failed to write response body", "error", err)
	}
}

// OK writes a 200 response.
func OK(w http.ResponseWriter, r *http.Request, v any) {
	JSON(w, r, http.StatusOK, v)
}

// Created writes a 201 response.
func Created(w http.ResponseWriter, r *http.Request, v any) {
	JSON(w, r, http.StatusCreated, v)
}

// NoContent writes a 204 response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
