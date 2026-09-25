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

	// ErrInvalidStatus indicates that a Project has an invalid status.
	ErrInvalidStatus = errors.New("invalid status")
)

const (
	// StatusActive indicates an active project.
	StatusActive = "active"
	// StatusLimbo indicates a project in limbo (we don't know if it's active or inactive; it may be in pause, indefinite hiatus, etc).
	StatusLimbo = "limbo"
	// StatusInactive indicates an inactive project.
	StatusInactive = "inactive"
)

// Project represents a musical project.
type Project struct {
	ID        int64        `json:"id"`
	OwnerID   users.UserID `json:"owner_id"`
	Handle    string       `json:"handle"`
	Name      string       `json:"name"`
	Status    string       `json:"status"`
	Bio       string       `json:"bio"`
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
// Returns ErrInvalidStatus if the project has an invalid status.
func (m *ProjectModel) Insert(ctx context.Context, project *Project) (*Project, error) {
	query := `
		INSERT INTO projects (owner_id, handle, name, status, bio)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, owner_id, handle, name, status, bio, created_at, updated_at`

	args := []any{project.OwnerID, project.Handle, project.Name, project.Status, project.Bio}

	insertedProject := &Project{}

	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := m.db.QueryRow(queryContext, query, args...).Scan(
		&insertedProject.ID,
		&insertedProject.OwnerID,
		&insertedProject.Handle,
		&insertedProject.Name,
		&insertedProject.Status,
		&insertedProject.Bio,
		&insertedProject.CreatedAt,
		&insertedProject.UpdatedAt,
	)

	if err != nil {
		return nil, mapProjectError(err)
	}

	return insertedProject, nil
}

// Get retrieves a project by ID.
// Returns ErrRecordNotFound when id is invalid or no project exists.
func (m *ProjectModel) Get(ctx context.Context, id int64) (*Project, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, owner_id, handle, name, status, bio, created_at, updated_at
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
		&project.Status,
		&project.Bio,
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

// Delete removes the project with the given ID.
// Returns ErrRecordNotFound if ID is less than 1.
// Deleting a project that does not exist succeeds.
func (m *ProjectModel) Delete(ctx context.Context, id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM projects WHERE id = $1`

	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := m.db.Exec(queryCtx, query, id)
	return err
}

// Update updates mutable fields of a project.
// Returns ErrRecordNotFound when id is invalid or no project exists.
// Returns ErrDuplicateHandle if a project with the given handle already exists.
// Returns ErrInvalidStatus if the project has an invalid status.
func (m *ProjectModel) Update(ctx context.Context, project *Project) (*Project, error) {
	if project.ID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		UPDATE projects
		SET handle = $1, name = $2, status = $3, bio = $4, updated_at = now()
		WHERE id = $5
		RETURNING id, owner_id, handle, name, status, bio, created_at, updated_at`

	args := []any{
		project.Handle,
		project.Name,
		project.Status,
		project.Bio,
		project.ID,
	}

	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var updatedProject Project

	err := m.db.QueryRow(queryCtx, query, args...).Scan(
		&updatedProject.ID,
		&updatedProject.OwnerID,
		&updatedProject.Handle,
		&updatedProject.Name,
		&updatedProject.Status,
		&updatedProject.Bio,
		&updatedProject.CreatedAt,
		&updatedProject.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, mapProjectError(err)
	}

	return &updatedProject, nil
}

func mapProjectError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case "23505":
		switch pgErr.ConstraintName {
		case "projects_handle_key":
			return ErrDuplicateHandle
		}
	case "23514":
		switch pgErr.ConstraintName {
		case "project_status_check":
			return ErrInvalidStatus
		}
	}

	return err
}
