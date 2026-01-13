package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"notification/pkg/app/service"
	"notification/pkg/domain/model"
	"notification/pkg/infrastructure/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) NotificationRepository(ctx context.Context) model.NotificationRepository {
	return repository.NewNotificationRepository(ctx, r.client)
}
