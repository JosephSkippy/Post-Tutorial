package store

import (
	"database/sql"
)

type Comment struct {
	ID        int64
	PostID    int64
	UserID    int64
	Comments  string
	CreatedAt string
}

type CommentStore struct {
	db *sql.DB
}
