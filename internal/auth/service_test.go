package auth

import (
	"testing"
	"time"

	"github.com/ErenKarakus1/File-Management-Platform/internal/models"
	"github.com/google/uuid"
)

func TestIssueAndParseToken(t *testing.T) {
	service := &Service{
		jwtSecret: []byte("test-secret"),
		tokenTTL:  time.Hour,
	}
	userID := uuid.New()

	token, err := service.issueToken(models.User{
		ID:    userID,
		Name:  "Eren",
		Email: "eren@example.com",
	})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	claims, err := service.ParseToken(token.AccessToken)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("expected user id %s, got %s", userID, claims.UserID)
	}
	if claims.Subject != userID.String() {
		t.Fatalf("expected subject %s, got %s", userID, claims.Subject)
	}
}

func TestParseTokenRejectsWrongSecret(t *testing.T) {
	service := &Service{
		jwtSecret: []byte("test-secret"),
		tokenTTL:  time.Hour,
	}
	token, err := service.issueToken(models.User{ID: uuid.New()})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	otherService := &Service{jwtSecret: []byte("other-secret")}
	if _, err := otherService.ParseToken(token.AccessToken); err == nil {
		t.Fatal("expected token signed with different secret to fail")
	}
}
