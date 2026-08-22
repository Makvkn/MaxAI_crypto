package apperr

import (
	"errors"
	"net/http"
	"testing"
)

func TestFromMapsUnknownErrorsToInternal(t *testing.T) {
	cause := errors.New("pq: connection refused on 10.0.0.4:5432")

	got := From(cause)

	if got.Code != CodeInternal {
		t.Errorf("code = %s, want %s", got.Code, CodeInternal)
	}
	if got.Message == cause.Error() {
		t.Error("user-facing message leaks the internal cause")
	}
	if !errors.Is(got, cause) {
		t.Error("internal cause was dropped and is unavailable to logs")
	}
}

func TestFromPreservesApplicationErrors(t *testing.T) {
	original := New(CodeWalletNotFound).WithDetail("wallet_id", "w_1")

	got := From(errors.Join(errors.New("context"), original))

	if got.Code != CodeWalletNotFound {
		t.Fatalf("code = %s, want %s", got.Code, CodeWalletNotFound)
	}
	if got.Details["wallet_id"] != "w_1" {
		t.Errorf("details = %v, want wallet_id preserved", got.Details)
	}
}

func TestWithDetailDoesNotMutateTheReceiver(t *testing.T) {
	base := New(CodeRateLimit)

	first := base.WithDetail("limit", 10)
	second := base.WithDetail("limit", 20)

	if len(base.Details) != 0 {
		t.Errorf("base error was mutated: %v", base.Details)
	}
	if first.Details["limit"] != 10 || second.Details["limit"] != 20 {
		t.Errorf("clones share state: first=%v second=%v", first.Details, second.Details)
	}
}

func TestEveryCodeMapsToAStatusAndCategory(t *testing.T) {
	for code, meta := range registry {
		if meta.category == "" {
			t.Errorf("%s has no category", code)
		}
		if meta.status < http.StatusOK || meta.status > 599 {
			t.Errorf("%s has invalid status %d", code, meta.status)
		}
		if meta.message == "" {
			t.Errorf("%s has no default message", code)
		}
	}
}
