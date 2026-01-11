package service

import (
	"context"
	"errors"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	"order/pkg/domain/model"
	"order/pkg/domain/service"
)

type ProductProvider interface {
	ActualPrice(ctx context.Context, productID uuid.UUID) (float64, error)
}

type OrderService interface {
	AddProductToOrder(ctx context.Context, customerID uuid.UUID, itemID uuid.UUID) (uuid.UUID, error)
	RemoveProductFromOrder(ctx context.Context, customerID uuid.UUID, itemID uuid.UUID) error

	SaveOrder(ctx context.Context, orderID uuid.UUID) error
	CancelOrder(ctx context.Context, orderID uuid.UUID) error
	DeleteOrder(ctx context.Context, orderID uuid.UUID) error
}

func NewOrderService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
	productProvider *ProductProvider,
) OrderService {
	return &orderService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
		productProvider: productProvider,
	}
}

type orderService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
	productProvider *ProductProvider
}

func (o orderService) AddProductToOrder(ctx context.Context, customerID, productID uuid.UUID) (uuid.UUID, error) {
	var orderID uuid.UUID
	err := o.luow.Execute(ctx, []string{"customer_" + customerID.String()}, func(provider RepositoryProvider) error {
		status := model.Open
		order, err := provider.OrderRepository(ctx).Find(model.FindSpec{
			CustomerID: &customerID,
			Status:     &status,
		})
		if err == nil {
			orderID = order.ID
			return nil
		}
		if err != nil && !errors.Is(err, model.ErrOrderNotFound) {
			return err
		}

		domainService := o.domainService(ctx, provider.OrderRepository(ctx))
		orderID, err = domainService.CreateOrder(customerID)
		return err
	})
	if err != nil {
		return uuid.Nil, err
	}

	price := 0.0
	//price, err := o.productProvider.ActualPrice(ctx, productID)
	//if err != nil {
	//	return uuid.Nil, err
	//}

	var itemID uuid.UUID
	return itemID, o.luow.Execute(ctx, []string{"order_" + orderID.String()}, func(provider RepositoryProvider) error {
		domainService := o.domainService(ctx, provider.OrderRepository(ctx))
		itemID, err = domainService.AddItem(orderID, productID, price)
		return err
	})
}

func (o orderService) RemoveProductFromOrder(ctx context.Context, customerID, itemID uuid.UUID) error {
	var orderID uuid.UUID
	err := o.luow.Execute(ctx, []string{"customer_" + customerID.String()}, func(provider RepositoryProvider) error {
		order, err := provider.OrderRepository(ctx).Find(model.FindSpec{
			CustomerID: &customerID,
		})
		if err != nil {
			return err
		}

		if order.CustomerID != customerID {
			return model.ErrOrderAccessDenied
		}

		orderID = order.ID
		return nil
	})
	if err != nil {
		return err
	}

	return o.luow.Execute(ctx, []string{"order_" + orderID.String()}, func(provider RepositoryProvider) error {
		domainService := o.domainService(ctx, provider.OrderRepository(ctx))
		return domainService.DeleteItem(orderID, itemID)
	})
}

func (o orderService) SaveOrder(ctx context.Context, orderID uuid.UUID) error {
	return o.luow.Execute(ctx, []string{"order_" + orderID.String()}, func(provider RepositoryProvider) error {
		order, err := provider.OrderRepository(ctx).Find(model.FindSpec{
			OrderID: &orderID,
		})
		if err != nil {
			return err
		}

		if order.Status != model.Open {
			return model.ErrInvalidOrderStatus
		}

		order.Status = model.Pending
		order.UpdatedAt = time.Now()

		return provider.OrderRepository(ctx).Store(order)
	})
}

func (o orderService) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	return o.luow.Execute(ctx, []string{"order_" + orderID.String()}, func(provider RepositoryProvider) error {
		order, err := provider.OrderRepository(ctx).Find(model.FindSpec{
			OrderID: &orderID,
		})
		if err != nil {
			return err
		}

		if order.Status != model.Open && order.Status != model.Pending {
			return model.ErrInvalidOrderStatus
		}

		order.Status = model.Cancelled
		order.UpdatedAt = time.Now()

		return provider.OrderRepository(ctx).Store(order)
	})
}

func (o orderService) DeleteOrder(ctx context.Context, orderID uuid.UUID) error {
	return o.luow.Execute(ctx, []string{"order_" + orderID.String()}, func(provider RepositoryProvider) error {
		_, err := provider.OrderRepository(ctx).Find(model.FindSpec{
			OrderID: &orderID,
		})
		if err != nil {
			return err
		}

		return provider.OrderRepository(ctx).Delete(orderID)
	})
}

func (o orderService) domainService(ctx context.Context, repo model.OrderRepository) service.Order {
	return service.NewOrderService(
		repo,
		&domainEventDispatcher{
			ctx:             ctx,
			eventDispatcher: o.eventDispatcher,
		},
	)
}
