package store

import (
	"context"
	"database/sql"
)

type FollowerStore struct {
	db *sql.DB
}

type Followers struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

func (s *FollowerStore) FollowUser(ctx context.Context, userID int64, followerID int64) error {
	query := `
		INSERT INTO followers (user_id, follower_id)
			VALUES ($1,$2)
			ON CONFLICT DO NOTHING
		`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		return err
	}
	return nil
}

func (s *FollowerStore) UnfollowUser(ctx context.Context, userID int64, followerID int64) error {

	query := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
		`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		return err
	}
	return nil
}
