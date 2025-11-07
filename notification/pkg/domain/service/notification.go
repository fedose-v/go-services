package service

import (
	"github.com/google/uuid"
	"notification/pkg/common/infrastructure/event"
	"notification/pkg/domain/model"
	"time"
)

type Notification interface {
	CreateNotification(name string, subject string, body string) (uuid.UUID, error)
	DeleteNotification(id uuid.UUID) error

	SendNotification(id uuid.UUID, recipient model.Recipient) error
}

func NewNotificationService(repo model.NotificationRepository, d event.Dispatcher) Notification {
	return &notificationService{repo, d}
}

type notificationService struct {
	repo model.NotificationRepository
	d    event.Dispatcher
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

func (n notificationService) DeleteNotification(id uuid.UUID) error {
	_, err := n.repo.Find(id)
	if err != nil {
		return err
	}

	err = n.repo.Delete(id)
	if err != nil {
		return err
	}

	return n.d.Dispatch(model.NotificationDeleted{ID: id})
}
