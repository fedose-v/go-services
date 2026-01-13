package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"payment/pkg/common/domain"
	"payment/pkg/payment/domain/model"
)

var (
	ErrAddingNegativeAmount = errors.New("adding negative amount")
	ErrNotEnoughAmount      = errors.New("not enough amount")
	ErrBalanceExisted       = errors.New("balance existed")
)

type PaymentService interface {
	CreateTransaction(orderID uuid.UUID, customerID uuid.UUID, amount float64) (uuid.UUID, error)
	CreateRefund(orderID uuid.UUID, customerID uuid.UUID, amount float64) (uuid.UUID, error)

	CreateCustomerBalance(customerID uuid.UUID) (uuid.UUID, error)
	UpdateBalance(customerID uuid.UUID, amount float64) error
}

func NewPaymentService(repo model.PaymentRepository, accountBalanceRepo model.CustomerBalanceRepository, dispatcher domain.EventDispatcher) PaymentService {
	return &paymentService{
		paymentRepo: repo,
		balanceRepo: accountBalanceRepo,
		dispatcher:  dispatcher,
	}
}

type paymentService struct {
	paymentRepo model.PaymentRepository
	balanceRepo model.CustomerBalanceRepository
	dispatcher  domain.EventDispatcher
}

func (p paymentService) CreateTransaction(orderID, customerID uuid.UUID, amount float64) (uuid.UUID, error) {
	if amount < 0 {
		return uuid.Nil, ErrAddingNegativeAmount
	}

	balance, err := p.balanceRepo.Find(customerID)
	if err != nil {
		return uuid.Nil, model.ErrBalanceNotFound
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
	_, err = p.balanceRepo.Store(balance)
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

func (p paymentService) CreateRefund(orderID, customerID uuid.UUID, amount float64) (uuid.UUID, error) {
	if amount < 0 {
		return uuid.Nil, ErrAddingNegativeAmount
	}

	balance, err := p.balanceRepo.Find(customerID)
	if err != nil {
		return uuid.Nil, model.ErrBalanceNotFound
	}

	currentTime := time.Now()
	balance.Amount += amount
	balance.UpdatedAt = &currentTime
	_, err = p.balanceRepo.Store(balance)
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

	return transactionID, p.dispatcher.Dispatch(model.RefundCreated{
		TransactionID: transactionID,
		OrderID:       orderID,
		CustomerID:    customerID,
		PaymentDate:   currentTime,
	})
}

func (p paymentService) CreateCustomerBalance(customerID uuid.UUID) (uuid.UUID, error) {
	_, err := p.balanceRepo.Find(customerID)
	if err == nil {
		return uuid.Nil, ErrBalanceExisted
	}

	currentTime := time.Now()
	balanceID, err := p.balanceRepo.Store(&model.CustomerAccountBalance{
		CustomerID: customerID,
		Amount:     0,
		CreatedAt:  currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return balanceID, p.dispatcher.Dispatch(model.CustomerAccountCreated{
		CustomerID: customerID,
		CreatedAt:  currentTime,
	})
}

func (p paymentService) UpdateBalance(customerID uuid.UUID, amount float64) error {
	if amount < 0 {
		return ErrAddingNegativeAmount
	}

	balance, err := p.balanceRepo.Find(customerID)
	if err != nil {
		return model.ErrBalanceNotFound
	}

	currentTime := time.Now()
	balance.Amount = amount
	balance.UpdatedAt = &currentTime
	_, err = p.balanceRepo.Store(balance)
	if err != nil {
		return err
	}

	return p.dispatcher.Dispatch(model.CustomerAmountUpdated{
		CustomerID: customerID,
		NewAmount:  amount,
	})
}
