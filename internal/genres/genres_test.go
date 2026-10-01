package genres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oscargsdev/undr-mono/internal/config"
	"github.com/oscargsdev/undr-mono/internal/database"
	"github.com/oscargsdev/undr-mono/internal/users"
)

func TestInsert(t *testing.T) {
	genreModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 3)

	testGenres := []struct {
		name  string
		genre Genre
	}{
		{"alt rock",
			Genre{
				Name:      "alt rock",
				CreatedBy: testUsers[0].ID,
			}},
		{"shoegaze",
			Genre{
				Name:      "shoegaze",
				CreatedBy: testUsers[1].ID,
			}},
		{"grunge",
			Genre{
				Name:      "grunge",
				CreatedBy: testUsers[2].ID,
			}},
	}

	for _, tc := range testGenres {
		t.Run(tc.name, func(t *testing.T) {
			newGenre := tc.genre
			insertedGenre, err := genreModel.Insert(t.Context(), &tc.genre)
			if err != nil {
				t.Fatalf("Insert() error = %v; want nil", err)
			}
			t.Cleanup(func() {
				err := genreModel.Delete(context.Background(), insertedGenre.ID)
				if err != nil {
					t.Errorf("Delete() error = %v; want nil", err)
				}
			})

			if insertedGenre.ID == 0 {
				t.Error("Insert() ID is zero; want non-zero")
			}

			if insertedGenre.Name != newGenre.Name {
				t.Errorf("Insert() Name = %q; want %q", insertedGenre.Name, newGenre.Name)
			}

			if insertedGenre.CreatedBy != newGenre.CreatedBy {
				t.Errorf("Insert() CreatedBy = %q; want %q", insertedGenre.CreatedBy, newGenre.CreatedBy)
			}

			if insertedGenre.CreatedAt.IsZero() {
				t.Error("Insert() CreatedAt is zero; want non-zero")
			}
		})
	}
}

func TestInsertDuplicate(t *testing.T) {
	genreModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 1)

	newGenre := Genre{
		Name:      "alt rock",
		CreatedBy: testUsers[0].ID,
	}

	duplicateGenre := newGenre

	insertedGenre, err := genreModel.Insert(t.Context(), &newGenre)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := genreModel.Delete(context.Background(), insertedGenre.ID)
		if err != nil {
			t.Errorf("Delete() error = %v; want nil", err)
		}
	})

	_, err = genreModel.Insert(t.Context(), &duplicateGenre)
	if err == nil {
		t.Fatalf("Insert() error = nil; want %v", ErrDuplicateGenre)
	}

	if !errors.Is(err, ErrDuplicateGenre) {
		t.Fatalf("Insert() error = %q; want %q", err, ErrDuplicateGenre)
	}

}

func TestGet(t *testing.T) {
	genreModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 1)

	newGenre := Genre{
		Name:      "alt rock",
		CreatedBy: testUsers[0].ID,
	}

	insertedGenre, err := genreModel.Insert(t.Context(), &newGenre)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := genreModel.Delete(context.Background(), insertedGenre.ID)
		if err != nil {
			t.Errorf("Delete() error = %v; want nil", err)
		}
	})

	retrievedGenre, err := genreModel.Get(t.Context(), insertedGenre.ID)
	if err != nil {
		t.Fatalf("Get() error = %v; want nil", err)
	}

	if retrievedGenre.ID != insertedGenre.ID {
		t.Errorf("Get() ID = %d; want %d", retrievedGenre.ID, insertedGenre.ID)
	}

	if retrievedGenre.Name != insertedGenre.Name {
		t.Errorf("Get() Name = %q; want %q", retrievedGenre.Name, insertedGenre.Name)
	}

	if retrievedGenre.CreatedBy != insertedGenre.CreatedBy {
		t.Errorf("Get() CreatedBy = %q; want %q", retrievedGenre.CreatedBy, insertedGenre.CreatedBy)
	}

	if !retrievedGenre.CreatedAt.Equal(insertedGenre.CreatedAt) {
		t.Errorf("Get() CreatedAt = %v; want %v", retrievedGenre.CreatedAt, insertedGenre.CreatedAt)
	}
}

func TestGetIDBelowOne(t *testing.T) {
	genreModel := NewGenreModel(nil)

	_, err := genreModel.Get(t.Context(), 0)
	if err == nil {
		t.Fatalf("Get() error = nil; want %v", ErrRecordNotFound)
	}

	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("Get() error = %q; want %q", err, ErrRecordNotFound)
	}
}

func TestGetNoGenre(t *testing.T) {
	genreModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 1)

	newGenre := Genre{
		Name:      "alt rock",
		CreatedBy: testUsers[0].ID,
	}

	insertedGenre, err := genreModel.Insert(t.Context(), &newGenre)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := genreModel.Delete(context.Background(), insertedGenre.ID)
		if err != nil {
			t.Errorf("Delete() error = %v; want nil", err)
		}
	})

	err = genreModel.Delete(t.Context(), insertedGenre.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v; want nil", err)
	}

	_, err = genreModel.Get(t.Context(), insertedGenre.ID)
	if err == nil {
		t.Fatalf("Get() error = nil; want %v", ErrRecordNotFound)
	}

	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("Get() error = %v; want %v", err, ErrRecordNotFound)
	}
}

func TestDelete(t *testing.T) {
	genreModel, userModel := getModels(t)
	testUsers := insertTestUsers(t, userModel, 1)

	genre := Genre{
		Name:      "alt rock",
		CreatedBy: testUsers[0].ID,
	}

	insertedGenre, err := genreModel.Insert(t.Context(), &genre)
	if err != nil {
		t.Fatalf("Insert() error = %v; want nil", err)
	}
	t.Cleanup(func() {
		err := genreModel.Delete(context.Background(), insertedGenre.ID)
		if err != nil {
			t.Errorf("Delete() error = %v; want nil", err)
		}
	})

	err = genreModel.Delete(t.Context(), insertedGenre.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v; want nil", err)
	}

	_, err = genreModel.Get(t.Context(), insertedGenre.ID)
	if err == nil {
		t.Fatalf("Get() error = nil; want %v", ErrRecordNotFound)
	}

	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("Get() error = %v; want %v", err, ErrRecordNotFound)
	}
}

func getModels(t testing.TB) (*GenreModel, *users.UserModel) {
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

	return NewGenreModel(db), users.NewUserModel(db)
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
