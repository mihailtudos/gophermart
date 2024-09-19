package queries

const GetUnfinishedOrders = `
	SELECT order_number, order_status, accrual
	FROM 
		orders
	WHERE 
		order_status IN ($1, $2)
`
