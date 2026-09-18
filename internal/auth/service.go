package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ErenKarakus1/File-Management-Platform/internal/models"
	"github.com/ErenKarakus1/File-Management-Platform/internal/password"
	"github.com/ErenKarakus1/File-Management-Platform/internal/repository"
	"github.com/ErenKarakus1/File-Management-Platform/internal/validation"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email is already registered")
	ErrInvalidRequest     = errors.New("invalid request")
)

type Service struct {
	users     *repository.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

type TokenPair struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	User        User   `json:"user"`
}

type Claims struct {
	UserID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

func NewService(db *pgxpool.Pool, jwtSecret string, tokenTTL time.Duration) *Service {
	return &Service{
		users:     repository.NewUserRepository(db),
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  tokenTTL,
	}
}

type User = models.User

func (s *Service) Register(ctx context.Context, req models.RegisterRequest) (TokenPair, error) {
	req.Normalize()
	if err := validation.ValidateRegisterRequest(req); err != nil {
		return TokenPair{}, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	hash, err := password.GenerateHash(req.Password)
	if err != nil {
		return TokenPair{}, err
	}

	user, err := s.users.Create(ctx, models.User{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
	})
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyRegistered) {
			return TokenPair{}, ErrEmailTaken
		}
		return TokenPair{}, err
	}

	return s.issueToken(user)
}

func (s *Service) Login(ctx context.Context, req models.LoginRequest) (TokenPair, error) {
	req.Normalize()
	if err := validation.ValidateLoginRequest(req); err != nil {
		return TokenPair{}, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, err
	}

	if err := password.CompareHashAndPassword(user.PasswordHash, req.Password); err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}

	return s.issueToken(user)
}

func (s *Service) UserByID(ctx context.Context, id uuid.UUID) (User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return s.jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

func (s *Service) issueToken(user User) (TokenPair, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.tokenTTL)
	claims := Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.tokenTTL.Seconds()),
		User:        user,
	}, nil
}
