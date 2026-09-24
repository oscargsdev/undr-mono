package projects

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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
				Status:  StatusActive,
			}},
		{"insert test_project_2",
			Project{
				OwnerID: testUsers[1].ID,
				Handle:  "project2",
				Name:    "test_project_2",
				Status:  StatusLimbo,
			}},
		{"insert test_project_3",
			Project{
				OwnerID: testUsers[2].ID,
				Handle:  "project3",
				Name:    "test_project_3",
				Status:  StatusActive,
			}},
	}

	for _, tc := range testsProjects {
		t.Run(tc.name, func(t *testing.T) {
			newProject := tc.project
			insertedProject, err := projectModel.Insert(t.Context(), &newProject)
			if err != nil {
				t.Fatalf("Insert() error = %v; want nil", err)
			}
			t.Cleanup(func() {
				err := projectModel.Delete(context.Background(), insertedProject.ID)
				if err != nil {
					t.Errorf("Delete() error = %v; want nil", err)
				}
			})

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

			if insertedProject.Status != newProject.Status {
				t.Errorf("Insert() Status = %q; want %q", insertedProject.Status, newProject.Status)
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
		Status:  StatusActive,
	}

	insertedProject, err := projectModel.Insert(t.Context(), originalProject)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := projectModel.Delete(context.Background(), insertedProject.ID)
		if err != nil {
			t.Errorf("Delete() error = %v; want nil", err)
		}
	})

	duplicateHandleProject := &Project{
		OwnerID: testUsers[1].ID,
		Handle:  originalHandle,
		Name:    originalName, // Projects can have the same name, Handle and ID is what differentiates them
		Status:  StatusActive,
	}

	_, err = projectModel.Insert(t.Context(), duplicateHandleProject)
	if err == nil {
		t.Fatalf("Insert() error = nil; want %v", ErrDuplicateHandle)
	}

	if !errors.Is(err, ErrDuplicateHandle) {
		t.Errorf("Insert() error = %v; want %v", err, ErrDuplicateHandle)
	}

}

func TestInsertInvalidStatus(t *testing.T) {
	projectModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 1)

	invalidStatusProject := &Project{
		OwnerID: testUsers[0].ID,
		Handle:  "project1",
		Name:    "Project 1",
		Status:  "invalid",
	}

	_, err := projectModel.Insert(t.Context(), invalidStatusProject)
	if err == nil {
		t.Fatalf("Insert() error = nil; want %v", ErrInvalidStatus)
	}

	if !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("Insert() error = %v; want %v", err, ErrInvalidStatus)
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
				Status:  StatusActive,
			}},
		{"get test_project_2",
			Project{
				OwnerID: testUsers[1].ID,
				Handle:  "project2",
				Name:    "test_project_2",
				Status:  StatusInactive,
			}},
		{"get test_project_3",
			Project{
				OwnerID: testUsers[2].ID,
				Handle:  "project3",
				Name:    "test_project_3",
				Status:  StatusInactive,
			}},
	}

	for _, tc := range testsProjects {
		t.Run(tc.name, func(t *testing.T) {
			insertedProject, err := projectModel.Insert(t.Context(), &tc.project)
			if err != nil {
				t.Fatalf("Insert() error = %v; want nil", err)
			}
			t.Cleanup(func() {
				err := projectModel.Delete(context.Background(), insertedProject.ID)
				if err != nil {
					t.Errorf("Delete() error = %v; want nil", err)
				}
			})

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

			if retrievedProject.Status != insertedProject.Status {
				t.Errorf("Get() Status = %q; want %q", retrievedProject.Status, insertedProject.Status)
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

	project := &Project{
		OwnerID: testUsers[0].ID,
		Name:    "test_project_1",
		Status:  StatusActive,
	}

	insertedProject, err := projectModel.Insert(t.Context(), project)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}

	err = projectModel.Delete(t.Context(), insertedProject.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v; want nil", err)
	}

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

func TestDelete(t *testing.T) {
	projectModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 1)

	project := &Project{
		OwnerID: testUsers[0].ID,
		Name:    "test_project_1",
		Status:  StatusActive,
	}

	insertedProject, err := projectModel.Insert(t.Context(), project)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := projectModel.Delete(context.Background(), insertedProject.ID)
		if err != nil {
			t.Errorf("Delete() error = %v; want nil", err)
		}
	})

	err = projectModel.Delete(t.Context(), insertedProject.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v; want nil", err)
	}

	_, err = projectModel.Get(t.Context(), insertedProject.ID)
	if err == nil {
		t.Fatalf("Get() error = nil; want %v", ErrRecordNotFound)
	}

	if !errors.Is(err, ErrRecordNotFound) {
		t.Errorf("Get() error = %v; want %v", err, ErrRecordNotFound)
	}
}

