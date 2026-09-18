package repository

const createFileQuery = `
	insert into files (
		id,
		owner_id,
		original_name,
		storage_path,
		content_type,
		size_bytes,
		checksum_status
	)
	values ($1, $2, $3, $4, $5, $6, $7)
	returning
		id,
		owner_id,
		original_name,
		storage_path,
		content_type,
		size_bytes,
		checksum_sha256,
		checksum_status,
		created_at,
		updated_at
`

const listFilesByOwnerQuery = `
	select
		id,
		owner_id,
		original_name,
		storage_path,
		content_type,
		size_bytes,
		checksum_sha256,
		checksum_status,
		created_at,
		updated_at
	from files
	where owner_id = $1
	order by created_at desc
`
