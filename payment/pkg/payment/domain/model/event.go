package model

import (
	"time"

	"github.com/google/uuid"
)

type TransactionCreated struct {
	TransactionID uuid.UUID
	OrderID       uuid.UUID
	CustomerID    uuid.UUID
	PaymentDate   time.Time
}

func (e TransactionCreated) Type() string {
	return "TransactionCreated"
}

type RefundCreated struct {
	TransactionID uuid.UUID
	OrderID       uuid.UUID
	CustomerID    uuid.UUID
	PaymentDate   time.Time
}

func (e RefundCreated) Type() string {
	return "RefundCreated"
}

type CustomerAmountUpdated struct {
	CustomerID uuid.UUID
	NewAmount  float64
}

func (e CustomerAmountUpdated) Type() string {
	return "CustomerAmountUpdated"
}

type CustomerAccountCreated struct {
	CustomerID uuid.UUID
	CreatedAt  time.Time
}

func (e CustomerAccountCreated) Type() string {
	return "CustomerAccountCreated"
}
