package validation

import (
	"errors"
	"net/mail"

	"github.com/ErenKarakus1/File-Management-Platform/internal/models"
)

func ValidateRegisterRequest(req models.RegisterRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if len(req.Name) < 3 {
		return errors.New("name must be at least 3 characters")
	}
	if len(req.Name) > 50 {
		return errors.New("name must be at most 50 characters")
	}
	if err := validateEmail(req.Email); err != nil {
		return err
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(req.Password) > 128 {
		return errors.New("password must be at most 128 characters")
	}
	return nil
}

func ValidateLoginRequest(req models.LoginRequest) error {
	if err := validateEmail(req.Email); err != nil {
		return err
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("invalid email")
	}
	return nil
}
