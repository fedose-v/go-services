package app

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"user/pkg/domain/model"
)

type RepositoryProvider interface {
	UserRepository(ctx context.Context) model.UserRepository
}

type unitOfWork struct {
	uow mysql.LockableUnitOfWorkWithRepositoryProvider[RepositoryProvider]
}
