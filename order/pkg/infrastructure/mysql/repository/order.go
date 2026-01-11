package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"

	"order/pkg/domain/model"
)

func NewOrderRepository(ctx context.Context, client mysql.ClientContext) model.OrderRepository {
	return &orderRepository{
		ctx:    ctx,
		client: client,
	}
}

type orderRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (o orderRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (o orderRepository) Store(order *model.Order) error {
	err := o.storeOrder(order)
	if err != nil {
		return err
	}

	err = o.storeItems(order.ID, order.Items)
	if err != nil {
		return err
	}

	return nil
}

func (o orderRepository) storeOrder(order *model.Order) error {
	const storeOrder = `
		INSERT INTO order (
			id,
		    customer_id,
		    status,
		    created_at,
		    updated_at,
		    deleted_at
		) VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			status=VALUES(status),
			updated_at=VALUES(updated_at),
			deleted_at=VALUES(deleted_at)
	`

	_, err := o.client.ExecContext(
		o.ctx, storeOrder,
		order.ID, order.CustomerID, order.Status, order.CreatedAt, order.UpdatedAt, order.DeletedAt,
	)
	return err
}

func (o orderRepository) storeItems(orderID uuid.UUID, items []model.Item) error {
	const deleteItems = `DELETE FROM order_item WHERE order_id = ?`
	_, err := o.client.ExecContext(o.ctx, deleteItems, orderID)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return nil
	}

	const storeItems = `INSERT INTO order_item (id, order_id, product_id, price) VALUES (?, ?, ?, ?)`
	for _, item := range items {
		_, err = o.client.ExecContext(o.ctx, storeItems, item.ID, orderID, item.ProductID, item.Price)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o orderRepository) Find(spec model.FindSpec) (*model.Order, error) {
	order, err := o.findOrder(spec)
	if err != nil {
		return nil, err
	}

	items, err := o.findItems(spec)
	if err != nil {
		return nil, err
	}

	var domainItems []model.Item
	for _, item := range items {
		domainItems = append(domainItems, model.Item{
			ID:        item.ID,
			ProductID: item.ProductID,
			Price:     item.Price,
		})
	}

	return &model.Order{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Status:     model.OrderStatus(order.Status),
		Items:      domainItems,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
		DeletedAt:  order.DeletedAt,
	}, nil
}

func (o orderRepository) Delete(id uuid.UUID) error {
	spec := model.FindSpec{
		OrderID:        &id,
		IncludeDeleted: false,
	}
	err := o.deleteItems(spec)
	if err != nil {
		return err
	}
	return o.deleteOrder(spec)
}

func (o orderRepository) findOrder(spec model.FindSpec) (*sqlxOrder, error) {
	const findOrder = `
		SELECT
			id,
			customer_id,
			status,
			created_at,
			updated_at,
			deleted_at
		FROM order
	`
	whereQuery, args := o.buildWhereConditions(spec)

	var order sqlxOrder
	err := o.client.GetContext(o.ctx, &order, findOrder+whereQuery, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrOrderNotFound
		}
		return nil, err
	}

	return &order, nil
}

func (o orderRepository) findItems(spec model.FindSpec) ([]sqlxItem, error) {
	const findOrderItems = `
		SELECT
			id,
			product_id,
			price
		FROM order_item
	`
	whereQuery, args := o.buildItemWhereConditions(spec)

	var items []sqlxItem
	err := o.client.SelectContext(o.ctx, &items, findOrderItems+whereQuery, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrOrderNotFound
		}
		return nil, err
	}

	return items, nil
}

func (o orderRepository) deleteOrder(spec model.FindSpec) error {
	const deleteOrder = `
		DELETE FROM order
	`
	whereQuery, args := o.buildWhereConditions(spec)

	_, err := o.client.ExecContext(o.ctx, deleteOrder+whereQuery, args...)
	return err
}

func (o orderRepository) deleteItems(spec model.FindSpec) error {
	const deleteOrderItems = `
		DELETE FROM order_item
	`
	whereQuery, args := o.buildItemWhereConditions(spec)

	_, err := o.client.ExecContext(o.ctx, deleteOrderItems+whereQuery, args...)
	return err
}

func (o orderRepository) buildWhereConditions(spec model.FindSpec) (query string, args []interface{}) {
	var parts []string
	if spec.OrderID != nil {
		parts = append(parts, "order_id = ?")
		args = append(args, *spec.OrderID)
	}
	if spec.CustomerID != nil {
		parts = append(parts, "customer_id = ?")
		args = append(args, *spec.CustomerID)
	}
	if spec.Status != nil {
		parts = append(parts, "status = ?")
		args = append(args, *spec.Status)
	}
	if !spec.IncludeDeleted {
		parts = append(parts, "deleted_at IS NULL")
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

func (o orderRepository) buildItemWhereConditions(spec model.FindSpec) (query string, args []interface{}) {
	var parts []string
	if spec.OrderID != nil {
		parts = append(parts, "order_id = ?")
		args = append(args, *spec.OrderID)
	}
	if !spec.IncludeDeleted {
		parts = append(parts, "deleted_at IS NULL")
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

type sqlxOrder struct {
	ID         uuid.UUID  `db:"id"`
	CustomerID uuid.UUID  `db:"customer_id"`
	Status     int        `db:"status"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

type sqlxItem struct {
	ID        uuid.UUID `db:"id"`
	ProductID uuid.UUID `db:"product_id"`
	Price     float64   `db:"price"`
}
