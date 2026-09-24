package projects

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

	// ErrDuplicateHandle indicates that a Project with the given handle already exists.
	ErrDuplicateHandle = errors.New("duplicate handle")
)

// Project represents a musical project.
type Project struct {
	ID        int64        `json:"id"`
	OwnerID   users.UserID `json:"owner_id"`
	Handle    string       `json:"handle"`
	Name      string       `json:"name"`
	CreatedAt time.Time    `json:"-"`
	UpdatedAt time.Time    `json:"-"`
}

// ProjectModel provides database operations for projects.
type ProjectModel struct {
	db *pgxpool.Pool
}

// NewProjectModel creates a ProjectModel backed by db.
func NewProjectModel(db *pgxpool.Pool) *ProjectModel {
	return &ProjectModel{db: db}
}

// Insert adds a project to the db.
// Returns ErrDuplicateHandle if a project with the given handle already exists.
func (m *ProjectModel) Insert(ctx context.Context, project *Project) (*Project, error) {
	query := `
		INSERT INTO projects (owner_id, handle, name)
		VALUES ($1, $2, $3)
		RETURNING id, owner_id, handle, name, created_at, updated_at`

	args := []any{project.OwnerID, project.Handle, project.Name}

	insertedProject := &Project{}

	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := m.db.QueryRow(queryContext, query, args...).Scan(
		&insertedProject.ID,
		&insertedProject.OwnerID,
		&insertedProject.Handle,
		&insertedProject.Name,
		&insertedProject.CreatedAt,
		&insertedProject.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr):
			switch pgErr.Code {
			case "23505":
				switch pgErr.ConstraintName {
				case "projects_handle_key":
					return nil, ErrDuplicateHandle
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

	return insertedProject, nil
}

// Get retrieves a project by ID.
// It returns ErrRecordNotFound when id is invalid or no project exists.
func (m *ProjectModel) Get(ctx context.Context, id int64) (*Project, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, owner_id, handle, name, created_at, updated_at
		FROM projects
		WHERE id = $1`

	var project Project

	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := m.db.QueryRow(queryCtx, query, id).Scan(
		&project.ID,
		&project.OwnerID,
		&project.Handle,
		&project.Name,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &project, nil
}
