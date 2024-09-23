package queries

// GetUserBalance is used to get the balance of an user by user id
const GetUserBalance = `
	SELECT current, withdrawn
	FROM user_loyalty_points
	WHERE user_id = $1
`

// UpdateUserLoyaltyPoints is used to update users balance when new points are received
const UpdateUserLoyaltyPoints = `
	UPDATE user_loyalty_points
	SET 
		current = current + $1
	WHERE 
		user_id = $2
`

// UpdateUserBalance is used to update the balance of an user
const UpdateUserBalance = `
	UPDATE user_loyalty_points
	SET 
		current = $1,
		withdrawn = $2
	WHERE 
		user_id = $3;
`
