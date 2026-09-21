package users

import (
	"testing"

	"github.com/spf13/afero"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

func TestValidate(t *testing.T) {
	u := &User{}
	if err := u.Validate(); err != fberrors.ErrEmptyUsername {
		t.Fatalf("expected ErrEmptyUsername, got %v", err)
	}

	u.Username = "test"
	if err := u.Validate(); err != fberrors.ErrEmptyPassword {
		t.Fatalf("expected ErrEmptyPassword, got %v", err)
	}

	u.Password = "password"
	if err := u.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if u.ViewMode != ListViewMode {
		t.Errorf("expected default view mode %q, got %q", ListViewMode, u.ViewMode)
	}
	if u.Commands == nil {
		t.Error("expected commands to be initialized")
	}
	if u.Sorting.By != "name" {
		t.Errorf("expected default sorting by name, got %q", u.Sorting.By)
	}
	if u.Rules == nil {
		t.Error("expected rules to be initialized")
	}
}

func TestValidateSpecificFields(t *testing.T) {
	u := &User{}
	if err := u.Validate("Username"); err != fberrors.ErrEmptyUsername {
		t.Fatalf("expected ErrEmptyUsername, got %v", err)
	}

	// Password is not checked when only Username is requested.
	u.Username = "test"
	if err := u.Validate("Username"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestInitSetsFs(t *testing.T) {
	u := &User{Scope: "/foo"}
	if err := u.Init("/base"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Fs == nil {
		t.Fatal("expected Fs to be initialized")
	}
}

func TestInitDoesNotOverwriteFs(t *testing.T) {
	existing := afero.NewMemMapFs()
	u := &User{Scope: "/foo", Fs: existing}
	if err := u.Init("/base"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Fs != existing {
		t.Error("expected Init to keep the already set Fs")
	}
}

func TestCleanRunsValidateAndInit(t *testing.T) {
	u := &User{Username: "test", Password: "password", Scope: "/foo"}
	if err := u.Clean("/base"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Fs == nil {
		t.Error("expected Clean to initialize Fs")
	}
	if u.ViewMode != ListViewMode {
		t.Error("expected Clean to apply validation defaults")
	}

	dirty := &User{}
	if err := dirty.Clean("/base"); err != fberrors.ErrEmptyUsername {
		t.Fatalf("expected ErrEmptyUsername, got %v", err)
	}
}
