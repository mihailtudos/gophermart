-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS session_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    device_info TEXT,
    ip_address TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP IF EXISTS session_tokens;
-- +goose StatementEnd
