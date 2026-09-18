package users

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDuplicateID          = errors.New("duplicate id")
	ErrDuplicateEmail       = errors.New("duplicate email")
	ErrDuplicateDisplayName = errors.New("duplicate display name")
)

type UserID string

type User struct {
	ID          UserID    `json:"id"`
	Email       string    `json:"email,omitzero"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

type UserModel struct {
	db *pgxpool.Pool
}

func NewUserModel(db *pgxpool.Pool) *UserModel {
	return &UserModel{db: db}
}

func (m *UserModel) Insert(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users(id, email, display_name, status) 
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, display_name, status, created_at, updated_at`
	args := []any{user.ID, user.Email, user.DisplayName, user.Status}

	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := m.db.QueryRow(queryCtx, query, args...).
		Scan(&user.ID, &user.Email, &user.DisplayName, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr):
			switch pgErr.Code {
			case "23505":
				switch pgErr.ConstraintName {
				case "users_pkey":
					return ErrDuplicateID
				case "users_email_key":
					return ErrDuplicateEmail
				case "users_display_name_key":
					return ErrDuplicateDisplayName
				default:
					return err
				}

			}
		default:
			return err
		}
	}
	return err
}
