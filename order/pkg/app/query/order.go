package query

import (
	"context"

	"github.com/google/uuid"

	appmodel "order/pkg/app/model"
)

type OrderQueryService interface {
	FindOrder(ctx context.Context, orderID uuid.UUID) (*appmodel.Order, error)
}
