package activity

import (
	"context"

	"github.com/google/uuid"

	appmodel "inventory/pkg/inventory/app/model"
	"inventory/pkg/inventory/app/service"
)

func NewInventoryServiceActivities(userService service.ProductService) *UserServiceActivities {
	return &UserServiceActivities{userService: userService}
}

type UserServiceActivities struct {
	userService service.ProductService
}

func (a *UserServiceActivities) FindProduct(ctx context.Context, userID uuid.UUID) (appmodel.Product, error) {
	return a.userService.FindProduct(ctx, userID)
}

func (a *UserServiceActivities) SetUserStatus() error {
	return nil
}
