package rediscache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	wbfredis "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"

	"github.com/ponchik327/Shortener/internal/domain"
)

const (
	_keyPrefix     = "link:"
	_retryAttempts = 3
	_retryDelay    = 50 * time.Millisecond
)

// cachedLink — payload, хранящийся в Redis для короткого кода.
type cachedLink struct {
	ID      int64  `json:"id"`
	LongURL string `json:"url"`
}

// LinkCache оборачивает Redis-клиент, предоставляя доменные операции для коротких ссылок.
type LinkCache struct {
	client   *wbfredis.Client
	strategy retry.Strategy
	ttl      time.Duration
}

// New создаёт новый LinkCache.
func New(client *wbfredis.Client, ttl time.Duration) *LinkCache {
	return &LinkCache{
		client:   client,
		strategy: retry.Strategy{Attempts: _retryAttempts, Delay: _retryDelay},
		ttl:      ttl,
	}
}

// GetByCode возвращает закэшированную запись ссылки по короткому коду.
// Возвращает domain.ErrCacheMiss, если ключ отсутствует в кэше.
func (c *LinkCache) GetByCode(ctx context.Context, code string) (linkID int64, longURL string, err error) {
	raw, err := c.client.GetWithRetry(ctx, c.strategy, _keyPrefix+code)
	if err != nil {
		if isNotFound(err) {
			return 0, "", domain.ErrCacheMiss
		}

		return 0, "", fmt.Errorf("redis get link %s: %w", code, err)
	}

	var cl cachedLink
	if err = json.Unmarshal([]byte(raw), &cl); err != nil {
		return 0, "", fmt.Errorf("unmarshal cached link %s: %w", code, err)
	}

	return cl.ID, cl.LongURL, nil
}

// Set сохраняет linkID и longURL в кэш с настроенным TTL.
func (c *LinkCache) Set(ctx context.Context, code string, linkID int64, longURL string) error {
	payload, err := json.Marshal(cachedLink{ID: linkID, LongURL: longURL})
	if err != nil {
		return fmt.Errorf("marshal link %s: %w", code, err)
	}

	if err = c.client.SetWithExpirationAndRetry(ctx, c.strategy, _keyPrefix+code, string(payload), c.ttl); err != nil {
		return fmt.Errorf("redis set link %s: %w", code, err)
	}

	return nil
}

func isNotFound(err error) bool {
	return errors.Is(err, wbfredis.NoMatches)
}
