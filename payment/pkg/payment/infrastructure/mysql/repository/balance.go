package repository

import (
	"context"
	"database/sql"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"payment/pkg/payment/domain/model"
)

func NewBalanceRepository(ctx context.Context, client mysql.ClientContext) model.CustomerBalanceRepository {
	return &balanceRepository{
		ctx:    ctx,
		client: client,
	}
}

type balanceRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (b balanceRepository) Store(balance model.CustomerAccountBalance) (uuid.UUID, error) {
	_, err := b.client.ExecContext(b.ctx,
		`
	INSERT INTO customer_account_balance (id, customer_id, amount, created_at, updated_at) VALUES (?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		amount=VALUES(amount),
		created_at=VALUES(created_at),
		updated_at=VALUES(updated_at)
	`,
		balance.ID[:],
		balance.CustomerID[:],
		balance.Amount,
		balance.CreatedAt,
		balance.UpdatedAt,
	)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	return balance.ID, nil
}

func (b balanceRepository) Find(customerID uuid.UUID) (*model.CustomerAccountBalance, error) {
	balance := struct {
		ID         uuid.UUID `db:"id"`
		CustomerID uuid.UUID `db:"customer_id"`
		Amount     float64   `db:"amount"`
		CreatedAt  time.Time `db:"created_at"`
		UpdatedAt  time.Time `db:"updated_at"`
	}{}

	err := b.client.GetContext(
		b.ctx,
		&balance,
		`SELECT id, customer_id, amount, created_at, updated_at FROM customer_account_balance WHERE customer_id = ?`,
		customerID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrBalanceNotFound
		}
		return nil, errors.WithStack(err)
	}

	return &model.CustomerAccountBalance{
		ID:         balance.ID,
		CustomerID: balance.CustomerID,
		Amount:     balance.Amount,
		CreatedAt:  balance.CreatedAt,
		UpdatedAt:  &balance.UpdatedAt,
	}, nil
}
