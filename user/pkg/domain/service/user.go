package service

import (
	"time"

	"github.com/google/uuid"

	"user/pkg/common/infrastructure/event"
	"user/pkg/domain/model"
)

type User interface {
	CreateUser(login string, name string, email string) (uuid.UUID, error)
	UpdateUser(userID uuid.UUID, login string, name string, email string) error
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

func (o userService) UpdateUser(userID uuid.UUID, login string, name string, email string) error {
	user, err := o.repo.Find(userID)
	if err != nil {
		return err
	}

	if len(login) > 0 {
		user.Login = login
	}
	if len(name) > 0 {
		user.Name = name
	}
	if len(email) > 0 {
		user.Email = email
	}

	err = o.repo.Store(user)
	if err != nil {
		return err
	}

	return o.dispatcher.Dispatch(model.UserUpdated{
		ID:    userID,
		Login: user.Login,
		Name:  user.Name,
		Email: user.Email,
	})
}
