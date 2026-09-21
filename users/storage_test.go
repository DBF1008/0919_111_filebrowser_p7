package users

import (
	"sync"
	"testing"

	"github.com/spf13/afero"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

// Interface is implemented by storage
var _ Store = &Storage{}

type fakeBackend struct {
	mux   sync.Mutex
	users map[uint]*User
}

func (f *fakeBackend) GetBy(i interface{}) (*User, error) {
	f.mux.Lock()
	defer f.mux.Unlock()

	switch id := i.(type) {
	case uint:
		user, ok := f.users[id]
		if !ok {
			return nil, fberrors.ErrNotExist
		}
		return user, nil
	case string:
		for _, user := range f.users {
			if user.Username == id {
				return user, nil
			}
		}
		return nil, fberrors.ErrNotExist
	default:
		return nil, fberrors.ErrInvalidDataType
	}
}

func (f *fakeBackend) Gets() ([]*User, error) {
	f.mux.Lock()
	defer f.mux.Unlock()

	all := make([]*User, 0, len(f.users))
	for _, user := range f.users {
		all = append(all, user)
	}
	return all, nil
}

func (f *fakeBackend) Save(u *User) error {
	f.mux.Lock()
	defer f.mux.Unlock()

	f.users[u.ID] = u
	return nil
}

func (f *fakeBackend) Update(u *User, _ ...string) error {
	f.mux.Lock()
	defer f.mux.Unlock()

	f.users[u.ID] = u
	return nil
}

func (f *fakeBackend) DeleteByID(id uint) error {
	f.mux.Lock()
	defer f.mux.Unlock()

	delete(f.users, id)
	return nil
}

func (f *fakeBackend) DeleteByUsername(username string) error {
	f.mux.Lock()
	defer f.mux.Unlock()

	for id, user := range f.users {
		if user.Username == username {
			delete(f.users, id)
		}
	}
	return nil
}

func (f *fakeBackend) CountAdmins() (int, error) {
	f.mux.Lock()
	defer f.mux.Unlock()

	count := 0
	for _, user := range f.users {
		if user.Perm.Admin {
			count++
		}
	}
	return count, nil
}

func TestStorageLastUpdate(t *testing.T) {
	t.Parallel()

	store := NewStorage(&fakeBackend{users: map[uint]*User{}})
	user := &User{ID: 1, Username: "john", Password: "secret", Fs: &afero.MemMapFs{}}

	if got := store.LastUpdate(user.ID); got != 0 {
		t.Errorf("expected no last update for unknown user, got %d", got)
	}

	if err := store.Update(user, "Username"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := store.LastUpdate(user.ID); got == 0 {
		t.Error("expected last update to be recorded")
	}
}

func TestStorageConcurrentLastUpdate(t *testing.T) {
	t.Parallel()

	store := NewStorage(&fakeBackend{users: map[uint]*User{}})
	user := &User{ID: 1, Username: "john", Password: "secret", Fs: &afero.MemMapFs{}}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				if err := store.Update(user, "Username"); err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				_ = store.LastUpdate(user.ID)
			}
		}()
	}
	wg.Wait()

	if got := store.LastUpdate(user.ID); got == 0 {
		t.Error("expected last update to be recorded")
	}
}
