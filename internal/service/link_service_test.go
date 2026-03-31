package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ponchik327/Shortener/internal/domain"
)

// --- mocks ---

type mockLinkRepo struct{ mock.Mock }

func (m *mockLinkRepo) InsertLink(ctx context.Context, link *domain.Link) (*domain.Link, error) {
	args := m.Called(ctx, link)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Link), args.Error(1)
}

func (m *mockLinkRepo) GetByShortCode(ctx context.Context, code string) (*domain.Link, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Link), args.Error(1)
}

func (m *mockLinkRepo) ShortCodeExists(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

type mockLinkCache struct{ mock.Mock }

func (m *mockLinkCache) GetByCode(ctx context.Context, code string) (int64, string, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(int64), args.String(1), args.Error(2)
}

func (m *mockLinkCache) Set(ctx context.Context, code string, linkID int64, longURL string) error {
	args := m.Called(ctx, code, linkID, longURL)
	return args.Error(0)
}

// --- tests ---

func TestCreateLink_CustomCode_Success(t *testing.T) {
	repo := &mockLinkRepo{}
	cache := &mockLinkCache{}

	expected := &domain.Link{ID: 1, ShortCode: "mycode", LongURL: "http://example.com"}
	input := &domain.Link{ShortCode: "mycode", LongURL: "http://example.com"}
	repo.On("ShortCodeExists", mock.Anything, "mycode").Return(false, nil).Once()
	repo.On("InsertLink", mock.Anything, input).Return(expected, nil).Once()
	cache.On("Set", mock.Anything, "mycode", int64(1), "http://example.com").Return(nil).Once()

	svc := NewLinkService(repo, cache, noopLogger{}, 6)
	link, err := svc.CreateLink(context.Background(), "http://example.com", "mycode")

	require.NoError(t, err)
	assert.Equal(t, expected, link)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestCreateLink_CustomCode_Taken(t *testing.T) {
	repo := &mockLinkRepo{}
	cache := &mockLinkCache{}

	repo.On("ShortCodeExists", mock.Anything, "taken").Return(true, nil).Once()

	svc := NewLinkService(repo, cache, noopLogger{}, 6)
	_, err := svc.CreateLink(context.Background(), "http://example.com", "taken")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrCodeTaken)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestCreateLink_CustomCode_ExistsError(t *testing.T) {
	repo := &mockLinkRepo{}
	cache := &mockLinkCache{}

	repo.On("ShortCodeExists", mock.Anything, "mycode").Return(false, errors.New("db error")).Once()

	svc := NewLinkService(repo, cache, noopLogger{}, 6)
	_, err := svc.CreateLink(context.Background(), "http://example.com", "mycode")

	require.Error(t, err)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestCreateLink_AutoGen_FirstAttempt(t *testing.T) {
	repo := &mockLinkRepo{}
	cache := &mockLinkCache{}

	expected := &domain.Link{ID: 1, ShortCode: "abc123", LongURL: "http://example.com"}
	repo.On("ShortCodeExists", mock.Anything, mock.Anything).Return(false, nil).Once()
	repo.On("InsertLink", mock.Anything, mock.Anything).Return(expected, nil).Once()
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	svc := NewLinkService(repo, cache, noopLogger{}, 6)
	link, err := svc.CreateLink(context.Background(), "http://example.com", "")

	require.NoError(t, err)
	assert.NotNil(t, link)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestCreateLink_AutoGen_RetryThenSuccess(t *testing.T) {
	repo := &mockLinkRepo{}
	cache := &mockLinkCache{}

	repo.On("ShortCodeExists", mock.Anything, mock.Anything).Return(true, nil).Once()
	repo.On("ShortCodeExists", mock.Anything, mock.Anything).Return(false, nil).Once()
	expected := &domain.Link{ID: 1, ShortCode: "abc123", LongURL: "http://example.com"}
	repo.On("InsertLink", mock.Anything, mock.Anything).Return(expected, nil).Once()
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	svc := NewLinkService(repo, cache, noopLogger{}, 6)
	link, err := svc.CreateLink(context.Background(), "http://example.com", "")

	require.NoError(t, err)
	assert.NotNil(t, link)
	repo.AssertNumberOfCalls(t, "ShortCodeExists", 2)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestCreateLink_AutoGen_AllExhausted(t *testing.T) {
	repo := &mockLinkRepo{}
	cache := &mockLinkCache{}

	repo.On("ShortCodeExists", mock.Anything, mock.Anything).Return(true, nil)

	svc := NewLinkService(repo, cache, noopLogger{}, 6)
	_, err := svc.CreateLink(context.Background(), "http://example.com", "")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrGenerationFailed)
	repo.AssertNumberOfCalls(t, "ShortCodeExists", 5)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestCreateLink_InsertError(t *testing.T) {
	repo := &mockLinkRepo{}
	cache := &mockLinkCache{}

	repo.On("ShortCodeExists", mock.Anything, mock.Anything).Return(false, nil).Once()
	repo.On("InsertLink", mock.Anything, mock.Anything).Return(nil, errors.New("insert failed")).Once()

	svc := NewLinkService(repo, cache, noopLogger{}, 6)
	_, err := svc.CreateLink(context.Background(), "http://example.com", "")

	require.Error(t, err)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestGetLink_CacheHit(t *testing.T) {
	repo := &mockLinkRepo{}
	cache := &mockLinkCache{}

	cache.On("GetByCode", mock.Anything, "abc").Return(int64(7), "http://dest.com", nil).Once()

	svc := NewLinkService(repo, cache, noopLogger{}, 6)
	link, err := svc.GetLink(context.Background(), "abc")

	require.NoError(t, err)
	assert.Equal(t, int64(7), link.ID)
	repo.AssertNotCalled(t, "GetByShortCode")

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestGetLink_CacheMiss(t *testing.T) {
	repo := &mockLinkRepo{}
	cache := &mockLinkCache{}

	cache.On("GetByCode", mock.Anything, "abc").Return(int64(0), "", domain.ErrCacheMiss).Once()
	expected := &domain.Link{ID: 1, ShortCode: "abc", LongURL: "http://dest.com"}
	repo.On("GetByShortCode", mock.Anything, "abc").Return(expected, nil).Once()
	cache.On("Set", mock.Anything, "abc", int64(1), "http://dest.com").Return(nil).Once()

	svc := NewLinkService(repo, cache, noopLogger{}, 6)
	link, err := svc.GetLink(context.Background(), "abc")

	require.NoError(t, err)
	assert.Equal(t, int64(1), link.ID)
	assert.Equal(t, "http://dest.com", link.LongURL)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestGetLink_CacheError_WarnLog(t *testing.T) {
	repo := &mockLinkRepo{}
	cache := &mockLinkCache{}
	cl := &capturingLogger{}

	cache.On("GetByCode", mock.Anything, "abc").Return(int64(0), "", errors.New("conn refused")).Once()
	expected := &domain.Link{ID: 1, ShortCode: "abc", LongURL: "http://dest.com"}
	repo.On("GetByShortCode", mock.Anything, "abc").Return(expected, nil).Once()
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	svc := NewLinkService(repo, cache, cl, 6)
	link, err := svc.GetLink(context.Background(), "abc")

	require.NoError(t, err)
	assert.NotNil(t, link)
	assert.NotEmpty(t, cl.warnMessages)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestGetLink_NotFound(t *testing.T) {
	repo := &mockLinkRepo{}
	cache := &mockLinkCache{}

	cache.On("GetByCode", mock.Anything, "zzz").Return(int64(0), "", domain.ErrCacheMiss).Once()
	repo.On("GetByShortCode", mock.Anything, "zzz").Return(nil, domain.ErrNotFound).Once()

	svc := NewLinkService(repo, cache, noopLogger{}, 6)
	_, err := svc.GetLink(context.Background(), "zzz")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}
