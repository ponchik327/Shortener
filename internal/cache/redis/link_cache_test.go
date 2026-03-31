package rediscache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rediscache "github.com/ponchik327/Shortener/internal/cache/redis"
	"github.com/ponchik327/Shortener/internal/domain"
)

func TestSet_GetByCode_Hit(t *testing.T) {
	cache := rediscache.New(testRedisClient, 5*time.Minute)
	code := t.Name()

	err := cache.Set(context.Background(), code, int64(1), "http://x.com")
	require.NoError(t, err)

	id, url, err := cache.GetByCode(context.Background(), code)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.Equal(t, "http://x.com", url)
}

func TestGetByCode_Miss(t *testing.T) {
	cache := rediscache.New(testRedisClient, 5*time.Minute)
	code := t.Name()

	_, _, err := cache.GetByCode(context.Background(), code)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCacheMiss)
}

func TestTTLExpiry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TTL test in -short mode")
	}

	cache := rediscache.New(testRedisClient, 100*time.Millisecond)
	code := t.Name()

	err := cache.Set(context.Background(), code, int64(1), "http://ttl.com")
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	_, _, err = cache.GetByCode(context.Background(), code)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCacheMiss)
}
