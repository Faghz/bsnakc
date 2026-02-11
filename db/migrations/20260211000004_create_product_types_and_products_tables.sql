-- migrate:up
CREATE TABLE IF NOT EXISTS product_types (
    id BIGSERIAL PRIMARY KEY,
    uid UUID NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT NOT NULL,
    deleted_at TIMESTAMPTZ,
    deleted_by TEXT
);

COMMENT ON COLUMN product_types.id IS 'Internal primary key';
COMMENT ON COLUMN product_types.uid IS 'External UUID identifier for API responses';
COMMENT ON COLUMN product_types.name IS 'Human-readable name of the product type';
COMMENT ON COLUMN product_types.created_at IS 'Record creation timestamp';
COMMENT ON COLUMN product_types.created_by IS 'Identifier of the user who created this record';
COMMENT ON COLUMN product_types.updated_at IS 'Last update timestamp';
COMMENT ON COLUMN product_types.updated_by IS 'Identifier of the user who last updated this record';
COMMENT ON COLUMN product_types.deleted_at IS 'Soft delete timestamp';
COMMENT ON COLUMN product_types.deleted_by IS 'Identifier of the user who deleted this record';

CREATE INDEX idx_product_types_uid ON product_types(uid);

CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    uid UUID NOT NULL UNIQUE,
    name TEXT NOT NULL,
    product_type_id BIGINT NOT NULL REFERENCES product_types(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT NOT NULL,
    deleted_at TIMESTAMPTZ,
    deleted_by TEXT
);

COMMENT ON COLUMN products.id IS 'Internal primary key';
COMMENT ON COLUMN products.uid IS 'External UUID identifier for API responses';
COMMENT ON COLUMN products.name IS 'Human-readable name of the product';
COMMENT ON COLUMN products.product_type_id IS 'Reference to the product type this product belongs to';
COMMENT ON COLUMN products.created_at IS 'Record creation timestamp';
COMMENT ON COLUMN products.created_by IS 'Identifier of the user who created this record';
COMMENT ON COLUMN products.updated_at IS 'Last update timestamp';
COMMENT ON COLUMN products.updated_by IS 'Identifier of the user who last updated this record';
COMMENT ON COLUMN products.deleted_at IS 'Soft delete timestamp';
COMMENT ON COLUMN products.deleted_by IS 'Identifier of the user who deleted this record';

CREATE INDEX idx_products_uid ON products(uid);
CREATE INDEX idx_products_product_type_id ON products(product_type_id);
-- migrate:down
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS product_types;
