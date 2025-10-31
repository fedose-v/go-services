package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID        uuid.UUID
	Name      string
	Price     float64
	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

var (
	ErrProductQuantityLessThanZero = errors.New("product quantity must be greater than zero")
	ErrProductNotFound             = errors.New("product not found")
)

type ProductRepository interface {
	NextID() (uuid.UUID, error)
	Store(product *Product) error
	Find(id uuid.UUID) (*Product, error)
	List() (*[]Product, error)
	Delete(id uuid.UUID) error
}
