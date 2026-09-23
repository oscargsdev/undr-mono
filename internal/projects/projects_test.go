package projects

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oscargsdev/undr-mono/internal/config"
	"github.com/oscargsdev/undr-mono/internal/database"
	"github.com/oscargsdev/undr-mono/internal/users"
)

var testUsers = []users.User{
	{
		ID:          "11111111-1111-1111-1111-111111111111",
		Email:       "user1@mail.com",
		DisplayName: "Display Name 1",
		Status:      "active",
	},
	{
		ID:          "22222222-2222-2222-2222-222222222222",
		Email:       "user2@mail.com",
		DisplayName: "Display Name 2",
		Status:      "active",
	},
	{
		ID:          "33333333-3333-3333-3333-333333333333",
		Email:       "user3@mail.com",
		DisplayName: "Display Name 3",
		Status:      "active",
	},
	{
		ID:          "44444444-4444-4444-4444-444444444444",
		Email:       "user4@mail.com",
		DisplayName: "Display Name 4",
		Status:      "active",
	},
}

func TestGetProject(t *testing.T) {
	projectModel, userModel := getModels(t)
	insertTestUsers(t, userModel, testUsers)

	testsProjects := []struct {
		name    string
		project Project
	}{
		{"get test_project_1", Project{OwnerID: testUsers[0].ID, Name: "test_project_1"}},
		{"get test_project_2", Project{OwnerID: testUsers[1].ID, Name: "test_project_2"}},
		{"get test_project_3", Project{OwnerID: testUsers[2].ID, Name: "test_project_3"}},
	}

	for _, tc := range testsProjects {
		t.Run(tc.name, func(t *testing.T) {
			insertedProject, err := projectModel.Insert(t.Context(), &tc.project)
			if err != nil {
				t.Fatalf("error while inserting project: %v", err)
			}

			retrievedProject, err := projectModel.Get(t.Context(), insertedProject.ID)
			if err != nil {
				t.Fatalf("failed to get project %v: %v", insertedProject.ID, err)
			}

			if retrievedProject.ID != insertedProject.ID {
				t.Fatalf("expected project id %d, got %d", insertedProject.ID, retrievedProject.ID)
			}
			if retrievedProject.Name != insertedProject.Name {
				t.Fatalf("expected project name: %q, got: %q", insertedProject.Name, retrievedProject.Name)
			}
		})
	}
}

func TestGetProjectIDBelowOne(t *testing.T) {
	projectModel := NewProjectModel(nil)

	t.Run("ID below 1", func(t *testing.T) {
		_, err := projectModel.Get(t.Context(), 0)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrRecordNotFound) {
			t.Fatalf("expected %v, got %v", ErrRecordNotFound, err)
		}
	})
}

func TestGetProjectNonExistent(t *testing.T) {
	projectModel, userModel := getModels(t)
	insertTestUsers(t, userModel, testUsers)

	insertedProject, err := projectModel.Insert(t.Context(), &Project{OwnerID: testUsers[0].ID, Name: "test_project_1"})
	if err != nil {
		t.Fatalf("error while inserting project: %v", err)
	}
	deleteProject(t, projectModel.db, insertedProject.ID)

	t.Run("ID not present", func(t *testing.T) {
		_, err := projectModel.Get(t.Context(), insertedProject.ID)
		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrRecordNotFound) {
			t.Fatalf("expected %v, got %v", ErrRecordNotFound, err)
		}
	})

}

func getModels(t testing.TB) (*ProjectModel, *users.UserModel) {
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

	return NewProjectModel(db), users.NewUserModel(db)
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

func insertTestUsers(t testing.TB, m *users.UserModel, testUsers []users.User) {
	t.Helper()

	for i, user := range testUsers {
		user.ID = users.UserID(uuid.New().String())
		user.Email = "user" + string(user.ID) + "@mail.com"
		user.DisplayName = "User " + string(user.ID)
		testUsers[i] = user
		m.Insert(t.Context(), &testUsers[i])
	}

	t.Cleanup(func() {
		for _, user := range testUsers {
			m.Delete(context.Background(), user.ID)
		}
	})
}
