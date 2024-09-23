package queries

// DeleteUserSession is used to delete an existing session
const DeleteUserSession = `
	DELETE FROM session_tokens 
		WHERE user_id = $1
`

// DeleteUserSession is used to create a new user session
const CreateNewUserSession = `
	INSERT INTO session_tokens (user_id, token, expires_at)
	VALUES ($1, $2, $3)
`
