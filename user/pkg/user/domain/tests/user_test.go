package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"user/pkg/common/domain"
	"user/pkg/user/domain/model"
	"user/pkg/user/domain/service"
)

func TestUserService_CreateUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	userService := service.NewUserService(repo, dispatcher)

	login := "john_doe"
	userID := uuid.New()

	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Login != nil && *spec.Login == login
	})).Return((*model.User)(nil), model.ErrUserNotFound)

	repo.On("NextID").Return(userID, nil)

	repo.On("Store", mock.MatchedBy(func(u model.User) bool {
		return u.UserID == userID &&
			u.Login == login &&
			u.Status == model.Blocked &&
			!u.CreatedAt.IsZero() &&
			!u.UpdatedAt.IsZero()
	})).Return(nil)

	dispatcher.On("Dispatch", mock.MatchedBy(func(e domain.Event) bool {
		evt, ok := e.(*model.UserCreated)
		return ok && evt.UserID == userID && evt.Login == login
	})).Return(nil)

	resultID, err := userService.CreateUser(model.Active, login)
	require.NoError(t, err)
	assert.Equal(t, userID, resultID)

	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestUserService_CreateUser_LoginAlreadyUsed(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	userService := service.NewUserService(repo, dispatcher)

	login := "existing_user"
	existingUser := &model.User{UserID: uuid.New(), Login: login}

	repo.On("Find", mock.AnythingOfType("model.FindSpec")).Return(existingUser, nil)

	_, err := userService.CreateUser(model.Active, login)
	require.ErrorIs(t, err, model.ErrUserLoginAlreadyUsed)

	repo.AssertExpectations(t)
	dispatcher.AssertNotCalled(t, "Dispatch")
}

func TestUserService_UpdateUserStatus_Success(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	userService := service.NewUserService(repo, dispatcher)

	userID := uuid.New()
	oldStatus := model.Blocked
	newStatus := model.Active
	user := &model.User{UserID: userID, Status: oldStatus}

	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.UserID != nil && *spec.UserID == userID
	})).Return(user, nil)

	repo.On("Store", mock.MatchedBy(func(u model.User) bool {
		return u.UserID == userID && u.Status == newStatus
	})).Return(nil)

	dispatcher.On("Dispatch", mock.MatchedBy(func(e domain.Event) bool {
		evt, ok := e.(*model.UserUpdated)
		return ok && evt.UserID == userID &&
			evt.UpdatedFields != nil &&
			evt.UpdatedFields.Status != nil &&
			*evt.UpdatedFields.Status == newStatus
	})).Return(nil)

	err := userService.UpdateUserStatus(userID, newStatus)
	require.NoError(t, err)

	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestUserService_UpdateUserStatus_NoChange(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	userService := service.NewUserService(repo, dispatcher)

	userID := uuid.New()
	status := model.Active
	user := &model.User{UserID: userID, Status: status}

	repo.On("Find", mock.AnythingOfType("model.FindSpec")).Return(user, nil)

	err := userService.UpdateUserStatus(userID, status)
	require.NoError(t, err)

	repo.AssertExpectations(t)
	dispatcher.AssertNotCalled(t, "Dispatch")
}

func TestUserService_UpdateUserEmail_Success(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	userService := service.NewUserService(repo, dispatcher)

	userID := uuid.New()
	email := "new@example.com"
	user := &model.User{UserID: userID, Email: nil}

	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.UserID != nil && *spec.UserID == userID
	})).Return(user, nil)

	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Email != nil && *spec.Email == email
	})).Return((*model.User)(nil), model.ErrUserNotFound)

	repo.On("Store", mock.MatchedBy(func(u model.User) bool {
		return u.UserID == userID && u.Email != nil && *u.Email == email
	})).Return(nil)

	dispatcher.On("Dispatch", mock.MatchedBy(func(e domain.Event) bool {
		evt, ok := e.(*model.UserUpdated)
		return ok && evt.UserID == userID &&
			evt.UpdatedFields != nil &&
			evt.UpdatedFields.Email != nil &&
			*evt.UpdatedFields.Email == email
	})).Return(nil)

	err := userService.UpdateUserEmail(userID, &email)
	require.NoError(t, err)

	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestUserService_UpdateUserEmail_EmailAlreadyUsed(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	userService := service.NewUserService(repo, dispatcher)

	userID := uuid.New()
	email := "taken@example.com"
	user := &model.User{UserID: userID, Email: nil}
	otherUser := &model.User{UserID: uuid.New(), Email: &email}

	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.UserID != nil && *spec.UserID == userID
	})).Return(user, nil)

	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Email != nil && *spec.Email == email
	})).Return(otherUser, nil)

	err := userService.UpdateUserEmail(userID, &email)
	require.ErrorIs(t, err, model.ErrUserEmailAlreadyUsed)

	repo.AssertExpectations(t)
	dispatcher.AssertNotCalled(t, "Dispatch")
}

