package repository

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"inventory/pkg/inventory/domain/model"
)

func NewProductRepository(ctx context.Context, client mysql.ClientContext) model.ProductRepository {
	return &productRepository{
		ctx:    ctx,
		client: client,
	}
}

type productRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (p *productRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (p *productRepository) Store(product *model.Product) error {
	if product.Quantity < 0 {
		return errors.WithStack(model.ErrProductQuantityLessThanZero)
	}

	_, err := p.client.ExecContext(p.ctx,
		`
		INSERT INTO product (id, name, price, quantity, created_at, updated_at, deleted_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name=VALUES(name),
			price=VALUES(price),
			quantity=VALUES(quantity),
			updated_at=VALUES(updated_at),
			deleted_at=VALUES(deleted_at)
		`,
		product.ID,
		product.Name,
		product.Price,
		product.Quantity,
		product.CreatedAt,
		product.UpdatedAt,
		toSQLNullTime(product.DeletedAt),
	)
	return errors.WithStack(err)
}

func (p *productRepository) Find(id uuid.UUID) (*model.Product, error) {
	row := struct {
		ID        uuid.UUID  `db:"id"`
		Name      string     `db:"name"`
		Price     float64    `db:"price"`
		Quantity  int        `db:"quantity"`
		CreatedAt time.Time  `db:"created_at"`
		UpdatedAt time.Time  `db:"updated_at"`
		DeletedAt *time.Time `db:"deleted_at"`
	}{}

	err := p.client.GetContext(
		p.ctx,
		&row,
		`SELECT id, name, price, quantity, created_at, updated_at, deleted_at FROM product WHERE id = ? AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrProductNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &model.Product{
		ID:        row.ID,
		Name:      row.Name,
		Price:     row.Price,
		Quantity:  row.Quantity,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: row.DeletedAt,
	}, nil
}

func (p *productRepository) List() ([]model.Product, error) {
	var rows []struct {
		ID        uuid.UUID  `db:"id"`
		Name      string     `db:"name"`
		Price     float64    `db:"price"`
		Quantity  int        `db:"quantity"`
		CreatedAt time.Time  `db:"created_at"`
		UpdatedAt time.Time  `db:"updated_at"`
		DeletedAt *time.Time `db:"deleted_at"`
	}

	err := p.client.SelectContext(
		p.ctx,
		&rows,
		`SELECT id, name, price, quantity, created_at, updated_at, deleted_at FROM product WHERE deleted_at IS NULL ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	products := make([]model.Product, len(rows))
	for i, r := range rows {
		products[i] = model.Product{
			ID:        r.ID,
			Name:      r.Name,
			Price:     r.Price,
			Quantity:  r.Quantity,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			DeletedAt: r.DeletedAt,
		}
	}

	return products, nil
}

func (p *productRepository) Delete(id uuid.UUID) error {
	now := time.Now()
	_, err := p.client.ExecContext(p.ctx,
		`UPDATE product SET deleted_at = ? WHERE id = ?`,
		now,
		id,
	)
	return errors.WithStack(err)
}

func toSQLNullTime(t *time.Time) sql.Null[time.Time] {
	if t == nil {
		return sql.Null[time.Time]{}
	}
	return sql.Null[time.Time]{
		V:     *t,
		Valid: true,
	}
}
