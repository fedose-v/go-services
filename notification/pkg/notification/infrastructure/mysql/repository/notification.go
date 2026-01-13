package repository

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"notification/pkg/notification/domain/model"
	"notification/pkg/notification/infrastructure/metrics"
)

func NewNotificationRepository(ctx context.Context, client mysql.ClientContext) model.NotificationRepository {
	return &notificationRepository{
		ctx:    ctx,
		client: client,
	}
}

type notificationRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (n notificationRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (n notificationRepository) Store(notification *model.Notification) (err error) {
	start := time.Now()
	defer func() {
		status := metrics.StatusSuccess
		if err != nil {
			status = metrics.StatusError
		}
		metrics.DatabaseDuration.WithLabelValues("store", "notification", status).Observe(time.Since(start).Seconds())
	}()

	_, err = n.client.ExecContext(n.ctx,
		`INSERT INTO notification (id, name, subject, body) VALUES (?, ?, ?)`,
		notification.ID, notification.Name, notification.Subject, notification.Body,
	)
	return errors.WithStack(err)
}

func (n notificationRepository) Find(id uuid.UUID) (_ *model.Notification, err error) {
	start := time.Now()
	defer func() {
		status := metrics.StatusSuccess
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			status = metrics.StatusError
		}
		metrics.DatabaseDuration.WithLabelValues("find", "notification", status).Observe(time.Since(start).Seconds())
	}()

	var notification *model.Notification
	err = n.client.SelectContext(n.ctx, &notification, "SELECT * FROM notification WHERE id = ? ORDER BY created_at DESC", id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return notification, nil
}
