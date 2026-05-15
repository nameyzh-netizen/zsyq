//go:build unit

package service

import (
	"context"
	"strings"
	"testing"

	"github.com/nameyzh-netizen/zsyq/internal/config"
	"github.com/stretchr/testify/require"
)

type settingAdminAPIKeyRepoStub struct {
	values  map[string]string
	deleted map[string]bool
}

func (s *settingAdminAPIKeyRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *settingAdminAPIKeyRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *settingAdminAPIKeyRepoStub) Set(ctx context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *settingAdminAPIKeyRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *settingAdminAPIKeyRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *settingAdminAPIKeyRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingAdminAPIKeyRepoStub) Delete(ctx context.Context, key string) error {
	if s.deleted == nil {
		s.deleted = map[string]bool{}
	}
	delete(s.values, key)
	s.deleted[key] = true
	return nil
}

func TestSettingService_AdminAPIKeyStoresHashOnly(t *testing.T) {
	repo := &settingAdminAPIKeyRepoStub{}
	svc := NewSettingService(repo, &config.Config{})

	key, err := svc.GenerateAdminAPIKey(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, key)
	require.True(t, strings.HasPrefix(key, AdminAPIKeyPrefix))

	stored := repo.values[SettingKeyAdminAPIKey]
	require.NotEqual(t, key, stored)
	record, ok := parseAdminAPIKeyRecord(stored)
	require.True(t, ok)
	require.Equal(t, adminAPIKeyHashAlgorithm, record.Algorithm)
	require.NotEmpty(t, record.Hash)
	require.NotContains(t, stored, key)

	valid, err := svc.ValidateAdminAPIKey(context.Background(), key)
	require.NoError(t, err)
	require.True(t, valid)

	valid, err = svc.ValidateAdminAPIKey(context.Background(), key+"wrong")
	require.NoError(t, err)
	require.False(t, valid)

	masked, exists, err := svc.GetAdminAPIKeyStatus(context.Background())
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, "admin-****", masked)
}

func TestSettingService_AdminAPIKeyMigratesLegacyPlaintext(t *testing.T) {
	legacyKey := AdminAPIKeyPrefix + "legacy"
	repo := &settingAdminAPIKeyRepoStub{values: map[string]string{SettingKeyAdminAPIKey: legacyKey}}
	svc := NewSettingService(repo, &config.Config{})

	valid, err := svc.ValidateAdminAPIKey(context.Background(), legacyKey)
	require.NoError(t, err)
	require.True(t, valid)

	stored := repo.values[SettingKeyAdminAPIKey]
	require.NotEqual(t, legacyKey, stored)
	record, ok := parseAdminAPIKeyRecord(stored)
	require.True(t, ok)
	require.Equal(t, hashAdminAPIKey(legacyKey), record.Hash)
}

func TestSettingService_AdminAPIKeyDeleteRemovesStoredKey(t *testing.T) {
	repo := &settingAdminAPIKeyRepoStub{values: map[string]string{SettingKeyAdminAPIKey: `{"version":1}`}}
	svc := NewSettingService(repo, &config.Config{})

	require.NoError(t, svc.DeleteAdminAPIKey(context.Background()))
	require.True(t, repo.deleted[SettingKeyAdminAPIKey])
	_, ok := repo.values[SettingKeyAdminAPIKey]
	require.False(t, ok)
}
