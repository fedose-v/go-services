package service

import (
	"time"

	"github.com/google/uuid"

	"inventory/pkg/common/domain"
	"inventory/pkg/inventory/domain/model"
)

type ProductService interface {
	CreateProduct(name string, quantity int, price float64) (uuid.UUID, error)
	IncreaseQuantity(productID uuid.UUID, quantity int) error
	DecreaseQuantity(productID uuid.UUID, quantity int) error
	UpdateProductName(productID uuid.UUID, newName string) error
	UpdateProductPrice(productID uuid.UUID, newPrice float64) error

	DeleteProduct(id uuid.UUID) error
}

func NewProductService(repo model.ProductRepository, d domain.EventDispatcher) ProductService {
	return &productService{repo, d}
}

type productService struct {
	repo            model.ProductRepository
	eventDispatcher domain.EventDispatcher
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

	return newProductID, p.eventDispatcher.Dispatch(&model.ProductCreated{
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

	return p.eventDispatcher.Dispatch(&model.ProductQuantityChanged{
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

	return p.eventDispatcher.Dispatch(&model.ProductQuantityChanged{
		ID:           productID,
		NewQuantity:  product.Quantity,
		PrevQuantity: product.Quantity + quantity,
	})
}

func (p productService) UpdateProductName(productID uuid.UUID, newName string) error {
	product, err := p.repo.Find(productID)
	if err != nil {
		return err
	}

	product.Name = newName
	err = p.repo.Store(product)
	if err != nil {
		return err
	}

	return p.eventDispatcher.Dispatch(&model.ProductNameChanged{
		ID:   product.ID,
		Name: newName,
	})
}

func (p productService) UpdateProductPrice(productID uuid.UUID, newPrice float64) error {
	product, err := p.repo.Find(productID)
	if err != nil {
		return err
	}

	product.Price = newPrice
	err = p.repo.Store(product)
	if err != nil {
		return err
	}

	return p.eventDispatcher.Dispatch(&model.ProductPriceChanged{
		ID:    product.ID,
		Price: product.Price,
	})
}

func (p productService) DeleteProduct(productID uuid.UUID) error {
	err := p.repo.Delete(productID)
	if err != nil {
		return err
	}

	return p.eventDispatcher.Dispatch(&model.ProductDeleted{
		ProductID: productID,
	})
}
