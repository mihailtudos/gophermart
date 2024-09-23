package queries

// GetOrderByOrderNumber is used to retrive an order by order number
const GetOrderByOrderNumber = `
	SELECT order_number, user_id, order_status 
	FROM orders 
	WHERE order_number = $1
`

// GetUnfinishedOrders is used to retrive all unfinished (new, processing) orders
const GetUnfinishedOrders = `
	SELECT order_number, order_status, accrual
	FROM 
		orders
	WHERE 
		order_status IN ($1, $2)
`

// GetUserOrders is used to retrieve all orders by user_id
const GetUserOrders = `
	SELECT order_number, created_at, order_status, accrual
	FROM orders
	WHERE user_id = $1
	ORDER BY created_at DESC
`

// InsertOrderRecord is used to insert new order records in the orders table
const InsertOrderRecord = `
	INSERT INTO orders(user_id, order_number, order_status)
	VALUES($1, $2, $3)
	RETURNING 
		order_number, user_id, order_status, accrual, created_at, updated_at
`

// UpdateOrderStatusAndAccrualPoints is used to update the status and the points of an order
const UpdateOrderStatusAndAccrualPoints = `
	UPDATE orders
	SET
		order_status = $1,
		accrual = $2
	WHERE
		order_number = $3
	RETURNING user_id
`
