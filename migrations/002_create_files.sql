create table if not exists files (
	id uuid primary key,
	owner_id uuid not null references users(id) on delete cascade,
	original_name text not null,
	storage_path text not null,
	content_type text not null,
	size_bytes bigint not null check (size_bytes > 0),
	checksum_sha256 text,
	checksum_status text not null default 'pending',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create index if not exists files_owner_id_created_at_idx on files (owner_id, created_at desc);
