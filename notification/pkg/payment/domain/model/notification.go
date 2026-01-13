package model

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

type Recipient struct {
	Name  string
	Email string
}

type Notification struct {
	ID      uuid.UUID
	Name    string
	Subject string
	Body    string
}

type NotificationRepository interface {
	NextID() (uuid.UUID, error)
	Store(notification *Notification) error
	Find(id uuid.UUID) (*Notification, error)
}
