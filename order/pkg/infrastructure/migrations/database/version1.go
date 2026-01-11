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
	return "Create 'order' table"
}

func (v version1) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS `+"`order`"+`
		(
			id          BINARY(16) NOT NULL PRIMARY KEY,
			customer_id BINARY(16) NOT NULL,
			status      TINYINT    NOT NULL DEFAULT 0 COMMENT '0: Open, 1: Pending, 2: Paid, 3: Cancelled, 4: Deleted',
			created_at  DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at  DATETIME   NULL
		) ENGINE = InnoDB
		  DEFAULT CHARSET = utf8mb4
		  COLLATE = utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
