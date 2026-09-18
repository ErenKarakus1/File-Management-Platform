package repository

const createUserQuery = `
	insert into users (id, name, email, password_hash)
	values ($1, $2, $3, $4)
	returning id, name, email, password_hash, created_at, updated_at
`

const getUserByEmailQuery = `
	select id, name, email, password_hash, created_at, updated_at
	from users
	where email = $1
`

const getUserByIDQuery = `
	select id, name, email, password_hash, created_at, updated_at
	from users
	where id = $1
`
