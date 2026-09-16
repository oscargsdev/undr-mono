package projects

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oscargsdev/undr-mono/internal/config"
	"github.com/oscargsdev/undr-mono/internal/database"
)

func TestGetProject(t *testing.T) {
	model := getModel(t)

	testsProjects := []struct {
		testName    string
		projectName string
	}{
		{"get test_project_1", "test_project_1"},
		{"get test_project_2", "test_project_2"},
		{"get test_project_3", "test_project_3"},
	}

	for _, tc := range testsProjects {
		t.Run(tc.testName, func(t *testing.T) {
			id := insertProject(t, model.db, tc.projectName)

			project, err := model.GetProject(t.Context(), id)
			if err != nil {
				t.Fatalf("failed to get project %v: %v", id, err)
			}

			if project.ID != id {
				t.Fatalf("expected project id %d, got %d", id, project.ID)
			}
			if project.Name != tc.projectName {
				t.Fatalf("expected project name: %q, got: %q", tc.projectName, project.Name)
			}
		})
	}
}

func TestGetProjectIDBelowOne(t *testing.T) {
	model := NewProjectModel(nil)

	t.Run("ID below 1", func(t *testing.T) {
		_, err := model.GetProject(t.Context(), 0)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrRecordNotFound) {
			t.Fatalf("expected %v, got %v", ErrRecordNotFound, err)
		}
	})
}

func TestGetProjectNonExistent(t *testing.T) {
	model := getModel(t)

	id := insertProject(t, model.db, "test_project")
	deleteProject(t, model.db, id)

	t.Run("ID not present", func(t *testing.T) {
		_, err := model.GetProject(t.Context(), id)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrRecordNotFound) {
			t.Fatalf("expected %v, got %v", ErrRecordNotFound, err)
		}
	})

}

func getModel(t testing.TB) *ProjectModel {
	t.Helper()

	cfg, err := config.LoadFromEnv("../../.env")
	if err != nil {
		t.Fatalf("error while loading config from env: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	db, err := database.OpenDB(ctx, cfg)
	if err != nil {
		t.Fatalf("error opening db connection: %v", err)
	}

	t.Cleanup(db.Close)

	return NewProjectModel(db)
}

func insertProject(t testing.TB, db *pgxpool.Pool, projectName string) int64 {
	t.Helper()

	var id int64
	query := `INSERT INTO projects (name) VALUES ($1) RETURNING ID`

	err := db.QueryRow(t.Context(), query, projectName).Scan(&id)
	if err != nil {
		t.Fatalf("error while inserting project: %v", err)
	}

	t.Cleanup(func() { deleteProject(t, db, id) })

	return id
}

func deleteProject(t testing.TB, db *pgxpool.Pool, projectID int64) {
	t.Helper()

	query := `DELETE FROM projects WHERE id = $1`

	queryCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.Exec(queryCtx, query, projectID)
	if err != nil {
		t.Fatalf("error while deleting from projects: %v", err)
	}
}
