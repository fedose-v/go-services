package transport

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	appmodel "inventory/pkg/inventory/app/model"
	"inventory/pkg/inventory/app/query"
	"inventory/pkg/inventory/app/service"

	"inventory/api/server/inventorypublicapi"
)

func NewInventoryInternalAPI(
	inventoryQueryService query.ProductQueryService,
	inventoryService service.ProductService,
) inventorypublicapi.InventoryPublicAPIServer {
	return &inventoryInternalAPI{
		inventoryQueryService: inventoryQueryService,
		inventoryService:      inventoryService,
	}
}

type inventoryInternalAPI struct {
	inventoryQueryService query.ProductQueryService
	inventoryService      service.ProductService

	inventorypublicapi.UnimplementedInventoryPublicAPIServer
}

func (u inventoryInternalAPI) StoreInventory(ctx context.Context, request *inventorypublicapi.StoreProductRequest) (*inventorypublicapi.StoreProductResponse, error) {
	var (
		productID uuid.UUID
		err       error
	)
	if request.ProductID != "" {
		productID, err = uuid.Parse(request.ProductID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.ProductID)
		}
	}

	productID, err = u.inventoryService.StoreProduct(ctx, appmodel.Product{
		ID:       productID,
		Name:     request.Name,
		Price:    request.Price,
		Quantity: int(request.Quantity),
	})
	if err != nil {
		return nil, err
	}

	return &inventorypublicapi.StoreProductResponse{
		ProductID: productID.String(),
	}, nil
}

func (u inventoryInternalAPI) FindInventory(ctx context.Context, request *inventorypublicapi.FindProductRequest) (*inventorypublicapi.FindProductResponse, error) {
	productID, err := uuid.Parse(request.ProductID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.ProductID)
	}
	product, err := u.inventoryQueryService.FindProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, status.Errorf(codes.NotFound, "product %q not found", request.ProductID)
	}
	return &inventorypublicapi.FindProductResponse{
		ProductID: productID.String(),
		Name:      product.Name,
		Price:     product.Price,
		Quantity:  int64(product.Quantity),
	}, nil
}
