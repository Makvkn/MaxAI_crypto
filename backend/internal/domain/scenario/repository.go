package scenario

import (
	"context"

	"github.com/google/uuid"
)

// Repository persists deterministic scenario calculations so AI claims can cite
// the exact result that was explained (§51, §73, §85).
type Repository interface {
	Create(ctx context.Context, result Result) (Result, error)
	GetByID(ctx context.Context, id uuid.UUID) (Result, error)
}
