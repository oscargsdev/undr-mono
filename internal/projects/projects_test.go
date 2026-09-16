package projects

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oscargsdev/undr-mono/internal/config"
	"github.com/oscargsdev/undr-mono/internal/database"
)

func TestGetProject(t *testing.T) {
	model := getModel(t)
	defer model.db.Close()

	testsExistingProjects := []struct {
		name                string
		id                  int64
		expectedProjectName string
	}{
		{"gets project 1 by ID", 1, "undersleep"},
		{"gets project 2 by ID", 2, "epifani"},
		{"gets project 3 by ID", 3, "bardo"},
	}

	for _, tc := range testsExistingProjects {
		t.Run(tc.name, func(t *testing.T) {
			project, err := model.GetProject(t.Context(), tc.id)
			if err != nil {
				t.Fatalf("failed to get project %v: %v", tc.id, err)
			}

			if project.Name != tc.expectedProjectName {
				t.Fatalf("expected project name: %v, got: %v", tc.expectedProjectName, project.Name)
			}
		})
	}

	testsNonExistingProjects := []struct {
		name  string
		id    int64
		error error
	}{
		{"ID lesser than 1", 0, ErrRecordNotFound},
		{"ID not present", 999, ErrRecordNotFound},
	}

	for _, tc := range testsNonExistingProjects {
		t.Run(tc.name, func(t *testing.T) {
			_, err := model.GetProject(t.Context(), tc.id)
			if err == nil {
				t.Fatal("expected error")
			}

			if !errors.Is(err, ErrRecordNotFound) {
				t.Fatalf("expected %v, got %v", tc.error, err)
			}
		})
	}
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

	model := ProjectModel{
		db: db,
	}

	return &model
}
