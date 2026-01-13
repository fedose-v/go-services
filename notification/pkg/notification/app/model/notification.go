package model

import (
	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID
	Name      string
	Subject   string
	Body      string
	CreatedAt int64
}
