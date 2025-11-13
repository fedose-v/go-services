package transport

import "user/api/server/userinternal"

func NewUserAPI() userinternal.MicroserviceTemplateInternalServiceServer {
	return &userAPI{}
}

type userAPI struct {
}

func (u userAPI) Ping(ctx context.Context, request *userinternal.PingRequest) (*userinternal.PingResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (u userAPI) CreateUser(ctx context.Context, request *userinternal.CreateUserRequest) (*userinternal.CreateUserResponse, error) {
	//TODO implement me
	panic("implement me")
}
