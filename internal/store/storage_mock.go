package store

import (
	"context"
	"database/sql"
	"time"
)

func MockNewStorage() Storage {
	return Storage{

		Users: &MockUserStore{},
	}
}

type MockUserStore struct {
	GetUserbyIDFunc func(ctx context.Context, userID int64) (*User, error)
}

func (m *MockUserStore) Create(ctx context.Context, tx *sql.Tx, u *User) error {
	return nil
}

func (m *MockUserStore) GetUserbyID(ctx context.Context, userID int64) (*User, error) {
	return &User{ID: userID}, nil
}

func (m *MockUserStore) GetUserByEmail(context.Context, string) (*User, error) {
	return &User{}, nil
}

func (m *MockUserStore) CreateandInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return nil
}

func (m *MockUserStore) Activate(ctx context.Context, t string) error {
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, id int64) error {
	return nil
}
