package transport

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"notification/api/server/notificationinternal"
	"notification/pkg/notification/app/query"
)

func NewNotificationInternalAPI(
	queryService query.NotificationQueryService,
) notificationinternal.NotificationInternalServiceServer {
	return &notificationInternalAPI{
		queryService: queryService,
	}
}

type notificationInternalAPI struct {
	queryService query.NotificationQueryService
	notificationinternal.UnimplementedNotificationInternalServiceServer
}

func (a *notificationInternalAPI) FindNotificationsForUser(ctx context.Context, request *notificationinternal.FindNotificationsForUserRequest) (*notificationinternal.FindNotificationsForUserResponse, error) {
	userID, err := uuid.Parse(request.NotificationID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid user id")
	}

	notification, err := a.queryService.Find(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := notificationinternal.Notification{
		NotificationID: notification.ID.String(),
		Name:           notification.Name,
		Subject:        notification.Subject,
		Body:           notification.Body,
		CreatedAt:      notification.CreatedAt,
	}

	return &notificationinternal.FindNotificationsForUserResponse{
		Notification: &result,
	}, nil
}
