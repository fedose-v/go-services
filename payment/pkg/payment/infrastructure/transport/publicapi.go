package transport

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"payment/api/server/paymentpublicapi"
	appmodel "payment/pkg/payment/app/model"
	"payment/pkg/payment/app/query"
	"payment/pkg/payment/app/service"
)

func NewPaymentInternalAPI(
	balanceQueryService query.AccountBalanceQueryService,
	paymentService service.PaymentService,
) paymentpublicapi.PaymentPublicAPIServer {
	return &paymentInternalAPI{
		balanceQueryService: balanceQueryService,
		paymentService:      paymentService,
	}
}

type paymentInternalAPI struct {
	balanceQueryService query.AccountBalanceQueryService
	paymentService      service.PaymentService

	paymentpublicapi.UnimplementedPaymentPublicAPIServer
}

func (u paymentInternalAPI) StoreUser(ctx context.Context, request *paymentpublicapi.StoreUserBalanceRequest) (*paymentpublicapi.StoreCustomerBalanceResponse, error) {
	var (
		customerID uuid.UUID
		err        error
	)
	if request.CustomerID != "" {
		customerID, err = uuid.Parse(request.CustomerID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.CustomerID)
		}
	}

	balanceID, err := u.paymentService.StoreUserBalance(ctx, appmodel.CustomerBalance{
		CustomerID: customerID,
		Amount:     request.Balance,
	})
	if err != nil {
		return nil, err
	}

	return &paymentpublicapi.StoreCustomerBalanceResponse{
		BalanceID: balanceID.String(),
	}, nil
}

func (u paymentInternalAPI) FindUser(ctx context.Context, request *paymentpublicapi.FindCustomerBalanceRequest) (*paymentpublicapi.FindCustomerBalanceResponse, error) {
	balanceID, err := uuid.Parse(request.CustomerID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid uuid %q", request.CustomerID)
	}
	balance, err := u.balanceQueryService.FindBalance(ctx, balanceID)
	if err != nil {
		return nil, err
	}
	if balance == nil {
		return nil, status.Errorf(codes.NotFound, "balance %q not found", request.CustomerID)
	}
	return &paymentpublicapi.FindCustomerBalanceResponse{
		CustomerID: balance.CustomerID.String(),
		Balance:    balance.Amount,
	}, nil
}
