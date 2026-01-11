package database

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func NewVersion1722266003(client mysql.ClientContext) migrator.Migration {
	return &version1722266003{
		client: client,
	}
}

type version1722266003 struct {
	client mysql.ClientContext
}

func (v version1722266003) Version() int64 {
	return 1722266003
}

func (v version1722266003) Description() string {
	return "Create 'user' table"
}

func (v version1722266003) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, `
		CREATE TABLE user
		(
			id         BINARY(16)   NOT NULL PRIMARY KEY,
			login      VARCHAR(100) NOT NULL UNIQUE,
			name       VARCHAR(255) NOT NULL,
			email      VARCHAR(255) NOT NULL UNIQUE,
			created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME     NULL
		) ENGINE = InnoDB
		  DEFAULT CHARSET = utf8mb4
		  COLLATE = utf8mb4_unicode_ci
	`)
	return errors.WithStack(err)
}
