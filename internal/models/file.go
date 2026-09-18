package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID             uuid.UUID `json:"id"`
	OwnerID        uuid.UUID `json:"owner_id"`
	OriginalName   string    `json:"original_name"`
	StoragePath    string    `json:"-"`
	ContentType    string    `json:"content_type"`
	SizeBytes      int64     `json:"size_bytes"`
	ChecksumSHA256 *string   `json:"checksum_sha256"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
