package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"notification/pkg/common/infrastructure/event"
	"notification/pkg/domain/model"
	"notification/pkg/domain/service"
)

func TestNotificationService(t *testing.T) {
	repo := &mockNotificationRepository{
		store: make(map[uuid.UUID]*model.Notification),
	}
	eventDispatcher := &mockEventDispatcher{
		events: make([]event.Event, 0),
	}

	notificationService := service.NewNotificationService(repo, eventDispatcher)

	name := "Test Notification"
	subject := "Something went wrong"
	body := "Something went wrong. Please contact support."

	t.Run("Create notification", func(t *testing.T) {
		notificationID, err := notificationService.CreateNotification(name, subject, body)
		require.NoError(t, err)

		require.NotNil(t, repo.store[notificationID])
		require.Equal(t, "Test Notification", repo.store[notificationID].Name)
		require.Equal(t, "Something went wrong", repo.store[notificationID].Subject)
		require.Equal(t, "Something went wrong. Please contact support.", repo.store[notificationID].Body)

		require.Len(t, eventDispatcher.events, 1)
		require.Equal(t, model.NotificationCreated{}.Type(), eventDispatcher.events[0].Type())
	})
	eventDispatcher.Reset()

	t.Run("Send notification", func(t *testing.T) {
		notificationID, err := notificationService.CreateNotification(name, subject, body)
		require.NoError(t, err)

		err = notificationService.SendNotification(notificationID, model.Recipient{Name: "Steve", Email: "test.steve@example.com"})
		require.NoError(t, err)

		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model.NotificationCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.NotificationSent{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Send non existed notification", func(t *testing.T) {
		notificationID := uuid.New()
		err := notificationService.SendNotification(notificationID, model.Recipient{Name: "Steve", Email: "test.steve@example.com"})
		require.ErrorIs(t, err, model.ErrNotificationNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})
	eventDispatcher.Reset()

	t.Run("Delete notification", func(t *testing.T) {
		notificationID, err := notificationService.CreateNotification(name, subject, body)
		require.NoError(t, err)

		err = notificationService.DeleteNotification(notificationID)
		require.NoError(t, err)

		require.NotNil(t, repo.store[notificationID])
		require.NotNil(t, repo.store[notificationID].DeletedAt)
		require.Len(t, eventDispatcher.events, 2)
		require.Equal(t, model.NotificationCreated{}.Type(), eventDispatcher.events[0].Type())
		require.Equal(t, model.NotificationDeleted{}.Type(), eventDispatcher.events[1].Type())
	})
	eventDispatcher.Reset()

	t.Run("Delete non existed notification", func(t *testing.T) {
		newID, _ := repo.NextID()
		err := notificationService.DeleteNotification(newID)
		require.ErrorIs(t, err, model.ErrNotificationNotFound)

		require.Len(t, eventDispatcher.events, 0)
	})
	eventDispatcher.Reset()
}

var _ model.NotificationRepository = &mockNotificationRepository{}

type mockNotificationRepository struct {
	store map[uuid.UUID]*model.Notification
}

func (m *mockNotificationRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockNotificationRepository) Store(notification *model.Notification) error {
	m.store[notification.ID] = notification
	return nil
}

func (m *mockNotificationRepository) Find(id uuid.UUID) (*model.Notification, error) {
	notification, ok := m.store[id]
	if !ok {
		return nil, model.ErrNotificationNotFound
	}
	if notification.DeletedAt != nil {
		return nil, model.ErrNotificationNotFound
	}
	return notification, nil
}

func (m *mockNotificationRepository) Delete(id uuid.UUID) error {
	deletedAt := time.Now()
	m.store[id].DeletedAt = &deletedAt
	return nil
}

type mockEventDispatcher struct {
	events []event.Event
}

func (m *mockEventDispatcher) Reset() {
	m.events = make([]event.Event, 0)
}

func (m *mockEventDispatcher) ListEvents() []event.Event {
	return m.events
}

func (m *mockEventDispatcher) Dispatch(evt event.Event) error {
	m.events = append(m.events, evt)
	return nil
}
