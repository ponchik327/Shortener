package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wb-go/wbf/logger"
)

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

func truncateTables(t *testing.T) {
	t.Helper()
	_, err := testPG.Exec(context.Background(), "TRUNCATE TABLE visits, links RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}
