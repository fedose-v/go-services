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

func NewPaymentRepository(ctx context.Context, client mysql.ClientContext) model.PaymentRepository {
	return &paymentRepository{
		ctx:    ctx,
		client: client,
	}
}

type paymentRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (p paymentRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (p paymentRepository) Store(transaction *model.Transaction) error {
	_, err := p.client.ExecContext(p.ctx,
		`
	INSERT INTO transaction (order_id, customer_id, type, transaction.amount, payment_date) VALUES (?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
		amount=VALUES(amount)
	`,
		transaction.OrderID,
		transaction.CustomerID,
		transaction.Type,
		transaction.Amount,
		transaction.PaymentDate,
	)
	return errors.WithStack(err)
}

func (p paymentRepository) Find(id uuid.UUID) (*model.Transaction, error) {
	transaction := struct {
		ID          uuid.UUID `db:"id"`
		OrderID     uuid.UUID `db:"order_id"`
		CustomerID  uuid.UUID `db:"customer_id"`
		Type        int       `db:"type"`
		Amount      float64   `db:"amount"`
		PaymentDate time.Time `db:"payment_date"`
	}{}

	err := p.client.GetContext(
		p.ctx,
		&transaction,
		`SELECT id, order_id, customer_id, type, transaction.amount, payment_date FROM transaction WHERE id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(model.ErrPaymentNotFound)
		}
		return nil, errors.WithStack(err)
	}

	return &model.Transaction{
		ID:          transaction.ID,
		OrderID:     transaction.OrderID,
		CustomerID:  transaction.CustomerID,
		Type:        model.TransactionType(transaction.Type),
		Amount:      transaction.Amount,
		PaymentDate: transaction.PaymentDate,
	}, nil
}
