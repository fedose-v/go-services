package service

import (
	"github.com/google/uuid"
	"time"

	"inventory/pkg/common/infrastructure/event"
	"inventory/pkg/domain/model"
)

type Product interface {
	CreateProduct(name string, quantity int, price float64) (uuid.UUID, error)
	IncreaseQuantity(ID uuid.UUID, quantity int) error
	DecreaseQuantity(ID uuid.UUID, quantity int) error

	DeleteProduct(ID uuid.UUID) error
}

func NewProductService(repo model.ProductRepository, d event.Dispatcher) Product {
	return &productService{repo, d}
}

type productService struct {
	repo model.ProductRepository
	d    event.Dispatcher
}

func (p productService) CreateProduct(name string, quantity int, price float64) (uuid.UUID, error) {
	newProductID, err := p.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}
	if quantity < 0 {
		return uuid.Nil, model.ErrProductQuantityLessThanZero
	}
	currentTime := time.Now()
	err = p.repo.Store(&model.Product{
		ID:        newProductID,
		Name:      name,
		Quantity:  quantity,
		Price:     price,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return newProductID, p.d.Dispatch(model.ProductCreated{
		ID:        newProductID,
		Name:      name,
		CreatedAt: currentTime,
	})
}

func (p productService) IncreaseQuantity(productID uuid.UUID, quantity int) error {
	product, err := p.repo.Find(productID)
	if err != nil {
		return err
	}

	product.Quantity += quantity
	err = p.repo.Store(product)
	if err != nil {
		return err
	}

	return p.d.Dispatch(model.ProductQuantityChanged{
		ID:           productID,
		NewQuantity:  product.Quantity,
		PrevQuantity: product.Quantity - quantity,
	})
}

func (p productService) DecreaseQuantity(productID uuid.UUID, quantity int) error {
	product, err := p.repo.Find(productID)
	if err != nil {
		return err
	}
	if (product.Quantity - quantity) < 0 {
		return model.ErrProductQuantityLessThanZero
	}

	product.Quantity -= quantity
	err = p.repo.Store(product)
	if err != nil {
		return err
	}

	return p.d.Dispatch(model.ProductQuantityChanged{
		ID:           productID,
		NewQuantity:  product.Quantity,
		PrevQuantity: product.Quantity + quantity,
	})
}

func (p productService) DeleteProduct(ID uuid.UUID) error {
	err := p.repo.Delete(ID)
	if err != nil {
		return err
	}

	return p.d.Dispatch(model.ProductDeleted{
		ID: ID,
	})
}
