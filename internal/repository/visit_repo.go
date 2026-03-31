package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"

	"github.com/ponchik327/Shortener/internal/domain"
)

// VisitRepository выполняет операции с таблицей visits в базе данных.
type VisitRepository struct {
	pg *pgxdriver.Postgres
}

// NewVisitRepository создаёт новый VisitRepository.
func NewVisitRepository(pg *pgxdriver.Postgres) *VisitRepository {
	return &VisitRepository{pg: pg}
}

// InsertVisit сохраняет единичный переход по ссылке.
func (r *VisitRepository) InsertVisit(ctx context.Context, visit *domain.Visit) error {
	sql, args, err := r.pg.Builder.
		Insert("visits").
		Columns("link_id", "user_agent", "visited_at").
		Values(visit.LinkID, visit.UserAgent, visit.VisitedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert visit sql: %w", err)
	}

	if _, err = r.pg.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("insert visit: %w", err)
	}

	return nil
}

// GetAnalytics возвращает агрегированную статистику переходов для указанного link ID.
func (r *VisitRepository) GetAnalytics(ctx context.Context, linkID int64) (*domain.Analytics, error) {
	analytics := &domain.Analytics{}

	total, err := r.queryTotal(ctx, linkID)
	if err != nil {
		return nil, err
	}

	analytics.TotalVisits = total

	byDay, err := r.queryByDay(ctx, linkID)
	if err != nil {
		return nil, err
	}

	analytics.ByDay = byDay

	byMonth, err := r.queryByMonth(ctx, linkID)
	if err != nil {
		return nil, err
	}

	analytics.ByMonth = byMonth

	byUA, err := r.queryByUserAgent(ctx, linkID)
	if err != nil {
		return nil, err
	}

	analytics.ByUserAgent = byUA

	return analytics, nil
}

func (r *VisitRepository) queryTotal(ctx context.Context, linkID int64) (int64, error) {
	sql, args, err := r.pg.Builder.
		Select("COUNT(*)").
		From("visits").
		Where("link_id = ?", linkID).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("build total visits sql: %w", err)
	}

	var total int64

	if err = r.pg.QueryRow(ctx, sql, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("query total visits: %w", err)
	}

	return total, nil
}

func (r *VisitRepository) queryByDay(ctx context.Context, linkID int64) ([]domain.DayCount, error) {
	sql, args, err := r.pg.Builder.
		Select("TO_CHAR(visited_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day", "COUNT(*) AS cnt").
		From("visits").
		Where("link_id = ?", linkID).
		GroupBy("day").
		OrderBy("day").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build visits by day sql: %w", err)
	}

	return queryAggregation(ctx, r, sql, args, func(rows pgx.Rows) (domain.DayCount, error) {
		var dc domain.DayCount
		if err := rows.Scan(&dc.Day, &dc.Count); err != nil {
			return dc, err
		}
		return dc, nil
	})
}

func (r *VisitRepository) queryByMonth(ctx context.Context, linkID int64) ([]domain.MonthCount, error) {
	sql, args, err := r.pg.Builder.
		Select("TO_CHAR(visited_at AT TIME ZONE 'UTC', 'YYYY-MM') AS month", "COUNT(*) AS cnt").
		From("visits").
		Where("link_id = ?", linkID).
		GroupBy("month").
		OrderBy("month").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build visits by month sql: %w", err)
	}

	return queryAggregation(ctx, r, sql, args, func(rows pgx.Rows) (domain.MonthCount, error) {
		var mc domain.MonthCount
		if err := rows.Scan(&mc.Month, &mc.Count); err != nil {
			return mc, err
		}
		return mc, nil
	})
}

func (r *VisitRepository) queryByUserAgent(ctx context.Context, linkID int64) ([]domain.UserAgentCount, error) {
	sql, args, err := r.pg.Builder.
		Select("user_agent", "COUNT(*) AS cnt").
		From("visits").
		Where("link_id = ?", linkID).
		GroupBy("user_agent").
		OrderBy("cnt DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build visits by user_agent sql: %w", err)
	}

	return queryAggregation(ctx, r, sql, args, func(rows pgx.Rows) (domain.UserAgentCount, error) {
		var uac domain.UserAgentCount
		if err := rows.Scan(&uac.UserAgent, &uac.Count); err != nil {
			return uac, err
		}
		return uac, nil
	})
}

func queryAggregation[T any](
	ctx context.Context,
	r *VisitRepository,
	sql string,
	args []any,
	scanRow func(pgx.Rows) (T, error),
) ([]T, error) {
	rows, err := r.pg.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query aggregation: %w", err)
	}
	defer rows.Close()

	var result []T
	for rows.Next() {
		item, err := scanRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return result, nil
}
