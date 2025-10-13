package store

import (
	"context"
	"database/sql"
	"errors"
)

var ErrRecordNotFound = errors.New("resource not found")

type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	Get(ctx context.Context, id int64) (*Post, error)
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
}

type Storage struct {
	Posts PostRepository

	Users UserRepository
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts: &PostsStore{db},
		Users: &UsersStore{db},
	}
}
