package query

import (
	"context"

	"github.com/google/uuid"

	"inventory/pkg/inventory/app/model"
)

type ProductQueryService interface {
	ListProducts(ctx context.Context) ([]model.Product, error)
	FindProduct(ctx context.Context, id uuid.UUID) (*model.Product, error)
}
