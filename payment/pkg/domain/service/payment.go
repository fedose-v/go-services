package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"payment/pkg/common/infrastructure/event"
	"payment/pkg/domain/model"
)

var (
	ErrAddingNegativeAmount = errors.New("adding negative amount")
	ErrNotEnoughAmount      = errors.New("not enough amount")
	ErrBalanceExisted       = errors.New("balance existed")
)

type Payment interface {
	CreateTransaction(orderID uuid.UUID, customerID uuid.UUID, amount float64) (uuid.UUID, error)
	CreateRefund(orderID uuid.UUID, customerID uuid.UUID, amount float64) (uuid.UUID, error)

	CreateCustomerBalance(customerID uuid.UUID) error
	AddAmountToBalance(customerID uuid.UUID, amount float64) error
}

func NewPaymentService(repo model.PaymentRepository, accountBalanceRepo model.CustomerBalanceRepository, dispatcher event.Dispatcher) Payment {
	return &paymentService{
		paymentRepo: repo,
		balanceRepo: accountBalanceRepo,
		dispatcher:  dispatcher,
	}
}

type paymentService struct {
	paymentRepo model.PaymentRepository
	balanceRepo model.CustomerBalanceRepository
	dispatcher  event.Dispatcher
}

func (p paymentService) CreateTransaction(orderID uuid.UUID, customerID uuid.UUID, amount float64) (uuid.UUID, error) {
	balance, err := p.balanceRepo.Find(customerID)
	if err != nil {
		return uuid.Nil, err
	}

	if balance.Amount < amount {
		return uuid.Nil, ErrNotEnoughAmount
	}

	transactionID, err := p.paymentRepo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	balance.Amount -= amount
	err = p.balanceRepo.Store(balance)
	if err != nil {
		return uuid.Nil, err
	}

	err = p.paymentRepo.Store(&model.Transaction{
		ID:          transactionID,
		OrderID:     orderID,
		CustomerID:  customerID,
		Type:        model.New,
		Amount:      amount,
		PaymentDate: currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return transactionID, p.dispatcher.Dispatch(model.TransactionCreated{
		TransactionID: transactionID,
		OrderID:       orderID,
		CustomerID:    customerID,
		PaymentDate:   currentTime,
	})
}

func (p paymentService) CreateRefund(orderID uuid.UUID, customerID uuid.UUID, amount float64) (uuid.UUID, error) {
	balance, err := p.balanceRepo.Find(customerID)
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	balance.Amount += amount
	balance.UpdatedAt = &currentTime
	err = p.balanceRepo.Store(balance)
	if err != nil {
		return uuid.Nil, err
	}

	transactionID, err := p.paymentRepo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	err = p.paymentRepo.Store(&model.Transaction{
		ID:          transactionID,
		OrderID:     orderID,
		CustomerID:  customerID,
		Type:        model.Refund,
		Amount:      amount,
		PaymentDate: currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return orderID, p.dispatcher.Dispatch(model.RefundCreated{
		TransactionID: transactionID,
		OrderID:       orderID,
		CustomerID:    customerID,
		PaymentDate:   currentTime,
	})
}

func (p paymentService) AddAmountToBalance(customerID uuid.UUID, amount float64) error {
	if amount < 0 {
		return ErrAddingNegativeAmount
	}

	balance, err := p.balanceRepo.Find(customerID)
	if err != nil {
		return err
	}

	currentTime := time.Now()
	newAmount := balance.Amount + amount
	balance.Amount = newAmount
	balance.UpdatedAt = &currentTime
	err = p.balanceRepo.Store(balance)
	if err != nil {
		return err
	}

	return p.dispatcher.Dispatch(model.CustomerAmountUpdated{
		CustomerID: customerID,
		NewAmount:  newAmount,
	})
}

func (p paymentService) CreateCustomerBalance(customerID uuid.UUID) error {
	_, err := p.balanceRepo.Find(customerID)
	if err == nil {
		return ErrBalanceExisted
	}

	currentTime := time.Now()
	err = p.balanceRepo.Store(&model.CustomerAccountBalance{
		CustomerID: customerID,
		Amount:     0,
		CreatedAt:  currentTime,
	})
	if err != nil {
		return err
	}

	return p.dispatcher.Dispatch(model.CustomerAccountCreated{
		CustomerID: customerID,
		CreatedAt:  currentTime,
	})
}
