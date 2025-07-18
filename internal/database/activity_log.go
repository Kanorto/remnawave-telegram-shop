package database

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ActivityLogRepository struct {
	pool *pgxpool.Pool
}

func NewActivityLogRepository(pool *pgxpool.Pool) *ActivityLogRepository {
	return &ActivityLogRepository{pool: pool}
}

func (r *ActivityLogRepository) Log(ctx context.Context, telegramID int64, action, content string) error {
	query := sq.Insert("activity_log").Columns("telegram_id", "action", "content").Values(telegramID, action, content).PlaceholderFormat(sq.Dollar)
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}