func TestDeleteIDBelowOne(t *testing.T) {
	projectModel := NewProjectModel(nil)

	err := projectModel.Delete(t.Context(), 0)
	if err == nil {
		t.Fatalf("Delete() error = nil; want %v", ErrRecordNotFound)
	}

	if !errors.Is(err, ErrRecordNotFound) {
		t.Errorf("Delete() error = %v; want %v", err, ErrRecordNotFound)
	}
}

func TestUpdate(t *testing.T) {
	projectModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 1)

	project := &Project{
		OwnerID: testUsers[0].ID,
		Name:    "test_project_1",
		Status:  StatusActive,
	}

	insertedProject, err := projectModel.Insert(t.Context(), project)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := projectModel.Delete(context.Background(), insertedProject.ID)
		if err != nil {
			t.Errorf("Delete() error = %v; want nil", err)
		}
	})

	updatedHandle := "updatedHandle"
	updatedName := "Updated Name"
	updatedStatus := StatusInactive

	project.ID = insertedProject.ID
	project.Handle = updatedHandle
	project.Name = updatedName
	project.Status = updatedStatus

	updatedProject, err := projectModel.Update(t.Context(), project)
	if err != nil {
		t.Fatalf("Update() error = %v; want nil", err)
	}

	if updatedProject.ID != insertedProject.ID {
		t.Errorf("Update() ID = %d; want %d", updatedProject.ID, insertedProject.ID)
	}

	if updatedProject.OwnerID != insertedProject.OwnerID {
		t.Errorf("Update() OwnerID = %q; want %q", updatedProject.OwnerID, insertedProject.OwnerID)
	}

	if updatedProject.Handle != updatedHandle {
		t.Errorf("Update() Handle = %q; want %q", updatedProject.Handle, updatedHandle)
	}

	if updatedProject.Name != updatedName {
		t.Errorf("Update() Name = %q; want %q", updatedProject.Name, updatedName)
	}

	if updatedProject.Status != updatedStatus {
		t.Errorf("Update() Status = %q; want %q", updatedProject.Status, updatedStatus)
	}

	if !updatedProject.CreatedAt.Equal(insertedProject.CreatedAt) {
		t.Errorf("Update() CreatedAt = %v; want %v", updatedProject.CreatedAt, insertedProject.CreatedAt)
	}

	if !updatedProject.UpdatedAt.After(insertedProject.UpdatedAt) {
		t.Errorf("Update() UpdatedAt = %v; want > %v", updatedProject.UpdatedAt, insertedProject.UpdatedAt)
	}
}

func TestUpdateProjectNotFound(t *testing.T) {
	projectModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 1)

	project := &Project{
		OwnerID: testUsers[0].ID,
		Handle:  "project1",
		Name:    "test_project_1",
		Status:  StatusActive,
	}

	insertedProject, err := projectModel.Insert(t.Context(), project)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}

	err = projectModel.Delete(t.Context(), insertedProject.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v; want nil", err)
	}

	updatedHandle := "updatedHandle"
	updatedName := "Updated Name"
	updatedStatus := StatusInactive

	project.ID = insertedProject.ID
	project.Handle = updatedHandle
	project.Name = updatedName
	project.Status = updatedStatus

	_, err = projectModel.Update(t.Context(), project)
	if err == nil {
		t.Fatalf("Update() error = nil; want %v", ErrRecordNotFound)
	}

	if !errors.Is(err, ErrRecordNotFound) {
		t.Errorf("Update() error = %v; want %v", err, ErrRecordNotFound)
	}

}

