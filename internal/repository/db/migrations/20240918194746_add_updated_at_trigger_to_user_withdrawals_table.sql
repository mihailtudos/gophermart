-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER update_user_withdrawals_updated_at
BEFORE UPDATE ON user_withdrawals
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_user_withdrawals_updated_at ON user_withdrawals;
-- +goose StatementEnd
