package queries

const GetUserWithdrawals = `
	SELECT order_number, sum, created_at
	FROM user_withdrawals
	WHERE user_id = $1
	ORDER BY created_at ASC
`