func TestUserService_UpdateUserEmail_Remove(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	userService := service.NewUserService(repo, dispatcher)

	userID := uuid.New()
	oldEmail := "old@example.com"
	user := &model.User{UserID: userID, Email: &oldEmail}

	repo.On("Find", mock.AnythingOfType("model.FindSpec")).Return(user, nil)
	repo.On("Store", mock.MatchedBy(func(u model.User) bool {
		return u.UserID == userID && u.Email == nil
	})).Return(nil)

	dispatcher.On("Dispatch", mock.MatchedBy(func(e domain.Event) bool {
		evt, ok := e.(*model.UserUpdated)
		return ok && evt.RemovedFields != nil && *evt.RemovedFields.Email
	})).Return(nil)

	err := userService.UpdateUserEmail(userID, nil)
	require.NoError(t, err)

	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestUserService_UpdateUserTelegram_TelegramAlreadyUsed(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	userService := service.NewUserService(repo, dispatcher)

	userID := uuid.New()
	telegram := "@taken"
	user := &model.User{UserID: userID, Telegram: nil}
	otherUser := &model.User{UserID: uuid.New(), Telegram: &telegram}

	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.UserID != nil && *spec.UserID == userID
	})).Return(user, nil)

	repo.On("Find", mock.MatchedBy(func(spec model.FindSpec) bool {
		return spec.Telegram != nil && *spec.Telegram == telegram
	})).Return(otherUser, nil)

	err := userService.UpdateUserTelegram(userID, &telegram)
	require.ErrorIs(t, err, model.ErrUserTelegramAlreadyUsed)
}

func TestUserService_DeleteUser_SoftDelete(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	userService := service.NewUserService(repo, dispatcher)

	userID := uuid.New()
	user := &model.User{UserID: userID, Status: model.Active}

	repo.On("Find", mock.AnythingOfType("model.FindSpec")).Return(user, nil)

	repo.On("Store", mock.MatchedBy(func(u model.User) bool {
		return u.UserID == userID &&
			u.Status == model.Deleted &&
			u.DeletedAt != nil
	})).Return(nil)

	dispatcher.On("Dispatch", mock.MatchedBy(func(e domain.Event) bool {
		evt, ok := e.(*model.UserDeleted)
		return ok && evt.UserID == userID && !evt.Hard
	})).Return(nil)

	err := userService.DeleteUser(userID, false)
	require.NoError(t, err)

	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestUserService_DeleteUser_HardDelete(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	userService := service.NewUserService(repo, dispatcher)

	userID := uuid.New()
	user := &model.User{UserID: userID, Status: model.Active}

	repo.On("Find", mock.AnythingOfType("model.FindSpec")).Return(user, nil)
	repo.On("HardDelete", userID).Return(nil)

	repo.On("Store", mock.MatchedBy(func(u model.User) bool {
		return u.UserID == userID && u.Status == model.Deleted
	})).Return(nil)

	dispatcher.On("Dispatch", mock.MatchedBy(func(e domain.Event) bool {
		evt, ok := e.(*model.UserDeleted)
		return ok && evt.Hard
	})).Return(nil)

	err := userService.DeleteUser(userID, true)
	require.NoError(t, err)

	repo.AssertExpectations(t)
	dispatcher.AssertExpectations(t)
}

func TestUserService_UserNotFound(t *testing.T) {
	repo := new(MockUserRepository)
	dispatcher := new(MockEventDispatcher)
	userService := service.NewUserService(repo, dispatcher)

	userID := uuid.New()
	repo.On("Find", mock.AnythingOfType("model.FindSpec")).Return((*model.User)(nil), model.ErrUserNotFound)

	tests := []func() error{
		func() error { _, err := userService.CreateUser(model.Active, "login"); return err },
		func() error { return userService.UpdateUserStatus(userID, model.Active) },
		func() error { email := "x"; return userService.UpdateUserEmail(userID, &email) },
		func() error { tg := "@x"; return userService.UpdateUserTelegram(userID, &tg) },
		func() error { return userService.DeleteUser(userID, false) },
	}

	for i, test := range tests {
		if i == 0 {
			continue
		}
		err := test()
		require.ErrorIs(t, err, model.ErrUserNotFound, "test %d", i)
	}
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) NextID() (uuid.UUID, error) {
	args := m.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserRepository) Store(user model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Find(spec model.FindSpec) (*model.User, error) {
	args := m.Called(spec)
	if user, ok := args.Get(0).(*model.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) HardDelete(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

type MockEventDispatcher struct {
	mock.Mock
}

func (m *MockEventDispatcher) Dispatch(event domain.Event) error {
	args := m.Called(event)
	return args.Error(0)
}
