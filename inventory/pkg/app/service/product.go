package service

import (
	"inventory/pkg/app/data"
)

type ProductQueryService interface {
	ListProducts() ([]data.ProductData, error)
}
