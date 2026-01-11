package service

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	"inventory/pkg/common/domain"
	appmodel "inventory/pkg/inventory/app/model"
	"inventory/pkg/inventory/domain/model"
	"inventory/pkg/inventory/domain/service"
)

type ProductService interface {
	StoreProduct(ctx context.Context, product appmodel.Product) (uuid.UUID, error)
	IncreaseQuantity(ctx context.Context, ID uuid.UUID, quantity int) error
	DecreaseQuantity(ctx context.Context, ID uuid.UUID, quantity int) error
	FindProduct(ctx context.Context, productID uuid.UUID) (appmodel.Product, error)
}

func NewProductService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) ProductService {
	return &productService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type productService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (p productService) StoreProduct(ctx context.Context, product appmodel.Product) (uuid.UUID, error) {
	productID := product.ID
	err := p.luow.Execute(ctx, []string{product.ID.String()}, func(provider RepositoryProvider) error {
		domainService := p.domainService(ctx, provider.ProductRepository(ctx))

		if product.ID == uuid.Nil {
			uID, err := domainService.CreateProduct(product.Name, product.Quantity, product.Price)
			if err != nil {
				return err
			}
			productID = uID
			return nil
		}

		err := domainService.UpdateProductName(productID, product.Name)
		if err != nil {
			return err
		}

		err = domainService.UpdateProductPrice(productID, product.Price)
		if err != nil {
			return err
		}

		return nil
	})

	return productID, err
}

func (p productService) IncreaseQuantity(ctx context.Context, productID uuid.UUID, quantity int) error {
	return p.luow.Execute(ctx, []string{productID.String()}, func(provider RepositoryProvider) error {
		return p.domainService(ctx, provider.ProductRepository(ctx)).IncreaseQuantity(productID, quantity)
	})
}

func (p productService) DecreaseQuantity(ctx context.Context, productID uuid.UUID, quantity int) error {
	return p.luow.Execute(ctx, []string{productID.String()}, func(provider RepositoryProvider) error {
		return p.domainService(ctx, provider.ProductRepository(ctx)).DecreaseQuantity(productID, quantity)
	})
}

func (p productService) FindProduct(ctx context.Context, productID uuid.UUID) (appmodel.Product, error) {
	var product appmodel.Product
	err := p.luow.Execute(ctx, []string{productID.String()}, func(provider RepositoryProvider) error {
		domainProduct, err := provider.ProductRepository(ctx).Find(productID)
		if err != nil {
			return err
		}

		product = appmodel.Product{
			ID:       productID,
			Name:     domainProduct.Name,
			Quantity: domainProduct.Quantity,
			Price:    domainProduct.Price,
		}
		return nil
	})

	return product, err
}

func (p productService) domainService(ctx context.Context, repository model.ProductRepository) service.ProductService {
	return service.NewProductService(repository, p.domainEventDispatcher(ctx))
}

func (p productService) domainEventDispatcher(ctx context.Context) domain.EventDispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: p.eventDispatcher,
	}
}
