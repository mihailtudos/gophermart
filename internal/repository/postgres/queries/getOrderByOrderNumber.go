package queries

// TODO - group queries in files by entities and rename files into snake case
const GetOrderByOrderNumber = `
	SELECT order_number, user_id, order_status 
	FROM orders 
	WHERE order_number = $1
`
