package users

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

func TestValidateAppliesDefaults(t *testing.T) {
	t.Parallel()

	u := &User{Username: "john", Password: "secret"}
	if err := u.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if u.ViewMode != ListViewMode {
		t.Errorf("expected view mode %q, got %q", ListViewMode, u.ViewMode)
	}
	if u.Commands == nil {
		t.Error("expected commands to be initialized")
	}
	if u.Sorting.By != "name" {
		t.Errorf("expected sorting by %q, got %q", "name", u.Sorting.By)
	}
	if u.Rules == nil {
		t.Error("expected rules to be initialized")
	}
}

func TestValidateRequiredFields(t *testing.T) {
	t.Parallel()

	u := &User{Password: "secret"}
	if err := u.Validate(); !errors.Is(err, fberrors.ErrEmptyUsername) {
		t.Errorf("expected ErrEmptyUsername, got %v", err)
	}

	u = &User{Username: "john"}
	if err := u.Validate(); !errors.Is(err, fberrors.ErrEmptyPassword) {
		t.Errorf("expected ErrEmptyPassword, got %v", err)
	}
}

func TestValidateSelectiveFields(t *testing.T) {
	t.Parallel()

	u := &User{Username: "john"}
	if err := u.Validate("Username"); err != nil {
		t.Fatalf("expected no error validating only username, got %v", err)
	}
	if u.ViewMode != "" {
		t.Errorf("expected view mode to be untouched, got %q", u.ViewMode)
	}

	u = &User{}
	if err := u.Validate("Password"); !errors.Is(err, fberrors.ErrEmptyPassword) {
		t.Errorf("expected ErrEmptyPassword, got %v", err)
	}
}

func TestValidateDoesNotInitFs(t *testing.T) {
	t.Parallel()

	u := &User{Username: "john", Password: "secret"}
	if err := u.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Fs != nil {
		t.Error("expected Validate to leave Fs unset")
	}
}

func TestInitSetsFs(t *testing.T) {
	t.Parallel()

	u := &User{Scope: "sub"}
	u.Init("/base")
	if u.Fs == nil {
		t.Fatal("expected Fs to be initialized")
	}

	if got, want := u.FullPath("file.txt"), "/base/sub/file.txt"; got != want {
		t.Errorf("expected full path %q, got %q", want, got)
	}
}

func TestInitKeepsExistingFs(t *testing.T) {
	t.Parallel()

	mem := &afero.MemMapFs{}
	u := &User{Scope: "sub", Fs: mem}
	u.Init("/base")
	if u.Fs != mem {
		t.Error("expected Init to keep the already set filesystem")
	}
}
