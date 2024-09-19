package queries

const GetUserOrders = `
		SELECT order_number, created_at, order_status, accrual
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
