package queries

const UpdateUserBalance = `
	UPDATE user_loyalty_points
	SET 
		current = $1,
		withdrawn = $2
	WHERE 
		user_id = $3;
`
