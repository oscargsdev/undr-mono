package projects

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Project struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProjectModel struct {
	db *pgxpool.Pool
}

func (m ProjectModel) GetProject(id int64) (*Project, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name
		FROM projects
		WHERE id = $1`

	var project Project
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.db.QueryRow(ctx, query, id).Scan(
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
