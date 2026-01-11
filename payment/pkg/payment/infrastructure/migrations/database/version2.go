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
	return 2
}

func (v version2) Description() string {
	return "Create 'transaction' table"
}

func (v version2) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE transaction
		(
			id           BINARY(16)     NOT NULL PRIMARY KEY,
			order_id     BINARY(16)     NOT NULL,
			customer_id  BINARY(16)     NOT NULL,
			type         TINYINT        NOT NULL DEFAULT 0 COMMENT '0: New, 1: Refund',
			amount       DECIMAL(10, 2) NOT NULL,
			payment_date DATETIME       NOT NULL,
			created_at   DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE = InnoDB
		  DEFAULT CHARSET = utf8mb4
		  COLLATE = utf8mb4_unicode_ci;
	`)

	return errors.WithStack(err)
}
