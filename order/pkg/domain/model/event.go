package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderCreated struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
}

func (e OrderCreated) Type() string {
	return "OrderCreated"
}

type OrderItemChanged struct {
	OrderID      uuid.UUID
	AddedItems   []uuid.UUID
	RemovedItems []uuid.UUID
}

func (e OrderItemChanged) Type() string {
	return "OrderItemChanged"
}

type OrderDeleted struct {
	OrderID   uuid.UUID
	DeletedAt time.Time
}

func (e OrderDeleted) Type() string {
	return "OrderDeleted"
}

type OrderStatusChanged struct {
	OrderID     uuid.UUID
	OrderStatus OrderStatus
}

func (e OrderStatusChanged) Type() string {
	return "OrderStatusChanged"
}
