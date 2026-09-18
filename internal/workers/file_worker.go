package workers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ErenKarakus1/File-Management-Platform/internal/repository"
)

type FileWorker struct {
	files        *repository.FileRepository
	workerCount  int
	pollInterval time.Duration
}

func NewFileWorker(files *repository.FileRepository, workerCount int, pollInterval time.Duration) *FileWorker {
	if workerCount < 1 {
		workerCount = 1
	}
	if pollInterval <= 0 {
		pollInterval = time.Second
	}

	return &FileWorker{
		files:        files,
		workerCount:  workerCount,
		pollInterval: pollInterval,
	}
}

func (w *FileWorker) Start(ctx context.Context) func() {
	var wg sync.WaitGroup
	for i := 0; i < w.workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			w.run(ctx, workerID)
		}(i + 1)
	}

	return wg.Wait
}

func (w *FileWorker) run(ctx context.Context, workerID int) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		if err := w.processNext(ctx); err != nil {
			if !errors.Is(err, repository.ErrPendingFileNotFound) && !errors.Is(err, context.Canceled) {
				log.Printf("file worker %d: %v", workerID, err)
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *FileWorker) processNext(ctx context.Context) error {
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
