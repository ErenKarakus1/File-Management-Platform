package files

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ErenKarakus1/File-Management-Platform/internal/models"
	"github.com/ErenKarakus1/File-Management-Platform/internal/repository"
	"github.com/google/uuid"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusReady      = "ready"
	StatusFailed     = "failed"
)

var ErrEmptyFile = errors.New("file is empty")
var ErrFileNotReady = errors.New("file is not ready")

type Service struct {
	files      *repository.FileRepository
	storageDir string
}

func NewService(files *repository.FileRepository, storageDir string) *Service {
	return &Service{
		files:      files,
		storageDir: storageDir,
	}
}

func (s *Service) Upload(ctx context.Context, ownerID uuid.UUID, header *multipart.FileHeader) (models.File, error) {
	if header.Size <= 0 {
		return models.File{}, ErrEmptyFile
	}

	src, err := header.Open()
	if err != nil {
		return models.File{}, err
	}
	defer src.Close()

	fileID := uuid.New()
	ownerDir := filepath.Join(s.storageDir, ownerID.String())
	if err := os.MkdirAll(ownerDir, 0755); err != nil {
		return models.File{}, err
	}

	storedName := fileID.String()
	storagePath := filepath.Join(ownerDir, storedName)

	dst, err := os.OpenFile(storagePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return models.File{}, err
	}
	defer dst.Close()

	sniffBuffer := make([]byte, 512)
	sniffBytes, err := io.ReadFull(src, sniffBuffer)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		_ = os.Remove(storagePath)
		return models.File{}, err
	}
	if sniffBytes == 0 {
		_ = os.Remove(storagePath)
		return models.File{}, ErrEmptyFile
	}

	contentType := http.DetectContentType(sniffBuffer[:sniffBytes])

	size, err := io.Copy(dst, io.MultiReader(bytes.NewReader(sniffBuffer[:sniffBytes]), src))
	if err != nil {
		_ = os.Remove(storagePath)
		return models.File{}, err
	}
	if size == 0 {
		_ = os.Remove(storagePath)
		return models.File{}, ErrEmptyFile
	}

	file := models.File{
		ID:           fileID,
		OwnerID:      ownerID,
		OriginalName: sanitizeOriginalName(header.Filename),
		StoragePath:  storagePath,
		ContentType:  contentType,
		SizeBytes:    size,
		Status:       StatusPending,
	}

	created, err := s.files.Create(ctx, file)
	if err != nil {
		_ = os.Remove(storagePath)
		return models.File{}, err
	}
	return created, nil
}

func (s *Service) List(ctx context.Context, ownerID uuid.UUID) ([]models.File, error) {
	return s.files.ListByOwner(ctx, ownerID)
}

func (s *Service) GetDownload(ctx context.Context, ownerID uuid.UUID, fileID uuid.UUID) (models.File, error) {
	file, err := s.files.GetByIDAndOwner(ctx, fileID, ownerID)
	if err != nil {
		return models.File{}, err
	}
	if file.Status != StatusReady {
		return models.File{}, ErrFileNotReady
	}
	return file, nil
}

func sanitizeOriginalName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "upload"
	}
	return base
}
