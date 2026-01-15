package consumer

import (
	"context"
	"time"

	"payment/pkg/payment/app/service"
	inframysql "payment/pkg/payment/infrastructure/mysql"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
)

type lockableUnitOfWorkForSync struct {
	uow mysql.LockableUnitOfWorkWithRepositoryProvider[service.RepositoryProvider]
}

func (l *lockableUnitOfWorkForSync) Execute(ctx context.Context, lockNames []string, f func(provider service.RepositoryProvider) error) error {
	if len(lockNames) == 1 {
		return l.uow.ExecuteWithRepositoryProvider(ctx, lockNames[0], time.Minute, f)
	}
	ln := lockNames[0]
	lns := lockNames[1:]
	return l.uow.ExecuteWithRepositoryProvider(ctx, ln, time.Minute, func(_ service.RepositoryProvider) error {
		return l.Execute(ctx, lns, f)
	})
}

func NewLockableUnitOfWork(pool mysql.ConnectionPool) service.LockableUnitOfWork {
	return &lockableUnitOfWorkForSync{
		mysql.NewLockableUnitOfWork(mysql.NewUnitOfWork(pool, inframysql.NewRepositoryProvider), mysql.NewLocker(pool)),
	}
}
