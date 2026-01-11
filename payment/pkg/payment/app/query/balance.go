package query

import (
	"context"

	"github.com/google/uuid"

	appmodel "payment/pkg/payment/app/model"
)

type AccountBalanceQueryService interface {
	FindBalance(ctx context.Context, id uuid.UUID) (*appmodel.CustomerBalance, error)
}
