package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrOrderNotFound = errors.New("order not found")
var ErrOrderAccessDenied = errors.New("order access denied")
var ErrInvalidOrderStatus = errors.New("invalid order status")

type OrderStatus int

const (
	Open OrderStatus = iota
	Pending
	Paid
	Cancelled
	Deleted
)

type Order struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Status     OrderStatus
	Items      []Item
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

type Item struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	Price     float64
}

type FindSpec struct {
	OrderID        *uuid.UUID
	CustomerID     *uuid.UUID
	Status         *OrderStatus
	IncludeDeleted bool
}

type OrderRepository interface {
	NextID() (uuid.UUID, error)
	Store(order *Order) error
	Find(spec FindSpec) (*Order, error)
	Delete(id uuid.UUID) error
}
