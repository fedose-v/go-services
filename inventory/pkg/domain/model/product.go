package model

import (
	"errors"

	"github.com/gofrs/uuid"
)

type Product struct {
	ID       uuid.UUID
	Name     string
	Price    float64
	Quantity int
}

var ErrProductNotFound = errors.New("product not found")

type ProductRepository interface {
	NextID() uuid.UUID
	Store(product *Product) error
	Find(id uuid.UUID) (*Product, error)
	List() (*[]Product, error)
	Delete(id uuid.UUID) error
}
