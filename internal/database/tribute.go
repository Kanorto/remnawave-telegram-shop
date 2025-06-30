package database

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type TributeEvent struct {
	ID               int64     `db:"id"`
	Name             string    `db:"name"`
	CreatedAt        time.Time `db:"created_at"`
	SentAt           time.Time `db:"sent_at"`
	SubscriptionName string    `db:"subscription_name"`
	SubscriptionID   int64     `db:"subscription_id"`
	PeriodID         int64     `db:"period_id"`
	Period           string    `db:"period"`
	Price            int64     `db:"price"`
	Amount           int64     `db:"amount"`
	Currency         string    `db:"currency"`
	UserID           int64     `db:"user_id"`
	TelegramUserID   int64     `db:"telegram_user_id"`
	ChannelID        int64     `db:"channel_id"`
	ChannelName      string    `db:"channel_name"`
	CancelReason     *string   `db:"cancel_reason"`
	ExpiresAt        time.Time `db:"expires_at"`
}

type TributeRepository struct {
	pool *pgxpool.Pool
}

func NewTributeRepository(pool *pgxpool.Pool) *TributeRepository {
	return &TributeRepository{pool: pool}
}

func (r *TributeRepository) Insert(ctx context.Context, ev *TributeEvent) error {
	builder := sq.Insert("tribute_event").
		Columns(
			"name", "created_at", "sent_at", "subscription_name", "subscription_id",
			"period_id", "period", "price", "amount", "currency", "user_id",
			"telegram_user_id", "channel_id", "channel_name", "cancel_reason", "expires_at",
		).
		Values(
			ev.Name, ev.CreatedAt, ev.SentAt, ev.SubscriptionName, ev.SubscriptionID,
			ev.PeriodID, ev.Period, ev.Price, ev.Amount, ev.Currency, ev.UserID,
			ev.TelegramUserID, ev.ChannelID, ev.ChannelName, ev.CancelReason, ev.ExpiresAt,
		).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}
