package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ponchik327/Shortener/internal/domain"
	"github.com/ponchik327/Shortener/internal/repository"
)

func insertTestLink(t *testing.T, code, url string) int64 {
	t.Helper()
	repo := repository.NewLinkRepository(testPG)
	link, err := repo.InsertLink(context.Background(), &domain.Link{
		ShortCode: code,
		LongURL:   url,
	})
	require.NoError(t, err)
	return link.ID
}

func insertVisitAt(
	t *testing.T,
	repo *repository.VisitRepository,
	linkID int64,
	ua string,
	visitedAt time.Time,
) {
	t.Helper()
	err := repo.InsertVisit(context.Background(), &domain.Visit{
		LinkID:    linkID,
		UserAgent: ua,
		VisitedAt: visitedAt,
	})
	require.NoError(t, err)
}

func TestInsertVisit(t *testing.T) {
	truncateTables(t)
	linkID := insertTestLink(t, "vtest1", "http://v1.com")

	repo := repository.NewVisitRepository(testPG)
	err := repo.InsertVisit(context.Background(), &domain.Visit{
		LinkID:    linkID,
		UserAgent: "TestAgent",
		VisitedAt: time.Now(),
	})
	require.NoError(t, err)
}

func TestGetAnalytics_Total(t *testing.T) {
	truncateTables(t)
	linkID := insertTestLink(t, "total1", "http://total1.com")

	now := time.Now().UTC()
	repo := repository.NewVisitRepository(testPG)
	insertVisitAt(t, repo, linkID, "UA1", now)
	insertVisitAt(t, repo, linkID, "UA2", now)
	insertVisitAt(t, repo, linkID, "UA3", now)

	analytics, err := repo.GetAnalytics(context.Background(), linkID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), analytics.TotalVisits)
}

func TestGetAnalytics_ByDay(t *testing.T) {
	truncateTables(t)
	linkID := insertTestLink(t, "byday1", "http://byday1.com")

	today := time.Now().UTC()
	yesterday := today.AddDate(0, 0, -1)

	repo := repository.NewVisitRepository(testPG)
	insertVisitAt(t, repo, linkID, "UA1", today)
	insertVisitAt(t, repo, linkID, "UA2", today)
	insertVisitAt(t, repo, linkID, "UA3", yesterday)

	analytics, err := repo.GetAnalytics(context.Background(), linkID)
	require.NoError(t, err)
	assert.Len(t, analytics.ByDay, 2)

	counts := make(map[string]int64)
	for _, d := range analytics.ByDay {
		counts[d.Day] = d.Count
	}
	assert.Equal(t, int64(2), counts[today.Format("2006-01-02")])
	assert.Equal(t, int64(1), counts[yesterday.Format("2006-01-02")])
}

func TestGetAnalytics_ByMonth(t *testing.T) {
	truncateTables(t)
	linkID := insertTestLink(t, "bymonth1", "http://bymonth1.com")

	jan := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	feb := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)

	repo := repository.NewVisitRepository(testPG)
	insertVisitAt(t, repo, linkID, "UA1", jan)
	insertVisitAt(t, repo, linkID, "UA2", feb)

	analytics, err := repo.GetAnalytics(context.Background(), linkID)
	require.NoError(t, err)
	assert.Len(t, analytics.ByMonth, 2)

	counts := make(map[string]int64)
	for _, mc := range analytics.ByMonth {
		counts[mc.Month] = mc.Count
	}
	assert.Equal(t, int64(1), counts["2026-01"])
	assert.Equal(t, int64(1), counts["2026-02"])
}

func TestGetAnalytics_ByUserAgent_Order(t *testing.T) {
	truncateTables(t)
	linkID := insertTestLink(t, "byua1", "http://byua1.com")

	now := time.Now().UTC()
	repo := repository.NewVisitRepository(testPG)
	insertVisitAt(t, repo, linkID, "Chrome/120", now)
	insertVisitAt(t, repo, linkID, "Chrome/120", now)
	insertVisitAt(t, repo, linkID, "Chrome/120", now)
	insertVisitAt(t, repo, linkID, "Firefox", now)

	analytics, err := repo.GetAnalytics(context.Background(), linkID)
	require.NoError(t, err)
	assert.Len(t, analytics.ByUserAgent, 2)

	counts := make(map[string]int64)
	for _, ua := range analytics.ByUserAgent {
		counts[ua.UserAgent] = ua.Count
	}
	assert.Equal(t, int64(3), counts["Chrome/120"])
	assert.Equal(t, int64(1), counts["Firefox"])
}

func TestGetAnalytics_Empty(t *testing.T) {
	truncateTables(t)
	linkID := insertTestLink(t, "empty1", "http://empty1.com")

	repo := repository.NewVisitRepository(testPG)
	analytics, err := repo.GetAnalytics(context.Background(), linkID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), analytics.TotalVisits)
	assert.Empty(t, analytics.ByDay)
	assert.Empty(t, analytics.ByMonth)
	assert.Empty(t, analytics.ByUserAgent)
}
