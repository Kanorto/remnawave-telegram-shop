CREATE TABLE activity_log (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL REFERENCES customer(telegram_id),
    action VARCHAR(20),
    content TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_activity_log_telegram_id ON activity_log USING hash (telegram_id);
