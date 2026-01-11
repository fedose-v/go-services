package service

import (
	"github.com/google/uuid"

	"inventory/pkg/app/data"
)

type ProductQueryService interface {
	ListProducts() ([]data.ProductData, error)
	FindProduct(id uuid.UUID) (*data.ProductData, error)
}
