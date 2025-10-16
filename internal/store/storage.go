package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound    = errors.New("resource not found")
	QueryTimeOutDuration = time.Second * 5
)

type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	Get(ctx context.Context, id int64) (*Post, error)
	DeletePost(ctx context.Context, id int64) (*Post, error)
	UpdatePost(ctx context.Context, post *Post) error
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
}

type CommentRepository interface {
	Create(ctx context.Context, comment *Comment) error
	GetCommentByID(ctx context.Context, post_id int64) ([]Comment, error)
	DeleteCommentByPostID(ctx context.Context, post_id int64) error
}

type Storage struct {
	Posts   PostRepository
	Users   UserRepository
	Comment CommentRepository
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:   &PostsStore{db},
		Users:   &UsersStore{db},
		Comment: &CommentStore{db},
	}
}