func TestUpdateDuplicates(t *testing.T) {
	projectModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 1)

	handle := "handle"

	project := &Project{
		OwnerID: testUsers[0].ID,
		Handle:  handle,
		Name:    "test_project_1",
		Status:  StatusActive,
	}

	project2 := &Project{
		OwnerID: testUsers[0].ID,
		Handle:  "anotherHandle",
		Name:    "test_project_1",
		Status:  StatusActive,
	}

	insertedProject, err := projectModel.Insert(t.Context(), project)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := projectModel.Delete(context.Background(), insertedProject.ID)
		if err != nil {
			t.Errorf("Delete() error = %v; want nil", err)
		}
	})

	insertedProject2, err := projectModel.Insert(t.Context(), project2)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := projectModel.Delete(context.Background(), insertedProject2.ID)
		if err != nil {
			t.Errorf("Delete() error = %v; want nil", err)
		}
	})

	insertedProject2.Handle = handle

	_, err = projectModel.Update(t.Context(), insertedProject2)
	if err == nil {
		t.Fatalf("Update() error = nil; want %v", ErrDuplicateHandle)
	}

	if !errors.Is(err, ErrDuplicateHandle) {
		t.Errorf("Update() error = %v; want %v", err, ErrDuplicateHandle)
	}
}

func TestUpdateInvalidStatus(t *testing.T) {
	projectModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 1)

	project := &Project{
		OwnerID: testUsers[0].ID,
		Name:    "test_project_1",
		Status:  StatusActive,
	}

	insertedProject, err := projectModel.Insert(t.Context(), project)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := projectModel.Delete(context.Background(), insertedProject.ID)
		if err != nil {
			t.Errorf("Delete() error = %v; want nil", err)
		}
	})

	updatedHandle := "updatedHandle"
	updatedName := "Updated Name"
	updatedStatus := "invalid"

	project.ID = insertedProject.ID
	project.Handle = updatedHandle
	project.Name = updatedName
	project.Status = updatedStatus

	_, err = projectModel.Update(t.Context(), project)
	if err == nil {
		t.Fatalf("Update() error = nil; want %v", ErrInvalidStatus)
	}

	if !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("Update() error = %v; want %v", err, ErrInvalidStatus)
	}
}

func TestMapProjectError(t *testing.T) {
	plainErr := errors.New("plain error")
	unknownConstraintErr := &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "unknown_constraint",
	}
	unknownCodeErr := &pgconn.PgError{Code: "99999"}

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "duplicate handle",
			err: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "projects_handle_key",
			},
			want: ErrDuplicateHandle,
		},
		{
			name: "invalid status",
			err: &pgconn.PgError{
				Code:           "23514",
				ConstraintName: "project_status_check",
			},
			want: ErrInvalidStatus,
		},
		{
			name: "wrapped PostgreSQL error",
			err: fmt.Errorf("wrapped: %w", &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "projects_handle_key",
			}),
			want: ErrDuplicateHandle,
		},
		{
			name: "unknown constraint",
			err:  unknownConstraintErr,
			want: unknownConstraintErr,
		},
		{
			name: "unknown code",
			err:  unknownCodeErr,
			want: unknownCodeErr,
		},
		{
			name: "non-PostgreSQL error",
			err:  plainErr,
			want: plainErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mapProjectError(tc.err)
			if got != tc.want {
				t.Errorf("mapProjectError() error = %v; want %v", got, tc.want)
			}
		})
	}
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
