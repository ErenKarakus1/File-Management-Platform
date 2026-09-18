# File Management Platform

Gin-based API for authenticated file management.

## Configuration

Environment variables:

- `HTTP_ADDR`, default `:8080`
- `DATABASE_URL`, default `postgres://postgres:postgres@localhost:5432/file_management?sslmode=disable`
- `JWT_SECRET`, required for non-local usage
- `TOKEN_TTL`, default `24h`

## Database

Apply migrations manually for now:

```sh
psql "$DATABASE_URL" -f migrations/001_create_users.sql
```

## Run

```sh
go run ./cmd/api
```

## Auth endpoints

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`

Use `Authorization: Bearer <access_token>` for protected endpoints.
