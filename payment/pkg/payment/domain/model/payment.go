package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrBalanceNotFound = errors.New("balance not found")
)

type TransactionType int

const (
	New TransactionType = iota
	Refund
)

type CustomerAccountBalance struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Amount     float64
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

type Transaction struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	CustomerID  uuid.UUID
	Type        TransactionType
	Amount      float64
	PaymentDate time.Time
}

type PaymentRepository interface {
	NextID() (uuid.UUID, error)
	Store(transaction *Transaction) error
	Find(id uuid.UUID) (*Transaction, error)
}

type CustomerBalanceRepository interface {
	Store(balance CustomerAccountBalance) (uuid.UUID, error)
	Find(customerID uuid.UUID) (*CustomerAccountBalance, error)
}
