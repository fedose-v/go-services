package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationCreated struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

func (e NotificationCreated) Type() string {
	return "NotificationCreated"
}

type NotificationDeleted struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

func (e NotificationDeleted) Type() string {
	return "NotificationDeleted"
}

type NotificationSent struct {
	ID             uuid.UUID
	RecipientName  string
	RecipientEmail string
	SentAt         time.Time
}

func (e NotificationSent) Type() string {
	return "NotificationSent"
}
