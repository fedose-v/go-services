package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"notification/pkg/notification/domain/model"
	"notification/pkg/notification/domain/service"
)

func TestNotificationService(t *testing.T) {
	repo := &mockNotificationRepository{
		store: make(map[uuid.UUID]*model.Notification),
	}

	notificationService := service.NewNotificationService(repo)

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
	})
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
	return notification, nil
}
