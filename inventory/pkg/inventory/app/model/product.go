package model

import (
	"github.com/google/uuid"
)

type Product struct {
	ID       uuid.UUID
	Name     string
	Price    float64
	Quantity int
}
