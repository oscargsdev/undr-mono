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

var (
	ids = []string{"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
		"33333333-3333-3333-3333-333333333333",
		"44444444-4444-4444-4444-444444444444"}
)

func TestInsert(t *testing.T) {
	model := getModel(t)
	deleteAllUsers(t, model.db)
	t.Cleanup(func() { deleteAllUsers(t, model.db) })

	var userID UserID = UserID(ids[0])
	email := "user1@mail.com"
	displayName := "User 1"
	status := "active"

	user := &User{
		ID:          userID,
		Email:       email,
		DisplayName: displayName,
		Status:      status,
	}

	err := model.Insert(t.Context(), user)
	if err != nil {
		t.Fatalf("did not expect error: %v", err)
	}

	if user.ID != userID {
		t.Fatalf("user id expected %q, got %q", userID, user.ID)
	}

	if user.Email != email {
		t.Fatalf("user email expected %q, got %q", email, user.Email)
	}

	if user.DisplayName != displayName {
		t.Fatalf("user display name expected %q, got %q", displayName, user.DisplayName)
	}

	if user.Status != status {
		t.Fatalf("user status expected %q, got %q", status, user.Status)
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

	var userID UserID = UserID(ids[0])
	email := "user1@mail.com"
	displayName := "User 1"
	status := "active"

	user := &User{
		ID:          userID,
		Email:       email,
		DisplayName: displayName,
		Status:      status,
	}

	err := model.Insert(t.Context(), user)
	if err != nil {
		t.Fatalf("did not expect error: %v", err)
	}
	t.Cleanup(func() { deleteUser(t, model.db, user.ID) })

	duplicateUsers := []struct {
		duplicateField string
		userID         UserID
		email          string
		displayName    string
		status         string
		expectedError  error
	}{
		{"ID", userID, "user2@mail.com", "User 2", status, ErrDuplicateID},
		{"Email", UserID(ids[2]), email, "User 3", status, ErrDuplicateEmail},
		{"Display name", UserID(ids[3]), "user4@mail.com", displayName, status, ErrDuplicateDisplayName},
	}

	for _, tc := range duplicateUsers {
		t.Run(("Duplicate " + tc.duplicateField), func(t *testing.T) {
			duplicateUser := &User{
				ID:          tc.userID,
				Email:       tc.email,
				DisplayName: tc.displayName,
				Status:      tc.status,
			}
			err := model.Insert(t.Context(), duplicateUser)
			if err == nil {
				t.Fatal("expected error")
			}

			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("expected %q, got %q", tc.expectedError, err)
			}
		})
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

	for _, id := range ids {
		deleteUser(t, db, UserID(id))
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
