-- migrate:up
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    uid UUID NOT NULL UNIQUE,
    email_encrypted TEXT NOT NULL,
    email_lookup_hash TEXT NOT NULL UNIQUE,
    name_encrypted TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by BIGINT NOT NULL,
    deleted_at TIMESTAMPTZ,
    deleted_by TEXT
);

CREATE INDEX idx_users_uid ON users(uid);
CREATE INDEX idx_users_email_lookup ON users(email_lookup_hash);
CREATE INDEX idx_users_username ON users(LOWER(username));

-- migrate:down
DROP TABLE IF EXISTS users;
