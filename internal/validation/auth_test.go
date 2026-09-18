package validation

import (
	"strings"
	"testing"

	"github.com/ErenKarakus1/File-Management-Platform/internal/models"
)

func TestValidateRegisterRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     models.RegisterRequest
		wantErr string
	}{
		{
			name: "valid",
			req: models.RegisterRequest{
				Name:     "Eren",
				Email:    "eren@example.com",
				Password: "password123",
			},
		},
		{
			name: "missing name",
			req: models.RegisterRequest{
				Email:    "eren@example.com",
				Password: "password123",
			},
			wantErr: "name is required",
		},
		{
			name: "invalid email",
			req: models.RegisterRequest{
				Name:     "Eren",
				Email:    "not-email",
				Password: "password123",
			},
			wantErr: "invalid email",
		},
		{
			name: "short password",
			req: models.RegisterRequest{
				Name:     "Eren",
				Email:    "eren@example.com",
				Password: "short",
			},
			wantErr: "password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegisterRequest(tt.req)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}

func TestValidateLoginRequest(t *testing.T) {
	if err := ValidateLoginRequest(models.LoginRequest{
		Email:    "eren@example.com",
		Password: "password123",
	}); err != nil {
		t.Fatalf("expected valid login request, got %v", err)
	}

	if err := ValidateLoginRequest(models.LoginRequest{
		Email:    "bad",
		Password: "password123",
	}); err == nil {
		t.Fatal("expected invalid email error")
	}
}
