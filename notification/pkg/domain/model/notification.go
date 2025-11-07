package model

import (
	"errors"
	"time"

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
	ID        uuid.UUID
	Name      string
	Subject   string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type NotificationRepository interface {
	NextID() (uuid.UUID, error)
	Store(notification *Notification) error
	Find(id uuid.UUID) (*Notification, error)
	Delete(id uuid.UUID) error
}
