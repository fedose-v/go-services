package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion3(client mysql.ClientContext) migrator.Migration {
	return &version3{
		client: client,
	}
}

type version3 struct {
	client mysql.ClientContext
}

func (v version3) Version() int64 {
	return 3
}

func (v version3) Description() string {
	return "Create 'local_product' table"
}

func (v version3) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE local_product (
		    product_id VARCHAR(64) NOT NULL,
		    name VARCHAR(255) NOT NULL,
		    price BIGINT NOT NULL,
		    PRIMARY KEY (product_id)
		)
		    ENGINE = InnoDB
		    CHARACTER SET = utf8mb4
		    COLLATE utf8mb4_unicode_ci;
	`)
	return errors.WithStack(err)
}
