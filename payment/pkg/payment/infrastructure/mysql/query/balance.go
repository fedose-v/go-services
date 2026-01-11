package query

import (
	"context"
	"database/sql"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	appmodel "payment/pkg/payment/app/model"
	"payment/pkg/payment/app/query"
	"payment/pkg/payment/domain/model"
)

func NewAccountBalanceQueryService(client mysql.ClientContext) query.AccountBalanceQueryService {
	return &accountQueryService{
		client: client,
	}
}

type accountQueryService struct {
	client mysql.ClientContext
}

func (a accountQueryService) FindBalance(ctx context.Context, id uuid.UUID) (*appmodel.CustomerBalance, error) {
	account := struct {
		CustomerID uuid.UUID `db:"user_id"`
		Amount     float64   `db:"balance"`
	}{}

	err := a.client.GetContext(
		ctx,
		&account,
		`SELECT customer_id, amount FROM customer_account_balance WHERE customer_id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrBalanceNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &appmodel.CustomerBalance{
		CustomerID: account.CustomerID,
		Amount:     account.Amount,
	}, nil
}
