-- +goose Up
-- +goose StatementBegin
BEGIN;
CREATE TABLE withdrawals (
    id SERIAL PRIMARY KEY,
    order_number TEXT NOT NULL,
    sum INT NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_withdrawals_user_id_processed_at ON withdrawals (user_id, processed_at DESC);
COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
BEGIN;
DROP INDEX idx_withdrawals_user_id_processed_at;
DROP TABLE withdrawals;
COMMIT;
-- +goose StatementEnd
