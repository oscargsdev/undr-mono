// Package users provides user persistence operations.
package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrRecordNotFound indicates that a requested record does not exist.
	ErrRecordNotFound = errors.New("record not found")

	// ErrInvalidID indicates that the passed ID is invalid or malformed.
	ErrInvalidID = errors.New("invalid id")

	// ErrDuplicateID indicates that a user with the given ID already exists.
	ErrDuplicateID = errors.New("duplicate id")

	// ErrDuplicateEmail indicates that a user with the given email address already exists.
	ErrDuplicateEmail = errors.New("duplicate email")

	// ErrDuplicateDisplayName indicates that a user with the given display name already exists.
	ErrDuplicateDisplayName = errors.New("duplicate display name")
)

// UserID identifies a user.
type UserID string

// User represents an undr user.
type User struct {
	ID          UserID    `json:"id"`
	Email       string    `json:"email,omitzero"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

// UserModel provides database operations for users.
type UserModel struct {
	db *pgxpool.Pool
}

// NewUserModel creates a UserModel backed by db.
func NewUserModel(db *pgxpool.Pool) *UserModel {
	return &UserModel{db: db}
}

// Insert adds a user to the db.
// Returns ErrDuplicateID if a user with the given ID already exists.
// Returns ErrDuplicateEmail if a user with the given email address already exists.
// Returns ErrDuplicateDisplayName if a user with the given display name already exists.
func (m *UserModel) Insert(ctx context.Context, user *User) (*User, error) {
	query := `
		INSERT INTO users(id, email, display_name, status) 
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, display_name, status, created_at, updated_at`
	args := []any{user.ID, user.Email, user.DisplayName, user.Status}

	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	insertedUser := User{}

	err := m.db.QueryRow(queryCtx, query, args...).Scan(
		&insertedUser.ID,
		&insertedUser.Email,
		&insertedUser.DisplayName,
		&insertedUser.Status,
		&insertedUser.CreatedAt,
		&insertedUser.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr):
			switch pgErr.Code {
			case "23505":
				switch pgErr.ConstraintName {
				case "users_pkey":
					return nil, ErrDuplicateID
				case "users_email_key":
					return nil, ErrDuplicateEmail
				case "users_display_name_key":
					return nil, ErrDuplicateDisplayName
				default:
					return nil, err
				}
			default:
				return nil, err
			}
		default:
			return nil, err
		}
	}

	return &insertedUser, nil
}

// Get retrieves a user by ID.
// Returns ErrInvalidID if the given ID is invalid or malformed.
// Returns ErrRecordNotFound if no user with the given ID exists.
func (m *UserModel) Get(ctx context.Context, id UserID) (*User, error) {
	parsedID, err := uuid.Parse(string(id))
	if err != nil {
		return nil, ErrInvalidID
	}

	query := `
		SELECT id, email, display_name, status, created_at, updated_at
		FROM users WHERE id = $1`

	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var user User

	err = m.db.QueryRow(queryCtx, query, parsedID).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
