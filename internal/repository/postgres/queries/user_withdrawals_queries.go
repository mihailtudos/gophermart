package queries

// CreateWithdrawalPointsRecord is used to create new withdrawal record
const CreateWithdrawalPointsRecord = `
		INSERT INTO user_withdrawals(user_id, order_number, sum)
		VALUES($1,$2,$3)
		RETURNING id
	`

// GetUserWithdrawals is used to get all user withdrawals records
const GetUserWithdrawals = `
	SELECT order_number, sum, created_at
	FROM user_withdrawals
	WHERE user_id = $1
	ORDER BY created_at ASC
`
