-- migrate:up
CREATE TABLE IF NOT EXISTS identities (
    id BIGSERIAL PRIMARY KEY,
    uid UUID NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    provider VARCHAR(50) NOT NULL,
    provider_id_encrypted TEXT,
    provider_id_lookup_hash TEXT,
    password_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT NOT NULL,
    deleted_at TIMESTAMPTZ,
    deleted_by TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (provider, provider_id_lookup_hash)
);

COMMENT ON COLUMN identities.id IS 'Internal primary key';
COMMENT ON COLUMN identities.uid IS 'External UUID identifier for API responses';
COMMENT ON COLUMN identities.user_id IS 'Reference to the associated user';
COMMENT ON COLUMN identities.provider IS 'Authentication provider type (local, discord, etc.)';
COMMENT ON COLUMN identities.provider_id_encrypted IS 'AES-256-GCM encrypted provider-specific user identifier';
COMMENT ON COLUMN identities.provider_id_lookup_hash IS 'HMAC-SHA256 hash for searchable provider ID lookups';
COMMENT ON COLUMN identities.password_hash IS 'Argon2id password hash for local authentication';
COMMENT ON COLUMN identities.created_at IS 'Record creation timestamp';
COMMENT ON COLUMN identities.created_by IS 'Identifier of the user who created this record';
COMMENT ON COLUMN identities.updated_at IS 'Last update timestamp';
COMMENT ON COLUMN identities.updated_by IS 'Identifier of the user who last updated this record';
COMMENT ON COLUMN identities.deleted_at IS 'Soft delete timestamp';
COMMENT ON COLUMN identities.deleted_by IS 'Identifier of the user who deleted this record';

CREATE INDEX idx_identities_uid ON identities(uid);
CREATE INDEX idx_identities_user_id ON identities(user_id);
CREATE INDEX idx_identities_provider_lookup ON identities(provider, provider_id_lookup_hash);

-- migrate:down
DROP TABLE IF EXISTS identities;
