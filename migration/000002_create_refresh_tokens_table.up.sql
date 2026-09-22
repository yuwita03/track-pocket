CREATE TABLE IF NOT EXISTS refresh_token(
    id UUID         PRIMARY KEY,
    user_id UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL,
    revoked         BOOLEAN NOT NULL DEFAULT FALSE,
    expired_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
)

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_token(user_id);