package model

import (
	"github.com/google/uuid"
)

type CustomerBalance struct {
	CustomerID uuid.UUID
	Amount     float64
}
