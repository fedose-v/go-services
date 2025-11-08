package model

import (
	"time"

	"github.com/google/uuid"
)

type UserCreated struct {
	ID        uuid.UUID
	Login     string
	Name      string
	Email     string
	CreatedAt time.Time
}

func (e UserCreated) Type() string {
	return "UserCreated"
}

type UserUpdated struct {
	ID    uuid.UUID
	Login string
	Name  string
	Email string
}

func (e UserUpdated) Type() string {
	return "UserUpdated"
}

type UserDeleted struct {
	ID uuid.UUID
}

func (e UserDeleted) Type() string {
	return "UserDeleted"
}
