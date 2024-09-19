package queries

const CreateWithdrawalPointsRecord = `
		INSERT INTO user_withdrawals(user_id, order_number, sum)
		VALUES($1,$2,$3)
		RETURNING id
	`
