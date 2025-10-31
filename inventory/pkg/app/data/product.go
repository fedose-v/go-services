package data

import (
	"github.com/google/uuid"
	"time"
)

type ProductData struct {
	ID        uuid.UUID
	Name      string
	Price     float64
	Quantity  int
	CreatedAt time.Time
}
