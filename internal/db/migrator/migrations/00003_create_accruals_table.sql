-- +goose Up
-- +goose StatementBegin
BEGIN;
CREATE TABLE accruals (
    id SERIAL PRIMARY KEY,
    accrual INT NOT NULL,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE
);

CREATE INDEX idx_accruals_order_id ON accruals (order_id);
COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
BEGIN;
DROP INDEX idx_accruals_order_id;
DROP TABLE accruals;
COMMIT;
-- +goose StatementEnd
