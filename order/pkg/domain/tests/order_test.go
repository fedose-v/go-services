package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"order/pkg/common/infrastructure/event"
	"order/pkg/domain/model"
	"order/pkg/domain/service"
)

func TestOrderService(t *testing.T) {
	repo := &mockOrderRepository{
		store: make(map[uuid.UUID]*model.Order),
	}
	eventDispatcher := &mockEventDispatcher{
		events: make([]event.Event, 0),
	}

	orderService := service.NewOrderService(repo, eventDispatcher)

	customerID := uuid.Must(uuid.NewV7())

	t.Run("Create order", func(t *testing.T) {
		orderID, err := orderService.CreateOrder(customerID)
		require.NoError(t, err)

		require.NotNil(t, repo.store[orderID])
		require.Equal(t, model.Open, repo.store[orderID].Status)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.OrderCreated{}.Type(), eventDispatcher.events[0].Type())
	})
	eventDispatcher.Reset()

	t.Run("Delete order", func(t *testing.T) {
		orderID, err := orderService.CreateOrder(customerID)
		require.NoError(t, err)

		err = orderService.DeleteOrder(orderID)
		require.NoError(t, err)

		require.NotNil(t, repo.store[orderID])
		require.NotNil(t, repo.store[orderID].DeletedAt)
		require.Equal(t, model.Deleted, repo.store[orderID].Status)
		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model.OrderCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.OrderDeleted{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Delete non existed order", func(t *testing.T) {
		newID, err := repo.NextID()
		require.NoError(t, err)

		err = orderService.DeleteOrder(newID)
		require.ErrorIs(t, err, model.ErrOrderNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})
	eventDispatcher.Reset()

	t.Run("Set order status", func(t *testing.T) {
		orderID, err := orderService.CreateOrder(customerID)
		require.NoError(t, err)

		err = orderService.SetStatus(orderID, model.Open)
		require.NoError(t, err)
		require.Equal(t, model.Open, repo.store[orderID].Status)

		err = orderService.SetStatus(orderID, model.Pending)
		require.NoError(t, err)
		require.Equal(t, model.Pending, repo.store[orderID].Status)

		err = orderService.SetStatus(orderID, model.Paid)
		require.NoError(t, err)
		require.Equal(t, model.Paid, repo.store[orderID].Status)

		err = orderService.SetStatus(orderID, model.Cancelled)
		require.NoError(t, err)
		require.Equal(t, model.Cancelled, repo.store[orderID].Status)

		err = orderService.SetStatus(orderID, model.Deleted)
		require.NoError(t, err)
		require.Equal(t, model.Deleted, repo.store[orderID].Status)

		require.Len(t, eventDispatcher.events, 6)
		require.Equal(t, model.OrderCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.OrderStatusChanged{}.Type(), eventDispatcher.events[1].Type())
		require.Equal(t, model.OrderStatusChanged{}.Type(), eventDispatcher.events[2].Type())
		require.Equal(t, model.OrderStatusChanged{}.Type(), eventDispatcher.events[3].Type())
		require.Equal(t, model.OrderStatusChanged{}.Type(), eventDispatcher.events[4].Type())
		require.Equal(t, model.OrderStatusChanged{}.Type(), eventDispatcher.events[5].Type())
	})
	eventDispatcher.Reset()

	t.Run("Add order item", func(t *testing.T) {
		orderID, err := orderService.CreateOrder(customerID)
		require.NoError(t, err)

		productID := uuid.Must(uuid.NewV7())
		itemID, err := orderService.AddItem(orderID, productID, 1.64)
		require.NoError(t, err)

		item := findItemByID(repo.store[orderID].Items, itemID)
		require.NotNil(t, item)
		require.Equal(t, productID, item.ProductID)
		require.Equal(t, 1.64, item.Price)

		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model.OrderCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.OrderItemChanged{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Add item to order with invalid status", func(t *testing.T) {
		orderID, err := orderService.CreateOrder(customerID)
		require.NoError(t, err)

		err = orderService.SetStatus(orderID, model.Pending)
		require.NoError(t, err)

		productID := uuid.Must(uuid.NewV7())
		itemID, err := orderService.AddItem(orderID, productID, 1.64)
		require.ErrorIs(t, err, service.ErrInvalidOrderStatus)

		item := findItemByID(repo.store[orderID].Items, itemID)
		require.Nil(t, item)

		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model.OrderCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.OrderStatusChanged{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Add item to non existed order", func(t *testing.T) {
		orderID := uuid.Must(uuid.NewV7())
		productID := uuid.Must(uuid.NewV7())

		itemID, err := orderService.AddItem(orderID, productID, 1.64)
		require.ErrorIs(t, err, model.ErrOrderNotFound)

		item := findItemByID(repo.store[orderID].Items, itemID)
		require.Nil(t, item)

		require.Len(t, eventDispatcher.events, 0)
	})
	eventDispatcher.Reset()
}

func findItemByID(items []model.Item, id uuid.UUID) *model.Item {
	for _, item := range items {
		if item.ID == id {
			return &item
		}
	}
	return nil
}

var _ model.OrderRepository = &mockOrderRepository{}

type mockOrderRepository struct {
	store map[uuid.UUID]*model.Order
}

func (m *mockOrderRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockOrderRepository) Store(order *model.Order) error {
	m.store[order.ID] = order
	return nil
}

func (m *mockOrderRepository) Find(id uuid.UUID) (*model.Order, error) {
	order, ok := m.store[id]
	if !ok {
		return nil, model.ErrOrderNotFound
	}
	if order.DeletedAt != nil {
		return nil, model.ErrOrderNotFound
	}
	return order, nil
}

func (m *mockOrderRepository) List() ([]model.Order, error) {
	var res []model.Order
	for _, order := range m.store {
		if order != nil && order.DeletedAt == nil {
			res = append(res, *order)
		}
	}
	return res, nil
}

func (m *mockOrderRepository) Delete(id uuid.UUID) error {
	order, ok := m.store[id]
	if !ok {
		return model.ErrOrderNotFound
	}
	now := time.Now()
	order.DeletedAt = &now
	return nil
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
