-- +goose Up
-- +goose StatementBegin
BEGIN;
CREATE TABLE accruals (
    id SERIAL PRIMARY KEY,
    accrual INT NOT NULL,
    order_id INT UNIQUE NOT NULL REFERENCES orders(id)
);

CREATE INDEX idx_accruals_order_id ON accruals (order_id);
COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS accruals;
-- +goose StatementEnd
