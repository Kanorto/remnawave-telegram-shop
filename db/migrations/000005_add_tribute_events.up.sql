CREATE TABLE tribute_event (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL,
    subscription_name TEXT NOT NULL,
    subscription_id BIGINT NOT NULL,
    period_id BIGINT NOT NULL,
    period VARCHAR(20) NOT NULL,
    price BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL,
    user_id BIGINT NOT NULL,
    telegram_user_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    channel_name TEXT NOT NULL,
    cancel_reason TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);
