package query

import (
	"context"

	"github.com/google/uuid"

	"order/pkg/app/data"
)

type OrderQueryService interface {
	FindUser(ctx context.Context, orderID uuid.UUID) (*data.Order, error)
}
