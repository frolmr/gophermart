-- +goose Up
-- +goose StatementBegin
BEGIN;

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    number TEXT UNIQUE NOT NULL,
    status order_status NOT NULL DEFAULT 'NEW',
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id INT NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_orders_user_id_uploaded_at ON orders (user_id, uploaded_at DESC);

COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
