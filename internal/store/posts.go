package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Version   int64     `json:"version"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type Feed struct {
	Post         Post
	CommentCount int64 `json:"comment_count"`
}

type PostsStore struct {
	db *sql.DB
}

func (s *PostsStore) Create(ctx context.Context, post *Post) error {
	query := `
			INSERT INTO posts (content, title, user_id, tags)
			VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags)).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostsStore) Get(ctx context.Context, id int64) (*Post, error) {
	query := `
			SELECT id, content, title, user_id, tags, created_at, updated_at, version
			FROM posts
			WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostsStore) DeletePost(ctx context.Context, id int64) (*Post, error) {

	query := `
			DELETE FROM posts WHERE id = $1 RETURNING title;
			`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(&post.Title)

	if err != nil {

		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &post, nil

}

func (s *PostsStore) UpdatePost(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content=$2, version= version + 1
		WHERE id = $3 AND version=$4
		RETURNING version
		;
		`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID, post.Version).Scan(&post.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	return nil
}

func (s *PostsStore) GetFeed(ctx context.Context, user_id int64, fq PaginatedFeedQuery) (*[]Feed, error) {

	query := `
SELECT
  p.id,
  p.user_id,
  p.title,
  p.content,
  p.created_at,
  p.version,
  p.tags,
  u.username,
  COUNT(c.id) AS comments_count
FROM
  posts p
  JOIN users u ON u.id = p.user_id
  LEFT JOIN followers f ON f.follower_id = p.user_id -- author
  AND f.user_id = $1 -- viewer only
  LEFT JOIN comments c ON c.post_id = p.id
WHERE
  (p.user_id = $1 -- my posts
  OR f.user_id IS NOT NULL
  )
  AND
  (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
  AND
  (p.tags @> $5 OR $5 = '{}')
GROUP BY
  p.id,
  u.username
ORDER BY
  p.created_at ` + fq.Sort + `
LIMIT $2
OFFSET $3;
`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	row, err := s.db.QueryContext(ctx, query, user_id, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	defer row.Close()

	var feeds []Feed
	for row.Next() {
		var post Post
		var feed Feed
		err := row.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.Version,
			pq.Array(&post.Tags),
			&post.User.Username,
			&feed.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		feed.Post = post
		feeds = append(feeds, feed)
	}

	return &feeds, nil
}
