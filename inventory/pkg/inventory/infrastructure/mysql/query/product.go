package queryservice

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	appmodel "inventory/pkg/inventory/app/model"
	"inventory/pkg/inventory/app/query"
	"inventory/pkg/inventory/domain/model"
)

func NewProductQueryService(client mysql.ClientContext) query.ProductQueryService {
	return &productQueryService{
		client: client,
	}
}

type productQueryService struct {
	client mysql.ClientContext
}

func (p *productQueryService) ListProducts(ctx context.Context) ([]appmodel.Product, error) {
	var products []appmodel.Product

	err := p.client.SelectContext(
		ctx,
		&products,
		`SELECT product_id, name, price, quantity, created_at 
		 FROM product 
		 WHERE deleted_at IS NULL`,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return products, nil
}

func (p *productQueryService) FindProduct(ctx context.Context, id uuid.UUID) (*appmodel.Product, error) {
	row := struct {
		ID       uuid.UUID `db:"product_id"`
		Name     string    `db:"name"`
		Price    float64   `db:"price"`
		Quantity int64     `db:"quantity"`
	}{}

	err := p.client.GetContext(
		ctx,
		&row,
		`SELECT product_id, name, price, quantity, created_at 
		 FROM product 
		 WHERE product_id = ? AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrProductNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &appmodel.Product{
		ID:       row.ID,
		Name:     row.Name,
		Price:    row.Price,
		Quantity: int(row.Quantity),
	}, nil
}
