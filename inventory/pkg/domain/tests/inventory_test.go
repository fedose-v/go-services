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
		store: make(map[uuid.UUID]*model.Product),
	}
	eventDispatcher := &mockEventDispatcher{
		events: make([]event.Event, 0),
	}

	productService := service.NewProductService(repo, eventDispatcher)

	name := "Test Product"
	quantity := 1
	price := 24.9

	t.Run("Create product", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		require.NotNil(t, repo.store[productID])
		require.Equal(t, "Test Product", repo.store[productID].Name)
		require.Equal(t, 1, repo.store[productID].Quantity)
		require.Equal(t, 24.9, repo.store[productID].Price)
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
		require.Equal(t, "Test Product", repo.store[productID].Name)
		require.Equal(t, 11, repo.store[productID].Quantity)
		require.Equal(t, 24.9, repo.store[productID].Price)
		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.ProductQuantityChanged{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Increase non existed product quantity", func(t *testing.T) {
		err := productService.IncreaseQuantity(uuid.New(), 10)
		require.ErrorIs(t, err, model.ErrProductNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})
	eventDispatcher.Reset()

	t.Run("Decrease product quantity", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		err = productService.DecreaseQuantity(productID, 1)
		require.NoError(t, err)

		require.NotNil(t, repo.store[productID])
		require.Equal(t, "Test Product", repo.store[productID].Name)
		require.Equal(t, 0, repo.store[productID].Quantity)
		require.Equal(t, 24.9, repo.store[productID].Price)
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
		require.Equal(t, "Test Product", repo.store[productID].Name)
		require.Equal(t, 1, repo.store[productID].Quantity)
		require.Equal(t, 24.9, repo.store[productID].Price)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
	})
	eventDispatcher.Reset()

	t.Run("Decrease non existed product quantity", func(t *testing.T) {
		err := productService.DecreaseQuantity(uuid.New(), 10)
		require.ErrorIs(t, err, model.ErrProductNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})
	eventDispatcher.Reset()

	t.Run("Delete product", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		err = productService.DeleteProduct(productID)
		require.NoError(t, err)

		require.NotNil(t, repo.store[productID])
		require.NotNil(t, repo.store[productID].DeletedAt)
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

func (m *mockProductRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockProductRepository) Store(product *model.Product) error {
	m.store[product.ID] = product
	return nil
}

func (m *mockProductRepository) Find(id uuid.UUID) (*model.Product, error) {
	product, ok := m.store[id]
	if !ok {
		return nil, model.ErrProductNotFound
	}
	if product.DeletedAt != nil {
		return nil, model.ErrProductNotFound
	}
	return product, nil
}

func (m *mockProductRepository) List() ([]model.Product, error) {
	var res []model.Product
	for _, product := range m.store {
		if product != nil && product.DeletedAt == nil {
			res = append(res, *product)
		}
	}
	return res, nil
}

func (m *mockProductRepository) Delete(id uuid.UUID) error {
	product, ok := m.store[id]
	if !ok {
		return model.ErrProductNotFound
	}
	now := time.Now()
	product.DeletedAt = &now
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
