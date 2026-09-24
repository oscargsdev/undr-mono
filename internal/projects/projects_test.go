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

func TestInsert(t *testing.T) {
	projectModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 3)

	testsProjects := []struct {
		name    string
		project Project
	}{
		{"insert test_project_1",
			Project{
				OwnerID: testUsers[0].ID,
				Handle:  "project1",
				Name:    "test_project_1",
			}},
		{"insert test_project_2",
			Project{
				OwnerID: testUsers[1].ID,
				Handle:  "project2",
				Name:    "test_project_2",
			}},
		{"insert test_project_3",
			Project{
				OwnerID: testUsers[2].ID,
				Handle:  "project3",
				Name:    "test_project_3",
			}},
	}

	for _, tc := range testsProjects {
		t.Run(tc.name, func(t *testing.T) {
			newProject := tc.project
			insertedProject, err := projectModel.Insert(t.Context(), &newProject)
			if err != nil {
				t.Fatalf("Insert() error = %v; want nil", err)
			}

			if insertedProject.ID == 0 {
				t.Errorf("Insert() ID is zero; want non-zero value")
			}

			if insertedProject.OwnerID != newProject.OwnerID {
				t.Errorf("Insert() OwnerID = %q; want %q", insertedProject.OwnerID, newProject.OwnerID)
			}

			if insertedProject.Handle != newProject.Handle {
				t.Errorf("Insert() Handle = %q; want %q", insertedProject.Handle, newProject.Handle)
			}

			if insertedProject.Name != newProject.Name {
				t.Errorf("Insert() Name = %q; want %q", insertedProject.Name, newProject.Name)
			}

			if insertedProject.CreatedAt.IsZero() {
				t.Errorf("Insert() CreatedAt is zero; want non-zero")
			}

			if insertedProject.UpdatedAt.IsZero() {
				t.Errorf("Insert() UpdatedAt is zero; want non-zero")
			}

			if !insertedProject.CreatedAt.Equal(insertedProject.UpdatedAt) {
				t.Errorf("Insert() CreatedAt = %v, UpdatedAt = %v; want equal timestamps", insertedProject.CreatedAt, insertedProject.UpdatedAt)
			}
		})
	}
}

func TestInsertDuplicates(t *testing.T) {
	projectModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 2)

	originalHandle := "project1"
	originalName := "Project 1"

	originalProject := &Project{
		OwnerID: testUsers[0].ID,
		Handle:  originalHandle,
		Name:    originalName,
	}

	_, err := projectModel.Insert(t.Context(), originalProject)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}

	duplicateHandleProject := &Project{
		OwnerID: testUsers[1].ID,
		Handle:  originalHandle,
		Name:    originalName, // Projects can have the same name, Handle and ID is what differentiates them
	}

	_, err = projectModel.Insert(t.Context(), duplicateHandleProject)
	if err == nil {
		t.Fatalf("Insert() error = nil; want %v", ErrDuplicateHandle)
	}

	if !errors.Is(err, ErrDuplicateHandle) {
		t.Errorf("Insert() error = %v; want %v", err, ErrDuplicateHandle)
	}

}

func TestGetProject(t *testing.T) {
	projectModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 3)

	testsProjects := []struct {
		name    string
		project Project
	}{
		{"get test_project_1",
			Project{
				OwnerID: testUsers[0].ID,
				Handle:  "project1",
				Name:    "test_project_1",
			}},
		{"get test_project_2",
			Project{
				OwnerID: testUsers[1].ID,
				Handle:  "project2",
				Name:    "test_project_2",
			}},
		{"get test_project_3",
			Project{
				OwnerID: testUsers[2].ID,
				Handle:  "project3",
				Name:    "test_project_3",
			}},
	}

	for _, tc := range testsProjects {
		t.Run(tc.name, func(t *testing.T) {
			insertedProject, err := projectModel.Insert(t.Context(), &tc.project)
			if err != nil {
				t.Fatalf("Insert() error = %v; want nil", err)
			}

			retrievedProject, err := projectModel.Get(t.Context(), insertedProject.ID)
			if err != nil {
				t.Fatalf("Get() error = %v; want nil", err)
			}

			if retrievedProject.ID != insertedProject.ID {
				t.Errorf("Get() ID = %d; want %d", retrievedProject.ID, insertedProject.ID)
			}

			if retrievedProject.OwnerID != insertedProject.OwnerID {
				t.Errorf("Get() OwnerID = %q; want %q", retrievedProject.OwnerID, insertedProject.OwnerID)
			}

			if retrievedProject.Handle != insertedProject.Handle {
				t.Errorf("Get() Handle = %q; want %q", retrievedProject.Handle, insertedProject.Handle)
			}

			if retrievedProject.Name != insertedProject.Name {
				t.Errorf("Get() Name = %q; want %q", retrievedProject.Name, insertedProject.Name)
			}

			if !retrievedProject.CreatedAt.Equal(insertedProject.CreatedAt) {
				t.Errorf("Get() CreatedAt = %v; want %v", retrievedProject.CreatedAt, insertedProject.CreatedAt)
			}

			if !retrievedProject.UpdatedAt.Equal(insertedProject.UpdatedAt) {
				t.Errorf("Get() UpdatedAt = %v; want %v", retrievedProject.UpdatedAt, insertedProject.UpdatedAt)
			}
		})
	}
}

func TestGetProjectIDBelowOne(t *testing.T) {
	projectModel := NewProjectModel(nil)

	t.Run("ID below 1", func(t *testing.T) {
		_, err := projectModel.Get(t.Context(), 0)
		if err == nil {
			t.Fatalf("Get() error = nil; want %v", ErrRecordNotFound)
		}

		if !errors.Is(err, ErrRecordNotFound) {
			t.Errorf("Get() error = %v; want %v", err, ErrRecordNotFound)
		}
	})
}

func TestGetProjectNonExistent(t *testing.T) {
	projectModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 1)

	insertedProject, err := projectModel.Insert(t.Context(), &Project{OwnerID: testUsers[0].ID, Name: "test_project_1"})
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	deleteProject(t, projectModel.db, insertedProject.ID)

	t.Run("ID not present", func(t *testing.T) {
		_, err := projectModel.Get(t.Context(), insertedProject.ID)
		if err == nil {
			t.Fatalf("Get() error = nil; want %v", ErrRecordNotFound)
		}

		if !errors.Is(err, ErrRecordNotFound) {
			t.Errorf("Get() error = %v; want %v", err, ErrRecordNotFound)
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

func insertTestUsers(t testing.TB, m *users.UserModel, n int) []users.User {
	t.Helper()

	insertedUsers := make([]users.User, 0, n)

	for range n {
		id := users.UserID(uuid.NewString())
		email := "user" + string(id) + "@mail.com"
		displayName := "User " + string(id)

		user := users.User{
			ID:          id,
			Email:       email,
			DisplayName: displayName,
			Status:      "active",
		}

		insertedUser, err := m.Insert(t.Context(), &user)
		if err != nil {
			t.Fatalf("error while inserting test users: %v", err)
		}

		t.Cleanup(func() { deleteTestUser(t, m, insertedUser) })
		insertedUsers = append(insertedUsers, *insertedUser)
	}

	return insertedUsers
}

func deleteTestUser(t testing.TB, m *users.UserModel, user *users.User) {
	t.Helper()

	err := m.Delete(context.Background(), user.ID)
	if err != nil {
		t.Errorf("error while deleting test users: %v", err)
	}
}
