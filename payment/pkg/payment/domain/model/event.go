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
	return "transaction_created"
}

type RefundCreated struct {
	TransactionID uuid.UUID
	OrderID       uuid.UUID
	CustomerID    uuid.UUID
	PaymentDate   time.Time
}

func (e RefundCreated) Type() string {
	return "refund_created"
}

type CustomerAmountUpdated struct {
	CustomerID uuid.UUID
	NewAmount  float64
}

func (e CustomerAmountUpdated) Type() string {
	return "customer_amount_updated"
}

type CustomerAccountCreated struct {
	CustomerID uuid.UUID
	CreatedAt  time.Time
}

func (e CustomerAccountCreated) Type() string {
	return "customer_account_created"
}
