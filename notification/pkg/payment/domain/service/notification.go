package service

import (
	"github.com/google/uuid"
	"order/pkg/common/event"

	"time"

	"notification/pkg/common/domain"
	"notification/pkg/domain/model"
)

type Notification interface {
	CreateNotification(name string, subject string, body string) (uuid.UUID, error)

	SendNotification(id uuid.UUID, recipient model.Recipient) error
}

func NewNotificationService(repo model.NotificationRepository, d domain.EventDispatcher) Notification {
	return &notificationService{repo, d}
}

type notificationService struct {
	repo model.NotificationRepository
	d    domain.EventDispatcher
}

func (n notificationService) CreateNotification(name string, subject string, body string) (uuid.UUID, error) {
	id, err := n.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	err = n.repo.Store(&model.Notification{
		ID:      id,
		Name:    name,
		Subject: subject,
		Body:    body,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return id, n.d.Dispatch(model.NotificationCreated{
		ID:        id,
		CreatedAt: time.Now(),
	})
}

func (n notificationService) SendNotification(ID uuid.UUID, recipient model.Recipient) error {
	notification, err := n.repo.Find(ID)
	if err != nil {
		return err
	}

	return n.d.Dispatch(model.NotificationSent{
		ID:             notification.ID,
		RecipientName:  recipient.Name,
		RecipientEmail: recipient.Email,
		SentAt:         time.Now(),
	})
}
