package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion1(client mysql.ClientContext) migrator.Migration {
	return &version1{
		client: client,
	}
}

type version1 struct {
	client mysql.ClientContext
}

func (v version1) Version() int64 {
	return 1
}

func (v version1) Description() string {
	return "Create 'customer_account_balance' table"
}

func (v version1) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE customer_account_balance
		(
			id          BINARY(16) 		PRIMARY KEY,
			customer_id BINARY(16)     	NOT NULL,
			amount      DECIMAL(15, 2) 	NOT NULL DEFAULT 0.00,
			created_at  DATETIME       	NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME       	NULL ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_customer_id (customer_id)
		) ENGINE = InnoDB
		  DEFAULT CHARSET = utf8mb4
		  COLLATE = utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
