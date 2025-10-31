package queryservice

import (
	"database/sql"
	"fmt"
	"inventory/pkg/app/data"
	"inventory/pkg/app/service"
)

func NewProductQueryService(db *sql.DB) service.ProductQueryService {
	return queryService{db: db}
}

type queryService struct {
	db *sql.DB
}

func (q queryService) ListProducts() ([]data.ProductData, error) {
	rows, err := q.db.Query(`
		SELECT 
			product_id, 
			name, 
			price, 
			quantity, 
			created_at
		FROM product 
		WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []data.ProductData

	for rows.Next() {
		var product data.ProductData
		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Quantity,
			&product.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}
