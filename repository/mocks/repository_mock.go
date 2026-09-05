package mocks

import (
	"context"

	"github.com/LalatinaHub/common/model"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetActiveUsersGroupedByVPN(ctx context.Context) (map[string][]model.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string][]model.User), args.Error(1)
}

func (m *MockUserRepository) DeductQuota(ctx context.Context, userID int64, usedBytes int64) (int64, bool, error) {
	args := m.Called(ctx, userID, usedBytes)
	return args.Get(0).(int64), args.Bool(1), args.Error(2)
}

func (m *MockUserRepository) DeductQuotaBatch(ctx context.Context, usages map[int64]int64) ([]int64, error) {
	args := m.Called(ctx, usages)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, u *model.User) (int64, error) {
	args := m.Called(ctx, u)
	return args.Get(0).(int64), args.Error(1)
}

// MockServerRepository is a mock implementation of repository.ServerRepository
type MockServerRepository struct {
	mock.Mock
}

func (m *MockServerRepository) GetAll(ctx context.Context) ([]model.Server, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Server), args.Error(1)
}

// MockKVRepository is a mock implementation of repository.KVRepository
type MockKVRepository struct {
	mock.Mock
}

func (m *MockKVRepository) GetAll(ctx context.Context) (map[string]any, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

// MockProxyRepository is a mock implementation of repository.ProxyRepository
type MockProxyRepository struct {
	mock.Mock
}

func (m *MockProxyRepository) GetRelays(ctx context.Context, excludedCountryCodes []string, maxPerCountry int) ([]model.ProxyNode, error) {
	args := m.Called(ctx, excludedCountryCodes, maxPerCountry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ProxyNode), args.Error(1)
}
