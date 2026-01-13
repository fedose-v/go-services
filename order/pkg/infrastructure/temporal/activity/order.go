package activity

import (
	"order/pkg/app/service"
)

func NewOrderServiceActivities(orderService service.OrderService) *OrderServiceActivities {
	return &OrderServiceActivities{
		orderService: orderService,
	}
}

type OrderServiceActivities struct {
	orderService service.OrderService
}
