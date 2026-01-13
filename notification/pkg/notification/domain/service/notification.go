package service

import (
	"github.com/google/uuid"

	"notification/pkg/notification/domain/model"
)

type Notification interface {
	CreateNotification(name string, subject string, body string) (uuid.UUID, error)
}

func NewNotificationService(repo model.NotificationRepository) Notification {
	return &notificationService{repo}
}

type notificationService struct {
	repo model.NotificationRepository
}

func (n notificationService) CreateNotification(name, subject, body string) (uuid.UUID, error) {
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
	return id, err
}
