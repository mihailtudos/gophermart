package queries

const InsertNewUser = `
	INSERT INTO users (login, password_hash)
	VALUES($1, $2)
	RETURNING id
`

const GetUserByLogin = `
	SELECT id, login, password_hash, created_at, version
		FROM users
		WHERE login = $1
`

const GetUserByID = `
	SELECT id, login, version, created_at, updated_at
		FROM users
		WHERE id = $1
`
