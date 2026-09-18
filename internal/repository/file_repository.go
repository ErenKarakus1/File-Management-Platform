package repository

import (
	"context"

	"github.com/ErenKarakus1/File-Management-Platform/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepository struct {
	db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file models.File) (models.File, error) {
	err := r.db.QueryRow(
		ctx,
		createFileQuery,
		file.ID,
		file.OwnerID,
		file.OriginalName,
		file.StoragePath,
		file.ContentType,
		file.SizeBytes,
		file.Status,
	).Scan(
		&file.ID,
		&file.OwnerID,
		&file.OriginalName,
		&file.StoragePath,
		&file.ContentType,
		&file.SizeBytes,
		&file.ChecksumSHA256,
		&file.Status,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
	return file, err
}

func (r *FileRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]models.File, error) {
	rows, err := r.db.Query(ctx, listFilesByOwnerQuery, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]models.File, 0)
	for rows.Next() {
		var file models.File
		if err := rows.Scan(
			&file.ID,
			&file.OwnerID,
			&file.OriginalName,
			&file.StoragePath,
			&file.ContentType,
			&file.SizeBytes,
			&file.ChecksumSHA256,
			&file.Status,
			&file.CreatedAt,
			&file.UpdatedAt,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}
