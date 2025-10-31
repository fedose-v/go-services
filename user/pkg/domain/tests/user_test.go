package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"user/pkg/common/infrastructure/event"
	"user/pkg/domain/model"
	"user/pkg/domain/service"
)

func TestUserService(t *testing.T) {
	repo := &mockUserRepository{
		store: map[uuid.UUID]*model.User{},
	}
	eventDispatcher := &mockEventDispatcher{}

	userService := service.NewUserService(repo, eventDispatcher)

	login := "victor.wembanyama"
	name := "Victor Wembanyama"
	email := "victor.wembanyama@example.com"
	t.Run("Create user", func(t *testing.T) {
		userID, err := userService.CreateUser(login, name, email)
		require.NoError(t, err)

		require.NotNil(t, repo.store[userID])
		require.Equal(t, repo.store[userID].Login, "victor.wembanyama")
		require.Equal(t, repo.store[userID].Name, "Victor Wembanyama")
		require.Equal(t, repo.store[userID].Email, "victor.wembanyama@example.com")
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.UserCreated{}.Type(), eventDispatcher.events[0].Type())
	})
	eventDispatcher.Reset()

	t.Run("Delete user", func(t *testing.T) {
		userID, err := userService.CreateUser(login, name, email)
		require.NoError(t, err)

		err = userService.DeleteUser(userID)
		require.NoError(t, err)

		require.Nil(t, repo.store[userID])
		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model.UserCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.UserDeleted{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Delete non existed user", func(t *testing.T) {
		newID, _ := repo.NextID()
		err := userService.DeleteUser(newID)
		require.ErrorIs(t, err, model.ErrUserNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})
	eventDispatcher.Reset()
}

var _ model.UserRepository = &mockUserRepository{}

type mockUserRepository struct {
	store map[uuid.UUID]*model.User
}

func (m mockUserRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m mockUserRepository) Store(user *model.User) error {
	m.store[user.ID] = user
	return nil
}

func (m mockUserRepository) Find(id uuid.UUID) (*model.User, error) {
	if user, ok := m.store[id]; ok && user.DeletedAt != nil {
		return user, nil
	}
	return nil, model.ErrUserNotFound
}

func (m mockUserRepository) List() (*[]model.User, error) {
	var res []model.User
	for _, v := range m.store {
		if v != nil {
			res = append(res, *v)
		}
	}
	return &res, nil
}

func (m mockUserRepository) Delete(id uuid.UUID) error {
	if user, ok := m.store[id]; ok && user.DeletedAt != nil {
		user.DeletedAt = toPtr(time.Now())
		return nil
	}
	return model.ErrUserNotFound
}

type MockEventDispatcher interface {
	event.Dispatcher
	Reset()
}

var _ MockEventDispatcher = &mockEventDispatcher{}

type mockEventDispatcher struct {
	events []event.Event
}

func (m mockEventDispatcher) Reset() {
	m.events = nil
}

func (m mockEventDispatcher) Dispatch(event event.Event) error {
	m.events = append(m.events, event)
	return nil
}

func toPtr[V any](v V) *V {
	return &v
}
