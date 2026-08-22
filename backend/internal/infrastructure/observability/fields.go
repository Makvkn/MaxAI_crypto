package observability

import (
	"log/slog"
	"strings"
)

// Canonical structured log field names (§123). Handlers, services, adapters and
// jobs must reuse these constants so that logs stay queryable across layers.
const (
	FieldRequestID      = "request_id"
	FieldUserID         = "user_id"
	FieldWalletID       = "wallet_id"
	FieldConversationID = "conversation_id"
	FieldMessageID      = "message_id"
	FieldJobID          = "job_id"
	FieldSyncJobID      = "sync_job_id"
	FieldProvider       = "provider"
	FieldOperation      = "operation"
	FieldDurationMS     = "duration_ms"
	FieldStatus         = "status"
	FieldErrorCode      = "error_code"
	FieldError          = "error"
	FieldChainID        = "chain_id"
	FieldComponent      = "component"
)

// RedactedAddress renders a blockchain address in a privacy-conscious form.
// The specification allows logging addresses only when necessary and prefers a
// reduced representation (§123), so callers never log the full string.
func RedactedAddress(address string) string {
	address = strings.TrimSpace(address)
	const keep = 4
	if len(address) <= keep*2+1 {
		return "***"
	}
	return address[:keep] + "..." + address[len(address)-keep:]
}

// Address builds a log attribute carrying a redacted blockchain address.
func Address(address string) slog.Attr {
	return slog.String("address", RedactedAddress(address))
}
