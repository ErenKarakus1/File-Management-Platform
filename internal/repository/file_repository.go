package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/File-Management-Platform/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPendingFileNotFound = errors.New("pending file not found")

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
		&file.ProcessedAt,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
	return file, err
}

func (r *FileRepository) ClaimPending(ctx context.Context) (models.File, error) {
	var file models.File
	err := scanFile(r.db.QueryRow(ctx, claimPendingFileQuery), &file)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.File{}, ErrPendingFileNotFound
		}
		return models.File{}, err
	}
	return file, nil
}

func (r *FileRepository) MarkReady(ctx context.Context, id uuid.UUID, checksumSHA256 string) error {
	_, err := r.db.Exec(ctx, markFileReadyQuery, id, checksumSHA256)
	return err
}

func (r *FileRepository) MarkFailed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, markFileFailedQuery, id)
	return err
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
		if err := scanFile(rows, &file); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

type fileScanner interface {
	Scan(dest ...any) error
}

func scanFile(scanner fileScanner, file *models.File) error {
	return scanner.Scan(
		&file.ID,
		&file.OwnerID,
		&file.OriginalName,
		&file.StoragePath,
		&file.ContentType,
		&file.SizeBytes,
		&file.ChecksumSHA256,
		&file.Status,
		&file.ProcessedAt,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
}
