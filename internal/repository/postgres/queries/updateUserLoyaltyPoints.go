package queries

const UpdateUserLoyaltyPoints = `
	UPDATE user_loyalty_points
	SET 
		current = current + $1
	WHERE 
		user_id = $2
`
