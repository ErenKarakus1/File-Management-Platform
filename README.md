# File Management Platform

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-008ECF?style=for-the-badge&logo=gin&logoColor=white)
![React](https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

An authenticated file management platform built with Go, Gin, React, PostgreSQL, Redis, and Docker. Users can upload files to disk, track processing status, download verified files, and delete files asynchronously through background workers.

## Architecture

```mermaid
flowchart LR
    Client[React Client] --> Web[Nginx Web Container]
    Web --> API[Gin API]

    API --> Postgres[(PostgreSQL)]
    API --> Redis[(Redis)]
    API --> Storage[(Disk Storage)]

    Workers[File Workers] --> Postgres
    Workers --> Storage

    API --> Workers
```

## Project Structure

```text
file-management-platform/
|-- cmd/
|   `-- api/
|-- internal/
|   |-- auth/
|   |-- config/
|   |-- database/
|   |-- files/
|   |-- models/
|   |-- password/
|   |-- ratelimit/
|   |-- repository/
|   |-- server/
|   |-- validation/
|   `-- workers/
|-- migrations/
|-- web/
|-- Dockerfile
`-- docker-compose.yml
```

## Tech Stack

### Backend

* Go
* Gin
* pgx
* JWT
* bcrypt

### Frontend

* React
* Vite
* lucide-react
* Nginx for Docker serving

### Infrastructure

* PostgreSQL
* Redis
* Docker Compose

---

## Features

* JWT authentication
* Password hashing with bcrypt
* Authenticated file uploads
* Frontend-only batch uploads
* Disk-backed file storage
* PostgreSQL metadata storage
* SHA-256 checksum processing
* Background processing workers
* Separate delete workers
* Authenticated file downloads
* Async file deletion
* Redis-backed rate limiting
* Startup database migrations
* React file manager UI
* Dockerized full stack
* Atomic PostgreSQL job claiming with `FOR UPDATE SKIP LOCKED`

---

## File Lifecycle

```mermaid
stateDiagram-v2
    [*] --> pending
    pending --> processing
    processing --> ready
    processing --> failed

    pending --> deleting
    processing --> deleting
    ready --> deleting
    failed --> deleting

    deleting --> [*]
```

### Statuses

* `pending`: file metadata exists and is waiting for a processing worker
* `processing`: worker is reading the disk file and calculating checksum
* `ready`: checksum was calculated and the file can be downloaded
* `failed`: processing failed or the disk file is missing
* `deleting`: delete was requested and a delete worker will remove the file

---

## Worker Flow

```mermaid
sequenceDiagram
    participant User
    participant API
    participant DB as PostgreSQL
    participant Disk
    participant Worker

    User->>API: Upload file
    API->>Disk: Save file bytes
    API->>DB: Insert metadata with status=pending
    API-->>User: Return file metadata

    Worker->>DB: Claim pending file
    Worker->>Disk: Read file
    Worker->>Worker: Calculate SHA-256
    Worker->>DB: status=ready, checksum_sha256=value
```

Delete workers run separately from processing workers so delete jobs and processing jobs do not starve each other.

---

## Authentication

Protected endpoints require:

```http
Authorization: Bearer <jwt_token>
```

The React client stores the access token in `localStorage` for local development convenience.

---

## Getting Started

### 1. Prerequisites

* Go 1.26+
* Node.js 22+
* Docker and Docker Compose

For non-Docker local development, PostgreSQL and Redis must also be available locally.

### 2. Run with Docker

```bash
docker compose up --build
```

Services:

```text
Frontend: http://localhost:5173
API:      http://localhost:8080
```

PostgreSQL, Redis, API, and web containers are started together.

### 3. Run locally without Docker

Create `.env` from `.env.example`, then make sure PostgreSQL and Redis are running locally.

```bash
go run ./cmd/api
```

In another terminal:

```bash
cd web
npm install
npm run dev
```

Local frontend:

```text
http://127.0.0.1:5173
```

---

## Configuration

Main environment variables:

```env
HTTP_ADDR=:8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/file_management?sslmode=disable
JWT_SECRET=change-me-local-dev-secret
TOKEN_TTL=24h
FILE_STORAGE_DIR=storage
MAX_UPLOAD_BYTES=52428800
FILE_WORKER_COUNT=2
FILE_DELETE_WORKER_COUNT=1
FILE_WORKER_POLL=1s
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
RATE_LIMIT=120
RATE_LIMIT_WINDOW=1m
```

`MAX_UPLOAD_BYTES=52428800` is 50 MiB.

---

## Public Endpoints

### Authentication

```http
POST /api/v1/auth/register
POST /api/v1/auth/login
```

### Health

```http
GET /healthz
```

---

## Protected Endpoints

### Auth

```http
GET /api/v1/auth/me
```

### Files

```http
POST   /api/v1/files
GET    /api/v1/files
GET    /api/v1/files/:id
GET    /api/v1/files/:id/download
DELETE /api/v1/files/:id
```

Uploads use multipart form data with field name:

```text
file
```

---

## Database

The API runs embedded SQL migrations on startup using:

```text
internal/database/migrations/
```

The root `migrations/` directory keeps the plain SQL migration files visible for inspection.

---

## Rate Limiting

Redis-backed fixed-window rate limiting is applied globally by client IP.

Default:

```env
RATE_LIMIT=120
RATE_LIMIT_WINDOW=1m
```

Responses include:

```http
X-RateLimit-Limit
X-RateLimit-Remaining
X-RateLimit-Reset
Retry-After
```

---

## Security Features

* bcrypt password hashing
* JWT authentication
* Owner-scoped file access
* Server-generated storage paths
* Uploaded file names stored only as metadata
* Content type detected from file bytes
* Backend-enforced max upload size
* Redis-backed rate limiting
* Storage paths hidden from API JSON responses

---

## Tests

Run backend tests:

```bash
go test ./...
```

Build frontend:

```bash
cd web
npm run build
```

---

## Future Improvements

* Failure reason for failed files
* Duplicate file detection by checksum
* Integrity re-check jobs
* Pagination for file lists
* More integration tests
* Per-route rate limits
* Refresh tokens
* Production object storage support
* Observability and metrics
