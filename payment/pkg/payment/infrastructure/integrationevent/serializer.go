package integrationevent

import (
	"encoding/json"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/pkg/errors"

	"payment/pkg/payment/domain/model"
)

func NewEventSerializer() outbox.EventSerializer[outbox.Event] {
	return &eventSerializer{}
}

type eventSerializer struct{}

func (s eventSerializer) Serialize(event outbox.Event) (string, error) {
	switch e := event.(type) {
	case *model.CustomerAccountCreated:
		b, err := json.Marshal(AccountCreated{
			CustomerID: e.CustomerID.String(),
			CreatedAt:  e.CreatedAt.Unix(),
		})
		return string(b), errors.WithStack(err)
	default:
		return "", errors.Errorf("unknown event %q", event.Type())
	}
}

type AccountCreated struct {
	CustomerID string `json:"customer_id"`
	CreatedAt  int64  `json:"created_at"`
}
