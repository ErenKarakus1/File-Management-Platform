package repository

const createFileQuery = `
	insert into files (
		id,
		owner_id,
		original_name,
		storage_path,
		content_type,
		size_bytes,
		status
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
		status,
		processed_at,
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
		status,
		processed_at,
		created_at,
		updated_at
	from files
	where owner_id = $1
	order by created_at desc
`

const getFileByIDAndOwnerQuery = `
	select
		id,
		owner_id,
		original_name,
		storage_path,
		content_type,
		size_bytes,
		checksum_sha256,
		status,
		processed_at,
		created_at,
		updated_at
	from files
	where id = $1 and owner_id = $2
`

const claimPendingFileQuery = `
	update files
	set status = 'processing',
		updated_at = now()
	where id = (
		select id
		from files
		where status = 'pending'
		order by created_at
		for update skip locked
		limit 1
	)
	returning
		id,
		owner_id,
		original_name,
		storage_path,
		content_type,
		size_bytes,
		checksum_sha256,
		status,
		processed_at,
		created_at,
		updated_at
`

const markFileReadyQuery = `
	update files
	set status = 'ready',
		checksum_sha256 = $2,
		processed_at = now(),
		updated_at = now()
	where id = $1
`

const markFileFailedQuery = `
	update files
	set status = 'failed',
		processed_at = now(),
		updated_at = now()
	where id = $1
`
