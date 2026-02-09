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
    created_by BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by BIGINT NOT NULL,
    deleted_at TIMESTAMPTZ,
    deleted_by TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (provider, provider_id_lookup_hash)
);

CREATE INDEX idx_identities_uid ON identities(uid);
CREATE INDEX idx_identities_user_id ON identities(user_id);
CREATE INDEX idx_identities_provider_lookup ON identities(provider, provider_id_lookup_hash);

-- migrate:down
DROP TABLE IF EXISTS identities;
