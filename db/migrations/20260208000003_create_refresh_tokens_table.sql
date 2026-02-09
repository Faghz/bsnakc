-- migrate:up
CREATE TABLE IF NOT EXISTS refresh_tokens (
    uid UUID PRIMARY KEY,
    user_id BIGINT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    ip_address VARCHAR(45),
    user_agent TEXT,

    CONSTRAINT fk_refresh_tokens_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Index for fast user lookup (for revoke-all operations)
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id) WHERE revoked_at IS NULL;

-- Index for cleanup of expired tokens
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;

-- Index for finding active tokens
CREATE INDEX idx_refresh_tokens_active ON refresh_tokens(revoked_at, expires_at);

-- migrate:down
DROP TABLE IF EXISTS refresh_tokens;
