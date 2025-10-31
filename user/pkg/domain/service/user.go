package service

import (
	"github.com/google/uuid"
	"time"

	"user/pkg/common/infrastructure/event"
	"user/pkg/domain/model"
)

type User interface {
	CreateUser(login string, name string, email string) (uuid.UUID, error)
	DeleteUser(userID uuid.UUID) error
}

func NewUserService(repo model.UserRepository, dispatcher event.Dispatcher) User {
	return &userService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type userService struct {
	repo       model.UserRepository
	dispatcher event.Dispatcher
}

func (o userService) CreateUser(login string, name string, email string) (uuid.UUID, error) {
	userID, err := o.repo.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	currentTime := time.Now()
	err = o.repo.Store(&model.User{
		ID:        userID,
		Login:     login,
		Name:      name,
		Email:     email,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return userID, o.dispatcher.Dispatch(model.UserCreated{
		ID:        userID,
		Login:     login,
		Name:      name,
		Email:     email,
		CreatedAt: currentTime,
	})
}

func (o userService) DeleteUser(userID uuid.UUID) error {
	err := o.repo.Delete(userID)
	if err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.UserDeleted{
		ID: userID,
	})
}
