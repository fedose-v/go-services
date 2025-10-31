package model

import (
	"time"
	
	"github.com/google/uuid"
)

type ProductCreated struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

func (e ProductCreated) Type() string {
	return "ProductCreated"
}

type ProductDeleted struct {
	ID uuid.UUID
}

func (e ProductDeleted) Type() string {
	return "ProductDeleted"
}

type ProductQuantityChanged struct {
	ID           uuid.UUID
	NewQuantity  int
	PrevQuantity int
}

func (e ProductQuantityChanged) Type() string {
	return "ProductQuantityChanged"
}
