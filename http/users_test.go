package fbhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
)

func newUsersTestStorage(t *testing.T) (*storage.Storage, *users.User, *users.User) {
	t.Helper()

	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	})

	store, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to get storage: %v", err)
	}

	if err := store.Settings.Save(&settings.Settings{Key: []byte("key")}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	admin := &users.User{
		Username: "admin",
		Password: "secret",
		Perm:     users.Permissions{Admin: true},
	}
	if err := store.Users.Save(admin); err != nil {
		t.Fatalf("failed to save admin: %v", err)
	}

	regular := &users.User{
		Username: "regular",
		Password: "secret",
	}
	if err := store.Users.Save(regular); err != nil {
		t.Fatalf("failed to save regular user: %v", err)
	}

	return store, admin, regular
}

func signedUserToken(t *testing.T, store *storage.Storage, user *users.User) string {
	t.Helper()

	set, err := store.Settings.Get()
	if err != nil {
		t.Fatalf("failed to get settings: %v", err)
	}

	claims := &authToken{
		User: userInfo{ID: user.ID},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "File Browser",
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(set.Key)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func serveUserRequest(t *testing.T, handler handleFunc, store *storage.Storage, token string, urlVars map[string]string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/users", reader)
	if urlVars != nil {
		req = mux.SetURLVars(req, urlVars)
	}
	req.Header.Set("X-Auth", token)

	recorder := httptest.NewRecorder()
	handle(handler, "", store, &settings.Server{Root: t.TempDir()}).ServeHTTP(recorder, req)
	return recorder
}

func TestUserPostHandlerErrorFormat(t *testing.T) {
	t.Parallel()

	t.Run("non-empty which returns 400 with message", func(t *testing.T) {
		t.Parallel()

		store, admin, _ := newUsersTestStorage(t)
		recorder := serveUserRequest(t, userPostHandler, store, signedUserToken(t, store, admin), nil, &modifyUserRequest{
			modifyRequest: modifyRequest{What: "user", Which: []string{"all"}},
			Data:          &users.User{Username: "new", Password: "verysecretpassword"},
		})

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", recorder.Code)
		}
		if body := recorder.Body.String(); !strings.Contains(body, "invalid request params") {
			t.Errorf("expected error message in body, got %q", body)
		}
	})

	t.Run("empty password returns 400 with message", func(t *testing.T) {
		t.Parallel()

		store, admin, _ := newUsersTestStorage(t)
		recorder := serveUserRequest(t, userPostHandler, store, signedUserToken(t, store, admin), nil, &modifyUserRequest{
			modifyRequest: modifyRequest{What: "user"},
			Data:          &users.User{Username: "new"},
		})

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", recorder.Code)
		}
		if body := recorder.Body.String(); !strings.Contains(body, "password is empty") {
			t.Errorf("expected error message in body, got %q", body)
		}
	})

	t.Run("valid user returns 201", func(t *testing.T) {
		t.Parallel()

		store, admin, _ := newUsersTestStorage(t)
		recorder := serveUserRequest(t, userPostHandler, store, signedUserToken(t, store, admin), nil, &modifyUserRequest{
			modifyRequest: modifyRequest{What: "user"},
			Data:          &users.User{Username: "new", Password: "verysecretpassword"},
		})

		if recorder.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d (%s)", recorder.Code, recorder.Body.String())
		}
	})
}

func TestUserPutHandlerErrorFormat(t *testing.T) {
	t.Parallel()

	t.Run("id mismatch returns 400 with message", func(t *testing.T) {
		t.Parallel()

		store, admin, regular := newUsersTestStorage(t)
		recorder := serveUserRequest(t, userPutHandler, store, signedUserToken(t, store, admin),
			map[string]string{"id": "2"}, &modifyUserRequest{
				modifyRequest: modifyRequest{What: "user", Which: []string{"Locale"}},
				Data:          &users.User{ID: regular.ID + 100, Locale: "fr"},
			})

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", recorder.Code)
		}
		if body := recorder.Body.String(); !strings.Contains(body, "invalid request params") {
			t.Errorf("expected error message in body, got %q", body)
		}
	})

	t.Run("non-admin forbidden field returns 403 with message", func(t *testing.T) {
		t.Parallel()

		store, _, regular := newUsersTestStorage(t)
		recorder := serveUserRequest(t, userPutHandler, store, signedUserToken(t, store, regular),
			map[string]string{"id": "2"}, &modifyUserRequest{
				modifyRequest: modifyRequest{What: "user", Which: []string{"Username"}},
				Data:          &users.User{ID: regular.ID, Username: "renamed"},
			})

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", recorder.Code)
		}
		if body := recorder.Body.String(); !strings.Contains(body, "permission denied") {
			t.Errorf("expected error message in body, got %q", body)
		}
	})

	t.Run("admin update returns 200", func(t *testing.T) {
		t.Parallel()

		store, admin, regular := newUsersTestStorage(t)
		recorder := serveUserRequest(t, userPutHandler, store, signedUserToken(t, store, admin),
			map[string]string{"id": "2"}, &modifyUserRequest{
				modifyRequest: modifyRequest{What: "user", Which: []string{"Locale"}},
				Data:          &users.User{ID: regular.ID, Locale: "fr"},
			})

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d (%s)", recorder.Code, recorder.Body.String())
		}

		updated, err := store.Users.Get("", regular.ID)
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		if updated.Locale != "fr" {
			t.Errorf("expected locale %q, got %q", "fr", updated.Locale)
		}
	})
}
