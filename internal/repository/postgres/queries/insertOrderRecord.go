package queries

const InsertOrderRecord = `
INSERT INTO orders(user_id, order_number, order_status)
VALUES($1, $2, $3)
RETURNING order_number, user_id, order_status, accrual, created_at, updated_at
`
