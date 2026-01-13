package query

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	appmodel "notification/pkg/notification/app/model"
	"notification/pkg/notification/app/query"
	"notification/pkg/notification/infrastructure/metrics"
)

func NewNotificationQueryService(client mysql.ClientContext) query.NotificationQueryService {
	return &notificationQueryService{
		client: client,
	}
}

type notificationQueryService struct {
	client mysql.ClientContext
}

func (s *notificationQueryService) Find(ctx context.Context, id uuid.UUID) (_ *appmodel.Notification, err error) {
	start := time.Now()
	defer func() {
		status := metrics.StatusSuccess
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			status = metrics.StatusError
		}
		metrics.DatabaseDuration.WithLabelValues("find_query", "notification", status).Observe(time.Since(start).Seconds())
	}()

	var data struct {
		ID        uuid.UUID `db:"notification_id"`
		Name      string    `db:"user_id"`
		Subject   string    `db:"order_id"`
		Body      string    `db:"message"`
		CreatedAt time.Time `db:"created_at"`
	}

	err = s.client.SelectContext(ctx, &data, "SELECT * FROM notification WHERE id = ? ORDER BY created_at DESC", id)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	notification := appmodel.Notification{
		ID:        data.ID,
		Name:      data.Name,
		Subject:   data.Subject,
		Body:      data.Body,
		CreatedAt: data.CreatedAt.Unix(),
	}

	return &notification, nil
}
