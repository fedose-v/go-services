package service

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	"notification/pkg/common/domain"
	"notification/pkg/domain/service"
)

type NotificationService interface {
	CreateNotification(ctx context.Context, name, subject, body string) (uuid.UUID, error)
}

func NewNotificationService(uow UnitOfWork, eventDispatcher outbox.EventDispatcher[outbox.Event]) NotificationService {
	return &notificationService{uow: uow, eventDispatcher: eventDispatcher}
}

type notificationService struct {
	uow             UnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (n *notificationService) CreateNotification(ctx context.Context, name, subject, body string) (uuid.UUID, error) {
	var notificationID uuid.UUID
	err := n.uow.Execute(ctx, func(provider RepositoryProvider) error {
		domainService := service.NewNotificationService(provider.NotificationRepository(ctx), n.domainEventDispatcher(ctx))
		id, err := domainService.CreateNotification(name, subject, body)
		if err != nil {
			return err
		}
		notificationID = id
		return nil
	})
	return notificationID, err
}

func (n *notificationService) domainEventDispatcher(ctx context.Context) domain.EventDispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: n.eventDispatcher,
	}
}
