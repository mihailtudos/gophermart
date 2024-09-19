package queries


const GetUnfinishedOrders = `
	SELECT number, status, accrual
	FROM orders
	WHERE status IN ($1, $2)
`