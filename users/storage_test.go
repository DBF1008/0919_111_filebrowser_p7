package users

import (
	"sync"
	"testing"
)

// Interface is implemented by storage
var _ Store = &Storage{}

type mockStorageBackend struct{}

func (m *mockStorageBackend) GetBy(interface{}) (*User, error)      { return &User{}, nil }
func (m *mockStorageBackend) Gets() ([]*User, error)                { return nil, nil }
func (m *mockStorageBackend) Save(u *User) error                    { return nil }
func (m *mockStorageBackend) Update(u *User, fields ...string) error { return nil }
func (m *mockStorageBackend) DeleteByID(uint) error                 { return nil }
func (m *mockStorageBackend) DeleteByUsername(string) error         { return nil }
func (m *mockStorageBackend) CountAdmins() (int, error)             { return 0, nil }

func TestLastUpdate(t *testing.T) {
	s := NewStorage(&mockStorageBackend{})

	if got := s.LastUpdate(1); got != 0 {
		t.Fatalf("expected 0 for never updated user, got %d", got)
	}

	user := &User{ID: 1, Username: "test", Password: "password"}
	if err := s.Update(user, "Username"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := s.LastUpdate(1); got == 0 {
		t.Fatal("expected non-zero timestamp after update")
	}
}

// TestLastUpdateConcurrent exercises LastUpdate and Update
// concurrently. It is meant to be run with the race detector
// enabled (go test -race).
func TestLastUpdateConcurrent(t *testing.T) {
	s := NewStorage(&mockStorageBackend{})
	user := &User{ID: 1, Username: "test", Password: "password"}

	// Initialize the user's filesystem upfront so the concurrent
	// updates below only touch the storage's internal state.
	if err := s.Update(user, "Username"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = s.Update(user, "Username")
			}
		}()
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = s.LastUpdate(1)
			}
		}()
	}
	wg.Wait()
}
