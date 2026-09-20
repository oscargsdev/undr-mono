package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oscargsdev/undr-mono/internal/config"
	"github.com/oscargsdev/undr-mono/internal/database"
)

const activeStatus = "active"

var testUsers = []User{
	{
		ID:          "11111111-1111-1111-1111-111111111111",
		Email:       "user1@mail.com",
		DisplayName: "Display Name 1",
		Status:      activeStatus,
	},
	{
		ID:          "22222222-2222-2222-2222-222222222222",
		Email:       "user2@mail.com",
		DisplayName: "Display Name 2",
		Status:      activeStatus,
	},
	{
		ID:          "33333333-3333-3333-3333-333333333333",
		Email:       "user3@mail.com",
		DisplayName: "Display Name 3",
		Status:      activeStatus,
	},
	{
		ID:          "44444444-4444-4444-4444-444444444444",
		Email:       "user4@mail.com",
		DisplayName: "Display Name 4",
		Status:      activeStatus,
	},
}

func TestInsert(t *testing.T) {
	model := getModel(t)
	deleteTestUsers(t, model)
	t.Cleanup(func() { deleteTestUsers(t, model) })

	expectedUser := testUsers[0]
	user := expectedUser

	insertedUser, err := model.Insert(t.Context(), &user)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}

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
	deleteTestUsers(t, model)
	t.Cleanup(func() { deleteTestUsers(t, model) })

	originalUser := testUsers[0]

	_, err := model.Insert(t.Context(), &originalUser)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}

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
	deleteTestUsers(t, model)
	t.Cleanup(func() { deleteTestUsers(t, model) })

	user := testUsers[0]

	insertedUser, err := model.Insert(t.Context(), &user)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}

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
	deleteTestUsers(t, model)
	t.Cleanup(func() { deleteTestUsers(t, model) })

	_, err := model.Get(t.Context(), testUsers[0].ID)
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

func deleteTestUsers(t testing.TB, m *UserModel) {
	t.Helper()

	for _, user := range testUsers {
		m.Delete(t.Context(), user.ID)
	}
}

// func deleteUser(t testing.TB, db *pgxpool.Pool, userID UserID) {
// 	t.Helper()

// 	query := `DELETE FROM users WHERE id = $1`

// 	queryCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	_, err := db.Exec(queryCtx, query, userID)
// 	if err != nil {
// 		t.Fatalf("error while deleting from users: %v", err)
// 	}
// }
