package projects

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oscargsdev/undr-mono/internal/users"
)

var (
	// ErrRecordNotFound indicates that a requested record does not exist.
	ErrRecordNotFound = errors.New("record not found")
)

// Project represents a musical project.
type Project struct {
	ID      int64        `json:"id"`
	OwnerID users.UserID `json:"owner_id"`
	Name    string       `json:"name"`
}

// ProjectModel provides database operations for projects.
type ProjectModel struct {
	db *pgxpool.Pool
}

// NewProjectModel creates a ProjectModel backed by db.
func NewProjectModel(db *pgxpool.Pool) *ProjectModel {
	return &ProjectModel{db: db}
}

func (m *ProjectModel) Insert(ctx context.Context, project *Project) (*Project, error) {
	query := `INSERT INTO projects (OWNER_ID, NAME) VALUES ($1, $2) RETURNING ID, OWNER_ID, NAME`
	args := []any{&project.OwnerID, &project.Name}

	insertedProject := &Project{}

	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := m.db.QueryRow(queryContext, query, args...).Scan(
		&insertedProject.ID,
		&insertedProject.OwnerID,
		&insertedProject.Name)

	if err != nil {
		// TODO: Handle pgx error on constraints
		return nil, err
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
		SELECT id, owner_id, name
		FROM projects
		WHERE id = $1`

	var project Project

	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := m.db.QueryRow(queryCtx, query, id).Scan(
		&project.ID,
		&project.OwnerID,
		&project.Name,
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
