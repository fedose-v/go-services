package service

import (
	"context"

	"github.com/google/uuid"

	"notification/pkg/notification/domain/service"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, name, subject, body string) (uuid.UUID, error)
}

func NewNotificationService(uow UnitOfWork) NotificationService {
	return &notificationService{uow: uow}
}

type notificationService struct {
	uow UnitOfWork
}

func (n *notificationService) CreateNotification(ctx context.Context, name, subject, body string) (uuid.UUID, error) {
	var notificationID uuid.UUID
	err := n.uow.Execute(ctx, func(provider RepositoryProvider) error {
		domainService := service.NewNotificationService(provider.NotificationRepository(ctx))
		id, err := domainService.CreateNotification(name, subject, body)
		if err != nil {
			return err
		}
		notificationID = id
		return nil
	})
	return notificationID, err
}
