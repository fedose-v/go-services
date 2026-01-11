package mysql

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"inventory/pkg/inventory/app/service"
	"inventory/pkg/inventory/infrastructure/mysql/repository"

	"inventory/pkg/inventory/domain/model"
)

func NewRepositoryProvider(client mysql.ClientContext) service.RepositoryProvider {
	return &repositoryProvider{client: client}
}

type repositoryProvider struct {
	client mysql.ClientContext
}

func (r *repositoryProvider) ProductRepository(ctx context.Context) model.ProductRepository {
	return repository.NewProductRepository(ctx, r.client)
}
