-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER update_orders_updated_at
BEFORE UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
-- +goose StatementEnd
