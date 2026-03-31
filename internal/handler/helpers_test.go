package handler

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/wb-go/wbf/logger"

	"github.com/ponchik327/Shortener/internal/domain"
)

// --- noopLogger ---

type noopLogger struct{}

func (noopLogger) Debug(msg string, args ...any)                                          {}
func (noopLogger) Info(msg string, args ...any)                                           {}
func (noopLogger) Warn(msg string, args ...any)                                           {}
func (noopLogger) Error(msg string, args ...any)                                          {}
func (noopLogger) Debugw(msg string, keysAndValues ...any)                                {}
func (noopLogger) Infow(msg string, keysAndValues ...any)                                 {}
func (noopLogger) Warnw(msg string, keysAndValues ...any)                                 {}
func (noopLogger) Errorw(msg string, keysAndValues ...any)                                {}
func (noopLogger) Ctx(_ context.Context) logger.Logger                                    { return noopLogger{} }
func (noopLogger) With(args ...any) logger.Logger                                         { return noopLogger{} }
func (noopLogger) WithGroup(name string) logger.Logger                                    { return noopLogger{} }
func (noopLogger) LogRequest(_ context.Context, _, _ string, _ int, _ time.Duration)      {}
func (noopLogger) Log(_ logger.Level, _ string, _ ...logger.Attr)                         {}
func (noopLogger) LogAttrs(_ context.Context, _ logger.Level, _ string, _ ...logger.Attr) {}

// --- constructor ---

func newTestHandler(lc linkCreator, lg linkGetter, vr visitRecorder, ap analyticsProvider) *Handler {
	return &Handler{
		linkCreator:       lc,
		linkGetter:        lg,
		visitRecorder:     vr,
		analyticsProvider: ap,
		log:               noopLogger{},
	}
}

// --- mocks ---

type mockLinkCreator struct{ mock.Mock }

func (m *mockLinkCreator) CreateLink(ctx context.Context, longURL, customCode string) (*domain.Link, error) {
	args := m.Called(ctx, longURL, customCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Link), args.Error(1)
}

type mockLinkGetter struct{ mock.Mock }

func (m *mockLinkGetter) GetLink(ctx context.Context, code string) (*domain.Link, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Link), args.Error(1)
}

type mockVisitRecorder struct{ mock.Mock }

func (m *mockVisitRecorder) RecordVisitAsync(linkID int64, userAgent string) {
	m.Called(linkID, userAgent)
}

type mockAnalyticsProvider struct{ mock.Mock }

func (m *mockAnalyticsProvider) GetAnalytics(ctx context.Context, code string) (*domain.Analytics, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Analytics), args.Error(1)
}
