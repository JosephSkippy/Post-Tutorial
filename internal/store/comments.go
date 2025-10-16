package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64  `json:"id"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
}

type CommentStore struct {
	db *sql.DB
}

func (s *CommentStore) Create(ctx context.Context, comment *Comment) error {
	query := `
			INSERT INTO comments (post_id, user_id, comments)
			VALUES ($1, $2, $3) RETURNING id, created_at DESC
	`

	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.UserID,
		comment.Comment).Scan(
		&comment.ID,
		&comment.CreatedAt)

	if err != nil {
		return err
	}

	return nil

}

func (s *CommentStore) GetCommentByID(ctx context.Context, post_id int64) ([]Comment, error) {

	query := `
		SELECT
				c.id, 
				c.post_id,
				u.username,
				c.comments,
				c.created_at, 
				u.id
			FROM comments c
			JOIN users u ON u.id = c.user_id
			WHERE c.post_id = $1
			ORDER BY c.created_at DESC;
	`

	rows, err := s.db.QueryContext(ctx, query, post_id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		c := Comment{
			User: User{},
		}
		err := rows.Scan(&c.ID, &c.PostID, &c.User.Username, &c.Comment, &c.CreatedAt, &c.User.ID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}
