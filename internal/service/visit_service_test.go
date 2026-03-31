package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ponchik327/Shortener/internal/domain"
)

// --- mocks ---

type mockVisitRepo struct{ mock.Mock }

func (m *mockVisitRepo) InsertVisit(ctx context.Context, visit *domain.Visit) error {
	args := m.Called(ctx, visit)
	return args.Error(0)
}

func (m *mockVisitRepo) GetAnalytics(ctx context.Context, linkID int64) (*domain.Analytics, error) {
	args := m.Called(ctx, linkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Analytics), args.Error(1)
}

type mockLinkSvc struct{ mock.Mock }

func (m *mockLinkSvc) GetLink(ctx context.Context, code string) (*domain.Link, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Link), args.Error(1)
}

// --- tests ---

func TestGetAnalytics_LinkNotFound(t *testing.T) {
	repo := &mockVisitRepo{}
	linkSvc := &mockLinkSvc{}

	linkSvc.On("GetLink", mock.Anything, "notfound").Return(nil, domain.ErrNotFound).Once()

	svc := NewVisitService(repo, linkSvc, noopLogger{})
	_, err := svc.GetAnalytics(context.Background(), "notfound")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	repo.AssertExpectations(t)
	linkSvc.AssertExpectations(t)
}

func TestGetAnalytics_Success(t *testing.T) {
	repo := &mockVisitRepo{}
	linkSvc := &mockLinkSvc{}

	link := &domain.Link{ID: 42, ShortCode: "abc", LongURL: "http://example.com"}
	linkSvc.On("GetLink", mock.Anything, "abc").Return(link, nil).Once()

	expected := &domain.Analytics{TotalVisits: 5}
	repo.On("GetAnalytics", mock.Anything, int64(42)).Return(expected, nil).Once()

	svc := NewVisitService(repo, linkSvc, noopLogger{})
	result, err := svc.GetAnalytics(context.Background(), "abc")

	require.NoError(t, err)
	assert.Equal(t, expected, result)

	repo.AssertExpectations(t)
	linkSvc.AssertExpectations(t)
}

func TestGetAnalytics_RepoError(t *testing.T) {
	repo := &mockVisitRepo{}
	linkSvc := &mockLinkSvc{}

	link := &domain.Link{ID: 1, ShortCode: "abc", LongURL: "http://example.com"}
	linkSvc.On("GetLink", mock.Anything, "abc").Return(link, nil).Once()
	repo.On("GetAnalytics", mock.Anything, int64(1)).Return(nil, assert.AnError).Once()

	svc := NewVisitService(repo, linkSvc, noopLogger{})
	_, err := svc.GetAnalytics(context.Background(), "abc")

	require.Error(t, err)

	repo.AssertExpectations(t)
	linkSvc.AssertExpectations(t)
}

func TestRecordVisitAsync_InsertVisitCalled(t *testing.T) {
	repo := &mockVisitRepo{}
	linkSvc := &mockLinkSvc{}

	done := make(chan struct{})
	repo.On("InsertVisit", mock.Anything, mock.MatchedBy(func(v *domain.Visit) bool {
		return v.LinkID == 99
	})).
		Run(func(_ mock.Arguments) { close(done) }).
		Return(nil).Once()

	svc := NewVisitService(repo, linkSvc, noopLogger{})
	svc.RecordVisitAsync(99, "Mozilla/5.0")

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("InsertVisit was not called within 3 seconds")
	}

	repo.AssertExpectations(t)
}
