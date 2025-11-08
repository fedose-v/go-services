package tests

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"

	"payment/pkg/common/infrastructure/event"
	"payment/pkg/domain/model"
	"payment/pkg/domain/service"
)

func TestPaymentService(t *testing.T) {
	paymentRepo := &mockPaymentRepository{
		store: make(map[uuid.UUID]*model.Transaction),
	}
	balanceRepo := &mockCustomerBalanceRepository{
		store: make(map[uuid.UUID]*model.CustomerAccountBalance),
	}
	eventDispatcher := &mockEventDispatcher{
		events: make([]event.Event, 0),
	}

	paymentService := service.NewPaymentService(paymentRepo, balanceRepo, eventDispatcher)

	customerID := uuid.Must(uuid.NewV7())
	orderID := uuid.Must(uuid.NewV7())

	t.Run("Create customer balance", func(t *testing.T) {
		t.Cleanup(func() {
			eventDispatcher.Reset()
			paymentRepo.Reset()
			balanceRepo.Reset()
		})
		err := paymentService.CreateCustomerBalance(customerID)
		require.NoError(t, err)

		balance, err := balanceRepo.Find(customerID)
		require.NoError(t, err)
		require.Equal(t, customerID, balance.CustomerID)
		require.Equal(t, 0.0, balance.Amount)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.CustomerAccountCreated{}.Type(), eventDispatcher.events[0].Type())
	})

	t.Run("Create already exist customer balance", func(t *testing.T) {
		t.Cleanup(func() {
			eventDispatcher.Reset()
			paymentRepo.Reset()
			balanceRepo.Reset()
		})
		err := paymentService.CreateCustomerBalance(customerID)
		require.NoError(t, err)

		err = paymentService.CreateCustomerBalance(customerID)
		require.ErrorIs(t, err, service.ErrBalanceExisted)

		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.CustomerAccountCreated{}.Type(), eventDispatcher.events[0].Type())
	})

	t.Run("Add amount to balance", func(t *testing.T) {
		t.Cleanup(func() {
			eventDispatcher.Reset()
			paymentRepo.Reset()
			balanceRepo.Reset()
		})
		err := paymentService.CreateCustomerBalance(customerID)
		require.NoError(t, err)
		eventDispatcher.Reset()

		amountToAdd := 100.0
		err = paymentService.AddAmountToBalance(customerID, amountToAdd)
		require.NoError(t, err)

		balance, err := balanceRepo.Find(customerID)
		require.Equal(t, amountToAdd, balance.Amount)

		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.CustomerAmountUpdated{}.Type(), eventDispatcher.events[0].Type())

		if event, ok := eventDispatcher.events[0].(model.CustomerAmountUpdated); ok {
			require.Equal(t, customerID, event.CustomerID)
			require.Equal(t, amountToAdd, event.NewAmount)
		}
	})

	t.Run("Add negative amount to balance", func(t *testing.T) {
		t.Cleanup(func() {
			eventDispatcher.Reset()
			paymentRepo.Reset()
			balanceRepo.Reset()
		})
		err := paymentService.AddAmountToBalance(customerID, -50.0)
		require.ErrorIs(t, err, service.ErrAddingNegativeAmount)

		require.Len(t, eventDispatcher.events, 0)
	})

	t.Run("Create transaction", func(t *testing.T) {
		t.Cleanup(func() {
			eventDispatcher.Reset()
			paymentRepo.Reset()
			balanceRepo.Reset()
		})
		err := paymentService.CreateCustomerBalance(customerID)
		require.NoError(t, err)
		err = paymentService.AddAmountToBalance(customerID, 200.0)
		require.NoError(t, err)
		eventDispatcher.Reset()

		amount := 50.0
		transactionID, err := paymentService.CreateTransaction(orderID, customerID, amount)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, transactionID)

		transaction, err := paymentRepo.Find(transactionID)
		require.NoError(t, err)
		require.Equal(t, orderID, transaction.OrderID)
		require.Equal(t, customerID, transaction.CustomerID)
		require.Equal(t, model.New, transaction.Type)
		require.Equal(t, amount, transaction.Amount)

		balance, err := balanceRepo.Find(customerID)
		require.NoError(t, err)
		require.Equal(t, 150.0, balance.Amount)

		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.TransactionCreated{}.Type(), eventDispatcher.events[0].Type())

		if event, ok := eventDispatcher.events[0].(model.TransactionCreated); ok {
			require.Equal(t, orderID, event.OrderID)
			require.Equal(t, customerID, event.CustomerID)
		}
	})

	t.Run("Create transaction with insufficient funds", func(t *testing.T) {
		t.Cleanup(func() {
			eventDispatcher.Reset()
			paymentRepo.Reset()
			balanceRepo.Reset()
		})
		insufficientCustomerID := uuid.Must(uuid.NewV7())
		err := paymentService.CreateCustomerBalance(insufficientCustomerID)
		require.NoError(t, err)
		err = paymentService.AddAmountToBalance(insufficientCustomerID, 10.0)
		require.NoError(t, err)
		eventDispatcher.Reset()

		amount := 50.0
		_, err = paymentService.CreateTransaction(orderID, insufficientCustomerID, amount)
		require.ErrorIs(t, err, service.ErrNotEnoughAmount)

		require.Len(t, eventDispatcher.events, 0)
	})

	t.Run("Create refund", func(t *testing.T) {
		t.Cleanup(func() {
			eventDispatcher.Reset()
			paymentRepo.Reset()
			balanceRepo.Reset()
		})
		err := paymentService.CreateCustomerBalance(customerID)
		require.NoError(t, err)

		originalBalance := 100.0
		err = paymentService.AddAmountToBalance(customerID, originalBalance)
		require.NoError(t, err)
		eventDispatcher.Reset()

		amount := 30.0
		transactionID, err := paymentService.CreateRefund(orderID, customerID, amount)
		require.NoError(t, err)

		transaction, err := paymentRepo.Find(transactionID)
		require.NoError(t, err)
		require.Equal(t, orderID, transaction.OrderID)
		require.Equal(t, customerID, transaction.CustomerID)
		require.Equal(t, model.Refund, transaction.Type)
		require.Equal(t, amount, transaction.Amount)

		balance, err := balanceRepo.Find(customerID)
		require.NoError(t, err)
		require.Equal(t, originalBalance+amount, balance.Amount)

		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.RefundCreated{}.Type(), eventDispatcher.events[0].Type())

		e := eventDispatcher.events[0].(model.RefundCreated)
		require.Equal(t, orderID, e.OrderID)
		require.Equal(t, customerID, e.CustomerID)

	})

	t.Run("Create transaction when customer balance not found", func(t *testing.T) {
		t.Cleanup(func() {
			eventDispatcher.Reset()
			paymentRepo.Reset()
			balanceRepo.Reset()
		})
		nonExistentCustomerID := uuid.Must(uuid.NewV7())
		amount := 50.0
		_, err := paymentService.CreateTransaction(orderID, nonExistentCustomerID, amount)
		require.ErrorIs(t, err, model.ErrPaymentNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})

	t.Run("Create refund when customer balance not found", func(t *testing.T) {
		t.Cleanup(func() {
			eventDispatcher.Reset()
			paymentRepo.Reset()
			balanceRepo.Reset()
		})
		nonExistentCustomerID := uuid.Must(uuid.NewV7())
		amount := 50.0
		_, err := paymentService.CreateRefund(orderID, nonExistentCustomerID, amount)
		require.ErrorIs(t, err, model.ErrPaymentNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})

	t.Run("Add amount to balance when customer not found", func(t *testing.T) {
		t.Cleanup(func() {
			eventDispatcher.Reset()
			paymentRepo.Reset()
			balanceRepo.Reset()
		})
		nonExistentCustomerID := uuid.Must(uuid.NewV7())
		amount := 50.0
		err := paymentService.AddAmountToBalance(nonExistentCustomerID, amount)
		require.ErrorIs(t, err, model.ErrPaymentNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})
}

var _ model.PaymentRepository = &mockPaymentRepository{}

type mockPaymentRepository struct {
	store map[uuid.UUID]*model.Transaction
}

func (m *mockPaymentRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockPaymentRepository) Store(transaction *model.Transaction) error {
	m.store[transaction.ID] = transaction
	return nil
}

func (m *mockPaymentRepository) Find(id uuid.UUID) (*model.Transaction, error) {
	transaction, ok := m.store[id]
	if !ok {
		return nil, model.ErrPaymentNotFound
	}
	return transaction, nil
}

func (m *mockPaymentRepository) Delete(id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockPaymentRepository) Reset() {
	m.store = make(map[uuid.UUID]*model.Transaction)
}

var _ model.CustomerBalanceRepository = &mockCustomerBalanceRepository{}

type mockCustomerBalanceRepository struct {
	store map[uuid.UUID]*model.CustomerAccountBalance
}

func (m *mockCustomerBalanceRepository) Store(balance *model.CustomerAccountBalance) error {
	m.store[balance.CustomerID] = balance
	return nil
}

func (m *mockCustomerBalanceRepository) Find(customerID uuid.UUID) (*model.CustomerAccountBalance, error) {
	balance, ok := m.store[customerID]
	if !ok {
		return nil, model.ErrPaymentNotFound
	}
	return balance, nil
}

func (m *mockCustomerBalanceRepository) Reset() {
	m.store = make(map[uuid.UUID]*model.CustomerAccountBalance)
}

type mockEventDispatcher struct {
	events []event.Event
}

func (m *mockEventDispatcher) Reset() {
	m.events = make([]event.Event, 0)
}

func (m *mockEventDispatcher) ListEvents() []event.Event {
	return m.events
}

func (m *mockEventDispatcher) Dispatch(evt event.Event) error {
	m.events = append(m.events, evt)
	return nil
}
