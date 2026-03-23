package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	db "github.com/taiwoajasa245/torah_ai_backend/db/sqlc"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInternalServer     = errors.New("internal server error")
)

type Repository interface {
	CreateUser(ctx context.Context, user User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetAllUsers(ctx context.Context) ([]User, error)
	SavePasswordReset(ctx context.Context, email, otp string, expiresAt time.Time) error
	GetPasswordReset(ctx context.Context, email string) (string, time.Time, error)
	DeletePasswordReset(ctx context.Context, email string) error
	UpdateUserPassword(ctx context.Context, email, hashed string) error
	GetUserByID(ctx context.Context, id string) (*User, error)
}

type repository struct {
	queries *db.Queries
}

func NewRepository(database *sql.DB) Repository {
	return &repository{
		queries: db.New(database),
	}
}

// db.User never leaves this file
func fromDBUser(u db.User) *User {
	return &User{
		ID:        u.ID,
		Email:     u.Email,
		UserName:  u.Username.String,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (r *repository) CreateUser(ctx context.Context, user User) (*User, error) {
	result, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:    user.Email,
		Password: user.Password,
		Username: sql.NullString{String: user.UserName, Valid: user.UserName != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("CreateUser: %w", err)
	}

	return fromDBUser(result), nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	result, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("GetUserByEmail: %w", err)
	}

	return fromDBUser(result), nil
}

func (r *repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	result, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("GetUserByID: %w", err)
	}

	return fromDBUser(result), nil
}

func (r *repository) GetAllUsers(ctx context.Context) ([]User, error) {
	rows, err := r.queries.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllUsers: %w", err)
	}

	var users []User
	for _, row := range rows {
		users = append(users, User{
			ID:    row.ID,
			Email: row.Email,
		})
	}

	return users, nil
}

func (r *repository) SavePasswordReset(ctx context.Context, email, otp string, expiresAt time.Time) error {
	err := r.queries.SavePasswordReset(ctx, db.SavePasswordResetParams{
		Email:     email,
		Otp:       otp,
		ExpiresAt: expiresAt.UTC(),
	})
	if err != nil {
		return fmt.Errorf("SavePasswordReset: %w", err)
	}
	return nil
}

func (r *repository) GetPasswordReset(ctx context.Context, email string) (string, time.Time, error) {
	result, err := r.queries.GetPasswordReset(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", time.Time{}, fmt.Errorf("no password reset record found")
		}
		return "", time.Time{}, fmt.Errorf("GetPasswordReset: %w", err)
	}

	return result.Otp, result.ExpiresAt, nil
}

func (r *repository) DeletePasswordReset(ctx context.Context, email string) error {
	err := r.queries.DeletePasswordReset(ctx, email)
	if err != nil {
		return fmt.Errorf("DeletePasswordReset: %w", err)
	}
	return nil
}

func (r *repository) UpdateUserPassword(ctx context.Context, email, hashed string) error {
	err := r.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		Password: hashed,
		Email:    email,
	})
	if err != nil {
		return fmt.Errorf("UpdateUserPassword: %w", err)
	}
	return nil
}
