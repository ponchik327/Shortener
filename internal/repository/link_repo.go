package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"

	"github.com/ponchik327/Shortener/internal/domain"
)

// LinkRepository выполняет операции с таблицей links в базе данных.
type LinkRepository struct {
	pg *pgxdriver.Postgres
}

// NewLinkRepository создаёт новый LinkRepository.
func NewLinkRepository(pg *pgxdriver.Postgres) *LinkRepository {
	return &LinkRepository{pg: pg}
}

// InsertLink вставляет новую ссылку и возвращает созданную запись (с заполненными id и created_at).
func (r *LinkRepository) InsertLink(ctx context.Context, link *domain.Link) (*domain.Link, error) {
	sql, args, err := r.pg.Builder.
		Insert("links").
		Columns("short_code", "long_url").
		Values(link.ShortCode, link.LongURL).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert link sql: %w", err)
	}

	result := &domain.Link{
		ShortCode: link.ShortCode,
		LongURL:   link.LongURL,
	}

	if err = r.pg.QueryRow(ctx, sql, args...).Scan(&result.ID, &result.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert link: %w", err)
	}

	return result, nil
}

// GetByShortCode возвращает ссылку по короткому коду.
// Возвращает domain.ErrNotFound, если ссылка с таким кодом не найдена.
func (r *LinkRepository) GetByShortCode(ctx context.Context, code string) (*domain.Link, error) {
	sql, args, err := r.pg.Builder.
		Select("id", "short_code", "long_url", "created_at").
		From("links").
		Where("short_code = ?", code).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get link sql: %w", err)
	}

	link := &domain.Link{}

	err = r.pg.QueryRow(ctx, sql, args...).Scan(
		&link.ID,
		&link.ShortCode,
		&link.LongURL,
		&link.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("get link by short code: %w", err)
	}

	return link, nil
}

// ShortCodeExists проверяет, занят ли уже данный короткий код.
func (r *LinkRepository) ShortCodeExists(ctx context.Context, code string) (bool, error) {
	sql, args, err := r.pg.Builder.
		Select("1").
		From("links").
		Where("short_code = ?", code).
		Limit(1).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("build exists sql: %w", err)
	}

	var dummy int

	err = r.pg.QueryRow(ctx, sql, args...).Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("check short code exists: %w", err)
	}

	return true, nil
}
