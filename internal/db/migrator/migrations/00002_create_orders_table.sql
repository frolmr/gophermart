-- +goose Up
-- +goose StatementBegin
BEGIN;

CREATE TYPE order_status AS ENUM ('NEW', 'REGISTERED', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    number TEXT UNIQUE NOT NULL,
    status order_status NOT NULL DEFAULT 'NEW',
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- CREATE INDEX idx_orders_user_id ON orders (user_id);
-- CREATE INDEX idx_orders_uploaded_at ON orders (uploaded_at DESC);
CREATE INDEX idx_orders_user_id_uploaded_at ON orders (user_id, uploaded_at DESC);

COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
BEGIN;
-- DROP INDEX idx_orders_user_id;
-- DROP INDEX idx_orders_uploaded_at;
DROP INDEX idx_orders_user_id_uploaded_at;
DROP TABLE orders;
DROP TYPE order_status;
COMMIT;
-- +goose StatementEnd
