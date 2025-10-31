package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Login     string
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	NextID() (uuid.UUID, error)
	Store(user *User) error
	Find(id uuid.UUID) (*User, error)
	Delete(id uuid.UUID) error
}
