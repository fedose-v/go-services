package domain

import "github.com/google/uuid"

type ProductCreated struct {
	ID   uuid.UUID
	Name string
}

func (e ProductCreated) Type() string {
	return "ProductCreated"
}

type ProductQuantityChanged struct {
	ID           uuid.UUID
	PrevQuantity int
	NewQuantity  int
}

func (e ProductQuantityChanged) Type() string {
	return "ProductQuantityChanged"
}
