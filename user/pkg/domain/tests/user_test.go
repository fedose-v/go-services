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
		store: make(map[uuid.UUID]*model.User),
	}
	eventDispatcher := &mockEventDispatcher{
		events: make([]event.Event, 0),
	}

	userService := service.NewUserService(repo, eventDispatcher)

	login := "victor.wembanyama"
	name := "Victor Wembanyama"
	email := "victor.wembanyama@example.com"

	t.Run("Create user", func(t *testing.T) {
		userID, err := userService.CreateUser(login, name, email)
		require.NoError(t, err)

		require.NotNil(t, repo.store[userID])
		require.Equal(t, "victor.wembanyama", repo.store[userID].Login)
		require.Equal(t, "Victor Wembanyama", repo.store[userID].Name)
		require.Equal(t, "victor.wembanyama@example.com", repo.store[userID].Email)
		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.UserCreated{}.Type(), eventDispatcher.events[0].Type())
	})
	eventDispatcher.Reset()

	t.Run("Update user", func(t *testing.T) {
		userID, err := userService.CreateUser(login, name, email)
		require.NoError(t, err)

		err = userService.UpdateUser(userID, "popular.victor.wembanyama", "Popular Victor Wembanyama", "popular.victor.wembanyama@example.com")
		require.NoError(t, err)

		require.NotNil(t, repo.store[userID])
		require.Equal(t, "popular.victor.wembanyama", repo.store[userID].Login)
		require.Equal(t, "Popular Victor Wembanyama", repo.store[userID].Name)
		require.Equal(t, "popular.victor.wembanyama@example.com", repo.store[userID].Email)

		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model.UserCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.UserUpdated{}.Type(), eventDispatcher.events[1].Type())
	})

	t.Run("Update non existed user", func(t *testing.T) {
		userID := uuid.New()
		err := userService.UpdateUser(userID, "popular.victor.wembanyama", "Popular Victor Wembanyama", "popular.victor.wembanyama@example.com")
		require.ErrorIs(t, err, model.ErrUserNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})

	t.Run("Delete user", func(t *testing.T) {
		userID, err := userService.CreateUser(login, name, email)
		require.NoError(t, err)

		err = userService.DeleteUser(userID)
		require.NoError(t, err)

		require.NotNil(t, repo.store[userID])
		require.NotNil(t, repo.store[userID].DeletedAt)
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

func (m *mockUserRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockUserRepository) Store(user *model.User) error {
	m.store[user.ID] = user
	return nil
}

func (m *mockUserRepository) Find(id uuid.UUID) (*model.User, error) {
	user, ok := m.store[id]
	if !ok {
		return nil, model.ErrUserNotFound
	}
	if user.DeletedAt != nil {
		return nil, model.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepository) List() ([]model.User, error) {
	res := make([]model.User, 0, len(m.store))
	for _, user := range m.store {
		if user != nil && user.DeletedAt == nil {
			res = append(res, *user)
		}
	}
	return res, nil
}

func (m *mockUserRepository) Delete(id uuid.UUID) error {
	user, ok := m.store[id]
	if !ok {
		return model.ErrUserNotFound
	}
	now := time.Now()
	user.DeletedAt = &now
	return nil
}

type mockEventDispatcher struct {
	events []event.Event
}

func (m *mockEventDispatcher) Reset() {
	m.events = make([]event.Event, 0)
}

func (m *mockEventDispatcher) Dispatch(evt event.Event) error {
	m.events = append(m.events, evt)
	return nil
}
