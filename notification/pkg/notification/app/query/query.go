package query

import (
	"context"

	"github.com/google/uuid"

	appmodel "notification/pkg/notification/app/model"
)

type NotificationQueryService interface {
	Find(ctx context.Context, id uuid.UUID) (*appmodel.Notification, error)
}
