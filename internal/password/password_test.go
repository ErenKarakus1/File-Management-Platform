package password

import "testing"

func TestGenerateHashAndCompare(t *testing.T) {
	hash, err := GenerateHash("password123")
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}
	if hash == "password123" {
		t.Fatal("hash should not equal raw password")
	}
	if err := CompareHashAndPassword(hash, "password123"); err != nil {
		t.Fatalf("expected password to match hash: %v", err)
	}
	if err := CompareHashAndPassword(hash, "wrong-password"); err == nil {
		t.Fatal("expected wrong password to fail")
	}
}
