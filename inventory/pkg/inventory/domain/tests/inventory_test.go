package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"inventory/pkg/common/domain"

	model2 "inventory/pkg/inventory/domain/model"
	"inventory/pkg/inventory/domain/service"
)

func TestProductService(t *testing.T) {
	repo := &mockProductRepository{
		store: make(map[uuid.UUID]*model2.Product),
	}
	eventDispatcher := &mockEventDispatcher{
		events: make([]domain.Event, 0),
	}

	productService := service.NewProductService(repo, eventDispatcher)

	name := "Test ProductService"
	quantity := 1
	price := 24.9

	t.Run("Create product", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		require.NotNil(t, repo.store[productID])
		require.Equal(t, "Test ProductService", repo.store[productID].Name)
		require.Equal(t, 1, repo.store[productID].Quantity)
		require.Equal(t, 24.9, repo.store[productID].Price)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model2.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
	})
	eventDispatcher.Reset()

	// nolint:dupl
	t.Run("Increase product quantity", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		err = productService.IncreaseQuantity(productID, 10)
		require.NoError(t, err)

		require.NotNil(t, repo.store[productID])
		require.Equal(t, "Test ProductService", repo.store[productID].Name)
		require.Equal(t, 11, repo.store[productID].Quantity)
		require.Equal(t, 24.9, repo.store[productID].Price)
		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model2.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model2.ProductQuantityChanged{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Increase non existed product quantity", func(t *testing.T) {
		err := productService.IncreaseQuantity(uuid.New(), 10)
		require.ErrorIs(t, err, model2.ErrProductNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})
	eventDispatcher.Reset()

	// nolint:dupl
	t.Run("Decrease product quantity", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		err = productService.DecreaseQuantity(productID, 1)
		require.NoError(t, err)

		require.NotNil(t, repo.store[productID])
		require.Equal(t, "Test ProductService", repo.store[productID].Name)
		require.Equal(t, 0, repo.store[productID].Quantity)
		require.Equal(t, 24.9, repo.store[productID].Price)
		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model2.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model2.ProductQuantityChanged{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Decrease product quantity to less than zero", func(t *testing.T) {
		productID, err := productService.CreateProduct(name, quantity, price)
		require.NoError(t, err)

		err = productService.DecreaseQuantity(productID, 2)
		require.ErrorIs(t, err, model2.ErrProductQuantityLessThanZero)

		require.NotNil(t, repo.store[productID])
		require.Equal(t, "Test ProductService", repo.store[productID].Name)
		require.Equal(t, 1, repo.store[productID].Quantity)
		require.Equal(t, 24.9, repo.store[productID].Price)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model2.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
	})
	eventDispatcher.Reset()

	t.Run("Decrease non existed product quantity", func(t *testing.T) {
		err := productService.DecreaseQuantity(uuid.New(), 10)
		require.ErrorIs(t, err, model2.ErrProductNotFound)

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
		require.Equal(t, model2.ProductCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model2.ProductDeleted{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Delete non existed product", func(t *testing.T) {
		newID, _ := repo.NextID()
		err := productService.DeleteProduct(newID)
		require.ErrorIs(t, err, model2.ErrProductNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})
	eventDispatcher.Reset()
}

var _ model2.ProductRepository = &mockProductRepository{}

type mockProductRepository struct {
	store map[uuid.UUID]*model2.Product
}

func (m *mockProductRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockProductRepository) Store(product *model2.Product) error {
	m.store[product.ID] = product
	return nil
}

func (m *mockProductRepository) Find(id uuid.UUID) (*model2.Product, error) {
	product, ok := m.store[id]
	if !ok {
		return nil, model2.ErrProductNotFound
	}
	if product.DeletedAt != nil {
		return nil, model2.ErrProductNotFound
	}
	return product, nil
}

func (m *mockProductRepository) List() ([]model2.Product, error) {
	var res []model2.Product
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
		return model2.ErrProductNotFound
	}
	now := time.Now()
	product.DeletedAt = &now
	return nil
}

type mockEventDispatcher struct {
	events []domain.Event
}

func (m *mockEventDispatcher) Reset() {
	m.events = make([]domain.Event, 0)
}

func (m *mockEventDispatcher) ListEvents() []domain.Event {
	return m.events
}

func (m *mockEventDispatcher) Dispatch(evt domain.Event) error {
	m.events = append(m.events, evt)
	return nil
}
