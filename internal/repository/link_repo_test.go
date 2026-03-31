package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ponchik327/Shortener/internal/domain"
	"github.com/ponchik327/Shortener/internal/repository"
)

func TestInsertLink(t *testing.T) {
	truncateTables(t)

	repo := repository.NewLinkRepository(testPG)
	link, err := repo.InsertLink(context.Background(), &domain.Link{
		ShortCode: "abc",
		LongURL:   "http://a.com",
	})

	require.NoError(t, err)
	assert.NotZero(t, link.ID)
	assert.Equal(t, "abc", link.ShortCode)
	assert.Equal(t, "http://a.com", link.LongURL)
	assert.False(t, link.CreatedAt.IsZero())
}

func TestInsertLink_DuplicateCode(t *testing.T) {
	truncateTables(t)

	repo := repository.NewLinkRepository(testPG)
	_, err := repo.InsertLink(context.Background(), &domain.Link{ShortCode: "dup", LongURL: "http://a.com"})
	require.NoError(t, err)

	_, err = repo.InsertLink(context.Background(), &domain.Link{ShortCode: "dup", LongURL: "http://b.com"})
	require.Error(t, err)
}

func TestGetByShortCode_Found(t *testing.T) {
	truncateTables(t)

	repo := repository.NewLinkRepository(testPG)
	inserted, err := repo.InsertLink(context.Background(), &domain.Link{
		ShortCode: "xyz",
		LongURL:   "http://xyz.com",
	})
	require.NoError(t, err)

	found, err := repo.GetByShortCode(context.Background(), "xyz")
	require.NoError(t, err)
	assert.Equal(t, inserted, found)
}

func TestGetByShortCode_NotFound(t *testing.T) {
	truncateTables(t)

	repo := repository.NewLinkRepository(testPG)
	_, err := repo.GetByShortCode(context.Background(), "zzz999")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestShortCodeExists_True(t *testing.T) {
	truncateTables(t)

	repo := repository.NewLinkRepository(testPG)
	_, err := repo.InsertLink(context.Background(), &domain.Link{ShortCode: "exists", LongURL: "http://x.com"})
	require.NoError(t, err)

	exists, err := repo.ShortCodeExists(context.Background(), "exists")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestShortCodeExists_False(t *testing.T) {
	truncateTables(t)

	repo := repository.NewLinkRepository(testPG)
	exists, err := repo.ShortCodeExists(context.Background(), "nosuchcode")
	require.NoError(t, err)
	assert.False(t, exists)
}
