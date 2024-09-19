package queries

const UpdateOrderStatusAndAccrualPoints = `
	UPDATE orders
	SET
		order_status = $1,
		accrual = $2
	WHERE
		order_number = $3
	RETURNING user_id
`
