package database

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
)

type BannedUserRepository struct {
	pool *pgxpool.Pool
}

func NewBannedUserRepository(pool *pgxpool.Pool) *BannedUserRepository {
	return &BannedUserRepository{pool: pool}
}

func (r *BannedUserRepository) IsBanned(ctx context.Context, telegramID int64) (bool, error) {
	query := sq.Select("1").From("banned_user").Where(sq.Eq{"telegram_id": telegramID}).PlaceholderFormat(sq.Dollar)
	sql, args, err := query.ToSql()
	if err != nil {
		return false, err
	}
	var tmp int
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&tmp)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (r *BannedUserRepository) Ban(ctx context.Context, telegramID int64) error {
	query := sq.Insert("banned_user").Columns("telegram_id").Values(telegramID).Suffix("ON CONFLICT DO NOTHING").PlaceholderFormat(sq.Dollar)
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *BannedUserRepository) Unban(ctx context.Context, telegramID int64) error {
	query := sq.Delete("banned_user").Where(sq.Eq{"telegram_id": telegramID}).PlaceholderFormat(sq.Dollar)
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *BannedUserRepository) Count(ctx context.Context) (int, error) {
	query := sq.Select("count(*)").From("banned_user").PlaceholderFormat(sq.Dollar)
	sql, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}
	var cnt int
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&cnt)
	return cnt, err
}
