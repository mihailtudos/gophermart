-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER update_user_loyalty_points_updated_at
BEFORE UPDATE ON user_loyalty_points
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_user_loyalty_points_updated_at ON user_loyalty_points;
-- +goose StatementEnd
