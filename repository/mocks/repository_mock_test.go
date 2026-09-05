package mocks_test

import (
	"context"
	"testing"

	"github.com/LalatinaHub/common/model"
	"github.com/LalatinaHub/common/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestMockUserRepository(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	ctx := context.Background()

	expectedUsers := map[string][]model.User{
		"vmess": {{ID: 1, Token: "t1"}},
	}

	mockRepo.On("GetActiveUsersGroupedByVPN", ctx).Return(expectedUsers, nil)
	mockRepo.On("DeductQuota", ctx, int64(1), int64(100)).Return(int64(900), false, nil)
	mockRepo.On("DeductQuotaBatch", ctx, map[int64]int64{1: 100}).Return([]int64{}, nil)
	mockRepo.On("CreateUser", ctx, &model.User{Token: "test"}).Return(int64(10), nil)

	users, err := mockRepo.GetActiveUsersGroupedByVPN(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)

	remaining, depleted, err := mockRepo.DeductQuota(ctx, 1, 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(900), remaining)
	assert.False(t, depleted)

	depletedUsers, err := mockRepo.DeductQuotaBatch(ctx, map[int64]int64{1: 100})
	assert.NoError(t, err)
	assert.Empty(t, depletedUsers)

	id, err := mockRepo.CreateUser(ctx, &model.User{Token: "test"})
	assert.NoError(t, err)
	assert.Equal(t, int64(10), id)

	mockRepo.AssertExpectations(t)
}

func TestMockServerRepository(t *testing.T) {
	mockRepo := new(mocks.MockServerRepository)
	ctx := context.Background()

	expectedServers := []model.Server{
		{ID: 1, Code: "SG1"},
	}

	mockRepo.On("GetAll", ctx).Return(expectedServers, nil)

	servers, err := mockRepo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedServers, servers)
	mockRepo.AssertExpectations(t)
}

func TestMockKVRepository(t *testing.T) {
	mockRepo := new(mocks.MockKVRepository)
	ctx := context.Background()

	expectedKV := map[string]any{
		"key1": "val1",
	}

	mockRepo.On("GetAll", ctx).Return(expectedKV, nil)

	kv, err := mockRepo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedKV, kv)
	mockRepo.AssertExpectations(t)
}

func TestMockProxyRepository(t *testing.T) {
	mockRepo := new(mocks.MockProxyRepository)
	ctx := context.Background()

	expectedRelays := []model.ProxyNode{
		{ID: 1, CountryCode: "SG"},
	}

	mockRepo.On("GetRelays", ctx, []string{"ID"}, 5).Return(expectedRelays, nil)

	relays, err := mockRepo.GetRelays(ctx, []string{"ID"}, 5)
	assert.NoError(t, err)
	assert.Equal(t, expectedRelays, relays)
	mockRepo.AssertExpectations(t)
}
