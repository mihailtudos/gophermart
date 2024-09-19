package queries

const UpdateOrderStatusAndAccrualPoints = `
	UPDATE orders
	SET
		status $1
		accrual $2
	WHERE
		order = $3
`