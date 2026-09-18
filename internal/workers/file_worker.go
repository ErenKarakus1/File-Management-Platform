package workers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ErenKarakus1/File-Management-Platform/internal/repository"
)

type FileWorker struct {
	files                 *repository.FileRepository
	processingWorkerCount int
	deleteWorkerCount     int
	pollInterval          time.Duration
}

func NewFileWorker(files *repository.FileRepository, processingWorkerCount int, deleteWorkerCount int, pollInterval time.Duration) *FileWorker {
	if processingWorkerCount < 1 {
		processingWorkerCount = 1
	}
	if deleteWorkerCount < 1 {
		deleteWorkerCount = 1
	}
	if pollInterval <= 0 {
		pollInterval = time.Second
	}

	return &FileWorker{
		files:                 files,
		processingWorkerCount: processingWorkerCount,
		deleteWorkerCount:     deleteWorkerCount,
		pollInterval:          pollInterval,
	}
}

func (w *FileWorker) Start(ctx context.Context) func() {
	var wg sync.WaitGroup
	for i := 0; i < w.processingWorkerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			w.runProcessing(ctx, workerID)
		}(i + 1)
	}
	for i := 0; i < w.deleteWorkerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			w.runDeleting(ctx, workerID)
		}(i + 1)
	}

	return wg.Wait
}

func (w *FileWorker) runProcessing(ctx context.Context, workerID int) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		if err := w.processNextPending(ctx); err != nil {
			if !errors.Is(err, repository.ErrPendingFileNotFound) && !errors.Is(err, context.Canceled) {
				log.Printf("file processing worker %d: %v", workerID, err)
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *FileWorker) runDeleting(ctx context.Context, workerID int) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		if err := w.processNextDelete(ctx); err != nil {
			if !errors.Is(err, repository.ErrFileNotFound) && !errors.Is(err, context.Canceled) {
				log.Printf("file delete worker %d: %v", workerID, err)
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *FileWorker) processNextPending(ctx context.Context) error {
	file, err := w.files.ClaimPending(ctx)
	if err != nil {
		return err
	}

	checksum, err := calculateSHA256(file.StoragePath)
	if err != nil {
		if markErr := w.files.MarkFailed(ctx, file.ID); markErr != nil {
			return markErr
		}
		return err
	}

	return w.files.MarkReady(ctx, file.ID, checksum)
}

func (w *FileWorker) processNextDelete(ctx context.Context) error {
	file, err := w.files.ClaimDeleting(ctx)
	if err != nil {
		return err
	}

	if err := os.Remove(file.StoragePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	_ = os.Remove(filepath.Dir(file.StoragePath))

	return w.files.DeleteByIDAndOwner(ctx, file.ID, file.OwnerID)
}

func calculateSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
