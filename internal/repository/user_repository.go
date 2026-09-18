package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/File-Management-Platform/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrEmailAlreadyRegistered = errors.New("email is already registered")
	ErrUserNotFound           = errors.New("user not found")
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user models.User) (models.User, error) {
	err := r.db.QueryRow(ctx, `
		insert into users (id, name, email, password_hash)
		values ($1, $2, $3, $4)
		returning id, name, email, password_hash, created_at, updated_at
	`, user.ID, user.Name, user.Email, user.PasswordHash).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, ErrEmailAlreadyRegistered
		}
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx, `
		select id, name, email, password_hash, created_at, updated_at
		from users
		where email = $1
	`, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx, `
		select id, name, email, password_hash, created_at, updated_at
		from users
		where id = $1
	`, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}
