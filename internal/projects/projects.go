package projects

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrRecordNotFound indicates that a requested record does not exist.
	ErrRecordNotFound = errors.New("record not found")
)

// Project represents a musical project.
type Project struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ProjectModel provides database operations for projects.
type ProjectModel struct {
	db *pgxpool.Pool
}

// NewProjectModel creates a ProjectModel backed by db.
func NewProjectModel(db *pgxpool.Pool) *ProjectModel {
	return &ProjectModel{db: db}
}

// GetProject retrieves a project by ID.
// It returns ErrRecordNotFound when id is invalid or no project exists.
func (m *ProjectModel) GetProject(ctx context.Context, id int64) (*Project, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name
		FROM projects
		WHERE id = $1`

	var project Project

	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := m.db.QueryRow(queryCtx, query, id).Scan(
		&project.ID,
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
