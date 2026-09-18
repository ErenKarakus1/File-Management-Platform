package workers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateSHA256(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.txt")
	if err := os.WriteFile(path, []byte("hello"), 0600); err != nil {
		t.Fatalf("write sample file: %v", err)
	}

	checksum, err := calculateSHA256(path)
	if err != nil {
		t.Fatalf("calculate checksum: %v", err)
	}

	const want = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if checksum != want {
		t.Fatalf("expected %s, got %s", want, checksum)
	}
}
