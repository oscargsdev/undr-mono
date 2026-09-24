package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oscargsdev/undr-mono/internal/config"
	"github.com/oscargsdev/undr-mono/internal/database"
)

const activeStatus = "active"

func TestInsert(t *testing.T) {
	model := getModel(t)
	testUsers := generateUserFixtures(t, 1)

	expectedUser := testUsers[0]
	user := expectedUser

	insertedUser, err := model.Insert(t.Context(), &user)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := model.Delete(context.Background(), insertedUser.ID)
		if err != nil {
			t.Errorf("error while deleting test users: %v", err)
		}
	})

	if insertedUser.ID != expectedUser.ID {
		t.Errorf("Insert() ID = %q; want %q", insertedUser.ID, expectedUser.ID)
	}

	if insertedUser.Email != expectedUser.Email {
		t.Errorf("Insert() Email = %q; want %q", insertedUser.Email, expectedUser.Email)
	}

	if insertedUser.DisplayName != expectedUser.DisplayName {
		t.Errorf("Insert() DisplayName = %q; want %q", insertedUser.DisplayName, expectedUser.DisplayName)
	}

	if insertedUser.Status != expectedUser.Status {
		t.Errorf("Insert() Status = %q; want %q", insertedUser.Status, expectedUser.Status)
	}

	if insertedUser.CreatedAt.IsZero() {
		t.Error("Insert() CreatedAt is zero; want non-zero")
	}

	if insertedUser.UpdatedAt.IsZero() {
		t.Error("Insert() UpdatedAt is zero; want non-zero")
	}

	if !insertedUser.CreatedAt.Equal(insertedUser.UpdatedAt) {
		t.Errorf("Insert() CreatedAt = %v, UpdatedAt = %v; want equal timestamps", insertedUser.CreatedAt, user.UpdatedAt)
	}
}

func TestInsertDuplicatedUser(t *testing.T) {
	model := getModel(t)
	testUsers := generateUserFixtures(t, 4)

	originalUser := testUsers[0]

	_, err := model.Insert(t.Context(), &originalUser)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := model.Delete(context.Background(), originalUser.ID)
		if err != nil {
			t.Errorf("error while deleting test users: %v", err)
		}
	})

	duplicateIDUser := testUsers[1]
	duplicateIDUser.ID = originalUser.ID

	duplicateEmailUser := testUsers[2]
	duplicateEmailUser.Email = originalUser.Email

	duplicateDisplayNameUser := testUsers[3]
	duplicateDisplayNameUser.DisplayName = originalUser.DisplayName

	duplicateUsers := []struct {
		name          string
		user          User
		expectedError error
	}{
		{
			name:          "Duplicate ID",
			user:          duplicateIDUser,
			expectedError: ErrDuplicateID,
		},
		{
			name:          "Duplicate email",
			user:          duplicateEmailUser,
			expectedError: ErrDuplicateEmail,
		},
		{
			name:          "Duplicate display name",
			user:          duplicateDisplayNameUser,
			expectedError: ErrDuplicateDisplayName,
		},
	}

	for _, tc := range duplicateUsers {
		t.Run(tc.name, func(t *testing.T) {
			duplicateUser := tc.user
			_, err := model.Insert(t.Context(), &duplicateUser)
			if err == nil {
				t.Fatalf("Insert() error = nil; want %v", tc.expectedError)
			}

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("Insert() error = %v; want %v", err, tc.expectedError)
			}
		})
	}
}

func TestGet(t *testing.T) {
	model := getModel(t)
	testUsers := generateUserFixtures(t, 1)

	user := testUsers[0]

	insertedUser, err := model.Insert(t.Context(), &user)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := model.Delete(context.Background(), insertedUser.ID)
		if err != nil {
			t.Errorf("error while deleting test users: %v", err)
		}
	})

	uuidTypes := []struct {
		name string
		id   string
	}{
		{"Canonical UUID", string(insertedUser.ID)},
		{"URN UUID", "urn:uuid:" + string(insertedUser.ID)},
	}

	for _, tc := range uuidTypes {
		t.Run(tc.name, func(t *testing.T) {
			retrievedUser, err := model.Get(t.Context(), UserID(tc.id))
			if err != nil {
				t.Fatalf("Get() error = %v; want nil", err)
			}

			if retrievedUser.ID != insertedUser.ID {
				t.Errorf("Get() ID = %q; want %q", retrievedUser.ID, insertedUser.ID)
			}

			if retrievedUser.Email != insertedUser.Email {
				t.Errorf("Get() Email = %q; want %q", retrievedUser.Email, insertedUser.Email)
			}

			if retrievedUser.DisplayName != insertedUser.DisplayName {
				t.Errorf("Get() DisplayName = %q; want %q", retrievedUser.DisplayName, insertedUser.DisplayName)
			}

			if retrievedUser.Status != insertedUser.Status {
				t.Errorf("Get() Status = %q; want %q", retrievedUser.Status, insertedUser.Status)
			}

			if !retrievedUser.CreatedAt.Equal(insertedUser.CreatedAt) {
				t.Errorf("Get() CreatedAt = %v; want %v", retrievedUser.CreatedAt, insertedUser.CreatedAt)
			}

			if !retrievedUser.UpdatedAt.Equal(insertedUser.UpdatedAt) {
				t.Errorf("Get() UpdatedAt = %v; want %v", retrievedUser.UpdatedAt, insertedUser.UpdatedAt)
			}

		})
	}
}

func TestGetInvalidID(t *testing.T) {
	model := NewUserModel(nil)

	_, err := model.Get(t.Context(), UserID(""))
	if err == nil {
		t.Fatalf("Get() error = nil; want %v", ErrInvalidID)
	}

	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("Get() error = %v; want %v", err, ErrInvalidID)
	}
}

func TestGetUserNotFound(t *testing.T) {
	model := getModel(t)

	_, err := model.Get(t.Context(), UserID(uuid.NewString()))
	if err == nil {
		t.Fatalf("Get() error = nil; want %v", ErrRecordNotFound)
	}

	if !errors.Is(err, ErrRecordNotFound) {
		t.Errorf("Get() error = %v; want %v", err, ErrRecordNotFound)
	}
}

func TestDelete(t *testing.T) {
	model := getModel(t)
	testUsers := generateUserFixtures(t, 1)

	user := testUsers[0]

	insertedUser, err := model.Insert(t.Context(), &user)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := model.Delete(context.Background(), insertedUser.ID)
		if err != nil {
			t.Errorf("Delete() error = %v; want nil", err)
		}
	})

	err = model.Delete(t.Context(), insertedUser.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v; want nil", err)
	}

	_, err = model.Get(t.Context(), insertedUser.ID)
	if err == nil {
		t.Fatalf("Get() error = nil; want %v", ErrRecordNotFound)
	}

	if !errors.Is(err, ErrRecordNotFound) {
		t.Errorf("Get() error = %v; want %v", err, ErrRecordNotFound)
	}
}

func getModel(t testing.TB) *UserModel {
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

	return NewUserModel(db)
}

func generateUserFixtures(t testing.TB, n int) []User {
	t.Helper()

	testUsers := make([]User, 0, n)

	for range n {
		id := UserID(uuid.NewString())
		email := "user" + string(id) + "@mail.com"
		displayName := "User " + string(id)

		user := User{
			ID:          id,
			Email:       email,
			DisplayName: displayName,
			Status:      activeStatus,
		}

		testUsers = append(testUsers, user)
	}

	return testUsers
}
