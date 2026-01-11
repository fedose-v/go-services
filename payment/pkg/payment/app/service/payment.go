package service

import (
	"context"
	"errors"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	"payment/pkg/common/domain"
	appmodel "payment/pkg/payment/app/model"
	"payment/pkg/payment/domain/model"
	"payment/pkg/payment/domain/service"
)

type PaymentService interface {
	StoreUserBalance(ctx context.Context, balance appmodel.CustomerBalance) (uuid.UUID, error)
}

func NewPaymentService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) PaymentService {
	return &paymentService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type paymentService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (p *paymentService) StoreUserBalance(ctx context.Context, balance appmodel.CustomerBalance) (uuid.UUID, error) {
	var balanceID uuid.UUID
	err := p.luow.Execute(ctx, []string{"balance_" + balance.CustomerID.String()}, func(provider RepositoryProvider) error {
		domainService := p.domainService(ctx, provider.PaymentRepository(ctx), provider.AccountBalanceRepository(ctx))

		domainBalance, err := provider.AccountBalanceRepository(ctx).Find(balance.CustomerID)
		balanceID = domainBalance.ID
		if errors.Is(err, model.ErrPaymentNotFound) {
			balanceID, err = domainService.CreateCustomerBalance(balance.CustomerID)
		}
		if err != nil {
			return err
		}

		return domainService.UpdateBalance(balance.CustomerID, balance.Amount)
	})
	return balanceID, err
}

func (p *paymentService) domainService(
	ctx context.Context,
	paymentRepo model.PaymentRepository,
	accountBalanceRepo model.CustomerBalanceRepository,
) service.PaymentService {
	return service.NewPaymentService(paymentRepo, accountBalanceRepo, p.domainEventDispatcher(ctx))
}

func (p *paymentService) domainEventDispatcher(ctx context.Context) domain.EventDispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: p.eventDispatcher,
	}
}
