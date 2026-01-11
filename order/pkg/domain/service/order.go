package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	commonevent "order/pkg/common/event"
	"order/pkg/domain/model"
)

var (
	ErrInvalidOrderStatus = errors.New("invalid order status")
)

type Order interface {
	CreateOrder(customerID uuid.UUID) (uuid.UUID, error)
	DeleteOrder(orderID uuid.UUID) error
	SetStatus(orderID uuid.UUID, status model.OrderStatus) error

	AddItem(orderID uuid.UUID, productID uuid.UUID, price float64) (uuid.UUID, error)
	DeleteItem(orderID uuid.UUID, itemID uuid.UUID) error
}

func NewOrderService(repo model.OrderRepository, dispatcher commonevent.Dispatcher) Order {
	return &orderService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type orderService struct {
	repo       model.OrderRepository
	dispatcher commonevent.Dispatcher
}

func (o orderService) CreateOrder(customerID uuid.UUID) (uuid.UUID, error) {
	orderID, err := o.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	err = o.repo.Store(&model.Order{
		ID:         orderID,
		CustomerID: customerID,
		Status:     model.Open,
		CreatedAt:  currentTime,
		UpdatedAt:  currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return orderID, o.dispatcher.Dispatch(model.OrderCreated{
		OrderID:    orderID,
		CustomerID: customerID,
	})
}

func (o orderService) DeleteOrder(orderID uuid.UUID) error {
	order, err := o.repo.Find(model.FindSpec{OrderID: &orderID})
	if err != nil {
		return model.ErrOrderNotFound
	}

	deletedAt := time.Now()
	order.DeletedAt = &deletedAt
	order.Status = model.Deleted

	err = o.repo.Store(order)
	if err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.OrderDeleted{
		OrderID:   orderID,
		DeletedAt: deletedAt,
	})
}

func (o orderService) SetStatus(orderID uuid.UUID, status model.OrderStatus) error {
	order, err := o.repo.Find(model.FindSpec{OrderID: &orderID})
	if err != nil {
		return model.ErrOrderNotFound
	}

	order.Status = status

	err = o.repo.Store(order)
	if err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.OrderStatusChanged{
		OrderID:     orderID,
		OrderStatus: status,
	})
}

func (o orderService) AddItem(orderID uuid.UUID, productID uuid.UUID, price float64) (uuid.UUID, error) {
	order, err := o.repo.Find(model.FindSpec{OrderID: &orderID})
	if err != nil {
		return uuid.Nil, err
	}

	if order.Status != model.Open {
		return uuid.Nil, ErrInvalidOrderStatus
	}

	itemID, err := o.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}
	order.Items = append(order.Items, model.Item{
		ID:        itemID,
		ProductID: productID,
		Price:     price,
	})
	err = o.repo.Store(order)
	if err != nil {
		return uuid.Nil, err
	}

	return itemID, o.dispatcher.Dispatch(model.OrderItemChanged{
		OrderID:    orderID,
		AddedItems: []uuid.UUID{itemID},
	})
}

func (o orderService) DeleteItem(orderID uuid.UUID, itemID uuid.UUID) error {
	order, err := o.repo.Find(model.FindSpec{OrderID: &orderID})
	if err != nil {
		return model.ErrOrderNotFound
	}

	newItems := removeItemByID(order.Items, itemID)
	if len(newItems) == len(order.Items) {
		return nil
	}
	order.Items = newItems

	err = o.repo.Store(order)
	if err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.OrderItemChanged{
		OrderID:      orderID,
		RemovedItems: []uuid.UUID{itemID},
	})
}

func removeItemByID(items []model.Item, id uuid.UUID) []model.Item {
	result := make([]model.Item, 0)
	for _, item := range items {
		if item.ID != id {
			result = append(result, item)
		}
	}
	return result
}
