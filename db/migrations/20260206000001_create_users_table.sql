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
    created_by TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT NOT NULL,
    deleted_at TIMESTAMPTZ,
    deleted_by TEXT
);

COMMENT ON COLUMN users.id IS 'Internal primary key';
COMMENT ON COLUMN users.uid IS 'External UUID identifier for API responses';
COMMENT ON COLUMN users.email_encrypted IS 'AES-256-GCM encrypted email address';
COMMENT ON COLUMN users.email_lookup_hash IS 'HMAC-SHA256 hash for searchable email lookups';
COMMENT ON COLUMN users.name_encrypted IS 'AES-256-GCM encrypted user display name';
COMMENT ON COLUMN users.username IS 'Unique username identifier';
COMMENT ON COLUMN users.avatar_url IS 'URL to user avatar image';
COMMENT ON COLUMN users.created_at IS 'Record creation timestamp';
COMMENT ON COLUMN users.created_by IS 'Identifier of the user who created this record';
COMMENT ON COLUMN users.updated_at IS 'Last update timestamp';
COMMENT ON COLUMN users.updated_by IS 'Identifier of the user who last updated this record';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp';
COMMENT ON COLUMN users.deleted_by IS 'Identifier of the user who deleted this record';

CREATE INDEX idx_users_uid ON users(uid);
CREATE INDEX idx_users_email_lookup ON users(email_lookup_hash);
CREATE INDEX idx_users_username ON users(LOWER(username));

-- migrate:down
DROP TABLE IF EXISTS users;
