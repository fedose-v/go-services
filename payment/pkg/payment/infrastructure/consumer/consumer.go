package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	appmodel "payment/pkg/payment/app/model"
	appservice "payment/pkg/payment/app/service"
)

type EventConsumer struct {
	conn           amqp.Connection
	paymentService appservice.PaymentService
	logger         logging.Logger
	ctx            context.Context
}

func NewEventConsumer(
	ctx context.Context,
	conn amqp.Connection,
	pool mysql.ConnectionPool,
	logger logging.Logger,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) (*EventConsumer, error) {
	luow := NewLockableUnitOfWork(pool)

	return &EventConsumer{
		conn:           conn,
		paymentService: appservice.NewPaymentService(luow, eventDispatcher),
		logger:         logger,
		ctx:            ctx,
	}, nil
}

func (c *EventConsumer) Handler() amqp.Handler {
	return c.handle
}

func (c *EventConsumer) handle(ctx context.Context, delivery amqp.Delivery) (err error) {
	l := c.logger.WithField("event_type", delivery.Type)
	l.Info("processing event")

	switch delivery.Type {
	case "user_created":
		var event struct {
			UserID string `json:"user_id"`
			Login  string `json:"login"`
		}
		if err = json.Unmarshal(delivery.Body, &event); err != nil {
			err = errors.Wrap(err, "failed to unmarshal user_created")
			return err
		}
		userID, parseErr := uuid.Parse(event.UserID)
		if parseErr != nil {
			l.Error(parseErr, "invalid user id in user event")
			return err
		}

		l.Info(fmt.Sprintf("Creating wallet for new user %s (%s)", event.Login, userID))
		balanceID, createErr := c.paymentService.StoreUserBalance(ctx, appmodel.CustomerBalance{
			CustomerID: userID,
			Amount:     100.00,
		})
		if createErr != nil {
			l.Error(createErr, "failed to create user wallet")
		}
		l.Info(fmt.Sprintf("Wallet %s was created for user %s", balanceID.String(), userID))
		return nil

	default:
		l.WithField("type", delivery.Type).Info("unhandled event type")
		return nil
	}
}
