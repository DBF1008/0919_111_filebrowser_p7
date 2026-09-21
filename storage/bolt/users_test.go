package bolt

import (
	"path/filepath"
	"testing"

	"github.com/asdine/storm/v3"

	"github.com/filebrowser/filebrowser/v2/users"
)

func openTestDB(t *testing.T) *storm.DB {
	t.Helper()

	db, err := storm.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	})
	return db
}

func TestUsersBackendUpdateFields(t *testing.T) {
	t.Parallel()

	backend := usersBackend{db: openTestDB(t)}

	user := &users.User{
		Username: "john",
		Password: "secret",
		Locale:   "en",
		ViewMode: "list",
	}
	if err := backend.Save(user); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	user.Locale = "fr"
	user.ViewMode = "mosaic"
	if err := backend.Update(user, "Locale", "ViewMode"); err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	got, err := backend.GetBy(user.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if got.Locale != "fr" {
		t.Errorf("expected locale %q, got %q", "fr", got.Locale)
	}
	if got.ViewMode != "mosaic" {
		t.Errorf("expected view mode %q, got %q", "mosaic", got.ViewMode)
	}
}

func TestUsersBackendUpdateInvalidFieldIsAtomic(t *testing.T) {
	t.Parallel()

	backend := usersBackend{db: openTestDB(t)}

	user := &users.User{
		Username: "john",
		Password: "secret",
		Locale:   "en",
	}
	if err := backend.Save(user); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	user.Locale = "fr"
	if err := backend.Update(user, "Locale", "DoesNotExist"); err == nil {
		t.Fatal("expected an error for an invalid field")
	}

	got, err := backend.GetBy(user.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if got.Locale != "en" {
		t.Errorf("expected no partial update, got locale %q", got.Locale)
	}
}

func TestUsersBackendUpdateWithoutFields(t *testing.T) {
	t.Parallel()

	backend := usersBackend{db: openTestDB(t)}

	user := &users.User{
		Username: "john",
		Password: "secret",
		Locale:   "en",
	}
	if err := backend.Save(user); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	user.Locale = "de"
	if err := backend.Update(user); err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	got, err := backend.GetBy(user.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if got.Locale != "de" {
		t.Errorf("expected locale %q, got %q", "de", got.Locale)
	}
}
