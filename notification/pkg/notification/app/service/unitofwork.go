package service

import (
	"context"

	"notification/pkg/notification/domain/model"
)

type RepositoryProvider interface {
	NotificationRepository(ctx context.Context) model.NotificationRepository
}

type UnitOfWork interface {
	Execute(ctx context.Context, f func(provider RepositoryProvider) error) error
}
