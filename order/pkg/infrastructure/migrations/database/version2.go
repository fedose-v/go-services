package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion2(client mysql.ClientContext) migrator.Migration {
	return &version2{
		client: client,
	}
}

type version2 struct {
	client mysql.ClientContext
}

func (v version2) Version() int64 {
	return 1
}

func (v version2) Description() string {
	return "Create 'order_item' table"
}

func (v version2) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS order_item
		(
		    id         BINARY(16) PRIMARY KEY,
		    order_id   BINARY(16)     NOT NULL,
		    product_id BINARY(16)     NOT NULL,
		    price      DECIMAL(10, 2) NOT NULL,
		    created_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    deleted_at DATETIME       NULL,
		    FOREIGN KEY (order_id) REFERENCES order (id) ON DELETE CASCADE
		) ENGINE = InnoDB
		  DEFAULT CHARSET = utf8mb4
		  COLLATE = utf8mb4_unicode_ci;
	`)
	return errors.WithStack(err)
}
