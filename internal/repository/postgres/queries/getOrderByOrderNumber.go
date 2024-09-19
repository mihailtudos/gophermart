package queries

const GetOrderByOrderNumber = `
	SELECT order_number, user_id, order_status 
	FROM orders 
	WHERE order_number = $1
`
