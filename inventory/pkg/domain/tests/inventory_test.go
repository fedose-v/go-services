package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"inventory/pkg/common/infrastructure/event"
	"inventory/pkg/domain/model"
	"inventory/pkg/domain/service"
)

func TestProductService(t *testing.T) {
	repo := &mockProductRepository{
		store: map[uuid.UUID]*model.Product{},
	}
	eventDispatcher := &mockEventDispatcher{}

	productService := service.NewProductService(repo, eventDispatcher)

	name := "Test Product"
	quantity := 1
	price := 24.9
	t.Run("Create product", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		require.NotNil(t, repo.store[productID])
		require.Equal(t, repo.store[productID].Name, "Test Product")
		require.Equal(t, repo.store[productID].Quantity, 1)
		require.Equal(t, repo.store[productID].Name, 24.9)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
	})
	eventDispatcher.Reset()

	t.Run("Increase product quantity", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		err = productService.IncreaseQuantity(productID, 10)
		require.NoError(t, err)

		require.NotNil(t, repo.store[productID])
		require.Equal(t, repo.store[productID].Name, "Test Product")
		require.Equal(t, repo.store[productID].Quantity, 11)
		require.Equal(t, repo.store[productID].Name, 24.9)
		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.ProductQuantityChanged{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Decrease product quantity", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		err = productService.DecreaseQuantity(productID, 1)
		require.NoError(t, err)

		require.NotNil(t, repo.store[productID])
		require.Equal(t, repo.store[productID].Name, "Test Product")
		require.Equal(t, repo.store[productID].Quantity, 0)
		require.Equal(t, repo.store[productID].Name, 24.9)
		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.ProductQuantityChanged{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Decrease product quantity to less than zero", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		err = productService.DecreaseQuantity(productID, 2)
		require.ErrorIs(t, err, model.ErrProductQuantityLessThanZero)

		require.NotNil(t, repo.store[productID])
		require.Equal(t, repo.store[productID].Name, "Test Product")
		require.Equal(t, repo.store[productID].Quantity, 1)
		require.Equal(t, repo.store[productID].Name, 24.9)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
	})
	eventDispatcher.Reset()

	t.Run("Delete product", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		err = productService.DeleteProduct(productID)
		require.NoError(t, err)

		require.Nil(t, repo.store[productID])
		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.ProductDeleted{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Delete non existed product", func(t *testing.T) {
		newID, _ := repo.NextID()
		err := productService.DeleteProduct(newID)
		require.ErrorIs(t, err, model.ErrProductNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})
	eventDispatcher.Reset()
}

var _ model.ProductRepository = &mockProductRepository{}

type mockProductRepository struct {
	store map[uuid.UUID]*model.Product
}

func (m mockProductRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m mockProductRepository) Store(product *model.Product) error {
	m.store[product.ID] = product
	return nil
}

func (m mockProductRepository) Find(id uuid.UUID) (*model.Product, error) {
	if product, ok := m.store[id]; ok && product.DeletedAt != nil {
		return product, nil
	}
	return nil, model.ErrProductNotFound
}

func (m mockProductRepository) List() (*[]model.Product, error) {
	var res []model.Product
	for _, v := range m.store {
		if v != nil {
			res = append(res, *v)
		}
	}
	return &res, nil
}

func (m mockProductRepository) Delete(id uuid.UUID) error {
	if product, ok := m.store[id]; ok && product.DeletedAt != nil {
		product.DeletedAt = toPtr(time.Now())
		return nil
	}
	return model.ErrProductNotFound
}

type MockEventDispatcher interface {
	event.Dispatcher
	Reset()
}

var _ MockEventDispatcher = &mockEventDispatcher{}

type mockEventDispatcher struct {
	events []event.Event
}

func (m mockEventDispatcher) Reset() {
	m.events = nil
}

func (m mockEventDispatcher) Dispatch(event event.Event) error {
	m.events = append(m.events, event)
	return nil
}

func toPtr[V any](v V) *V {
	return &v
}
