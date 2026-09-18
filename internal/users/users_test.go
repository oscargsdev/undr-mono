package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
	deleteAllUsers(t, model.db)
	t.Cleanup(func() { deleteAllUsers(t, model.db) })

	expectedUser := testUsers[0]
	user := expectedUser

	err := model.Insert(t.Context(), &user)
	if err != nil {
		t.Fatalf("did not expect error: %v", err)
	}

	if user.ID != expectedUser.ID {
		t.Fatalf("user id expected %q, got %q", expectedUser.ID, user.ID)
	}

	if user.Email != expectedUser.Email {
		t.Fatalf("user email expected %q, got %q", expectedUser.Email, user.Email)
	}

	if user.DisplayName != expectedUser.DisplayName {
		t.Fatalf("user display name expected %q, got %q", expectedUser.DisplayName, user.DisplayName)
	}

	if user.Status != expectedUser.Status {
		t.Fatalf("user status expected %q, got %q", expectedUser.Status, user.Status)
	}

	if user.CreatedAt.IsZero() {
		t.Fatal("expected non-zero value for user created at")
	}

	if user.UpdatedAt.IsZero() {
		t.Fatal("expected non-zero value for user updated at")
	}
}

func TestInsertDuplicatedUser(t *testing.T) {
	model := getModel(t)
	deleteAllUsers(t, model.db)
	t.Cleanup(func() { deleteAllUsers(t, model.db) })

	originalUser := testUsers[0]

	err := model.Insert(t.Context(), &originalUser)
	if err != nil {
		t.Fatalf("did not expect error: %v", err)
	}

	duplicateUsers := []struct {
		duplicateField string
		user           User
		expectedError  error
	}{
		{
			duplicateField: "ID",
			user: User{
				ID:          originalUser.ID,
				Email:       testUsers[1].Email,
				DisplayName: testUsers[1].DisplayName,
				Status:      testUsers[1].Status,
			},
			expectedError: ErrDuplicateID,
		},
		{
			duplicateField: "Email",
			user: User{
				ID:          testUsers[2].ID,
				Email:       originalUser.Email,
				DisplayName: testUsers[2].DisplayName,
				Status:      testUsers[2].Status,
			},
			expectedError: ErrDuplicateEmail,
		},
		{
			duplicateField: "Display name",
			user: User{
				ID:          testUsers[3].ID,
				Email:       testUsers[3].Email,
				DisplayName: originalUser.DisplayName,
				Status:      testUsers[3].Status,
			},
			expectedError: ErrDuplicateDisplayName,
		},
	}

	for _, tc := range duplicateUsers {
		t.Run(("Duplicate " + tc.duplicateField), func(t *testing.T) {
			duplicateUser := tc.user
			err := model.Insert(t.Context(), &duplicateUser)
			if err == nil {
				t.Fatal("expected error")
			}

			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("expected %q, got %q", tc.expectedError, err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	model := getModel(t)
	deleteAllUsers(t, model.db)
	t.Cleanup(func() { deleteAllUsers(t, model.db) })

	insertedUser := testUsers[0]

	err := model.Insert(t.Context(), &insertedUser)
	if err != nil {
		t.Fatalf("error while inserting user: %v", err)
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
				t.Fatalf("did not expect error: %v", err)
			}

			if retrievedUser.ID != insertedUser.ID {
				t.Fatalf("retrieved user id expected %q, got %q", insertedUser.ID, retrievedUser.ID)
			}

			if retrievedUser.Email != insertedUser.Email {
				t.Fatalf("retrieved user email expected %q, got %q", insertedUser.Email, retrievedUser.Email)
			}

			if retrievedUser.DisplayName != insertedUser.DisplayName {
				t.Fatalf("retrieved user display name expected %q, got %q", insertedUser.DisplayName, retrievedUser.DisplayName)
			}

			if retrievedUser.Status != insertedUser.Status {
				t.Fatalf("retrieved user status expected %q, got %q", insertedUser.Status, retrievedUser.Status)
			}

			if retrievedUser.CreatedAt.IsZero() {
				t.Fatal("retrieved user created at expected to be non-zero value")
			}

			if retrievedUser.UpdatedAt.IsZero() {
				t.Fatal("retrieved user updated at expected to be non-zero value")
			}

		})
	}
}

func TestGetInvalidID(t *testing.T) {
	model := NewUserModel(nil)

	_, err := model.Get(t.Context(), UserID(""))
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected %v, got %v", ErrInvalidID, err)
	}
}

func TestGetUserNotFound(t *testing.T) {
	model := getModel(t)
	deleteAllUsers(t, model.db)
	t.Cleanup(func() { deleteAllUsers(t, model.db) })

	_, err := model.Get(t.Context(), testUsers[0].ID)
	if err == nil {
		t.Fatal("expected error")
	} else {
		if !errors.Is(err, ErrRecordNotFound) {
			t.Fatalf("expected %v, got %v", ErrRecordNotFound, err)
		}
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

func deleteAllUsers(t testing.TB, db *pgxpool.Pool) {
	t.Helper()

	for _, user := range testUsers {
		deleteUser(t, db, user.ID)
	}
}

func deleteUser(t testing.TB, db *pgxpool.Pool, userID UserID) {
	t.Helper()

	query := `DELETE FROM users WHERE id = $1`

	queryCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.Exec(queryCtx, query, userID)
	if err != nil {
		t.Fatalf("error while deleting from users: %v", err)
	}
}
