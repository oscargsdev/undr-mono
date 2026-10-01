package genres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oscargsdev/undr-mono/internal/users"
)

var (
	// ErrRecordNotFound indicates that a requested record does not exist.
	ErrRecordNotFound = errors.New("record not found")

	// ErrDuplicateGenre indicates that a Genre with the given name already exists.
	ErrDuplicateGenre = errors.New("duplicate genre")
)

// Genre represents a musical genre.
type Genre struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	CreatedBy users.UserID `json:"created_by"`
	CreatedAt time.Time    `json:"-"`
}

// GenreModel provides database operations for genres.
type GenreModel struct {
	db *pgxpool.Pool
}

// NewGenreModel creates a GenreModel backed by db.
func NewGenreModel(db *pgxpool.Pool) *GenreModel {
	return &GenreModel{db: db}
}

// Insert adds a genre to the db.
// Returns ErrDuplicateGenre if a genre with the given name already exists.
func (m *GenreModel) Insert(ctx context.Context, genre *Genre) (*Genre, error) {
	query := `
	INSERT INTO genres(name, created_by)
	VALUES ($1, $2)
	RETURNING id, name, created_by, created_at`

	args := []any{&genre.Name, &genre.CreatedBy}

	insertedGenre := &Genre{}

	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := m.db.QueryRow(queryContext, query, args...).Scan(
		&insertedGenre.ID,
		&insertedGenre.Name,
		&insertedGenre.CreatedBy,
		&insertedGenre.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) {
			return nil, err
		}

		switch pgErr.Code {
		case "23505":
			switch pgErr.ConstraintName {
			case "genres_name_key":
				return nil, ErrDuplicateGenre
			}

		}

		return nil, err
	}

	return insertedGenre, nil
}

// Get retrieves a genre by ID.
// Returns ErrRecordNotFound when id is invalid or no genre exists.
func (m *GenreModel) Get(ctx context.Context, id int64) (*Genre, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, created_by, created_at
		FROM genres
		WHERE id = $1`

	retrievedGenre := &Genre{}

	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := m.db.QueryRow(queryContext, query, id).Scan(
		&retrievedGenre.ID,
		&retrievedGenre.Name,
		&retrievedGenre.CreatedBy,
		&retrievedGenre.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return retrievedGenre, nil
}

// Delete removes the genre with the given ID.
// Returns ErrRecordNotFound if ID is less than 1.
// Deleting a genre that does not exist succeeds.
func (m *GenreModel) Delete(ctx context.Context, id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM genres WHERE id = $1`

	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := m.db.Exec(queryCtx, query, id)
	return err
}
