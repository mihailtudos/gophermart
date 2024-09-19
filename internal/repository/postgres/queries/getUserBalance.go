package queries

const GetUserBalanceStmt = `
	SELECT current, withdrawn
	FROM user_loyalty_points
	WHERE user_id = $1
`
