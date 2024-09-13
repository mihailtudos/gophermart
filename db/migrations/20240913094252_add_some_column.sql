-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    number TEXT UNIQUE,
    status TEXT NOT NULL,
    accrual int DEFAULT 0,
    updated_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
