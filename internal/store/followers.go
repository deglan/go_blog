package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Follower struct {
	UserId     int64  `json:"user_id"`
	FollowerId int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) FollowUser(ctx context.Context, followerId, userId int64) error {
	query := `
		INSERT INTO followers (user_id, follower_id) 
		VALUES ($1, $2)
	`

	_, err := s.db.ExecContext(ctx, query, followerId, userId)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrorAlreadyExists
		}
	}
	return err
}

func (s *FollowerStore) UnfollowUser(ctx context.Context, followerId, userId int64) error {
	query := `
	DELETE FROM followers 
	WHERE user_id = $1 AND follower_id = $2
`

	_, err := s.db.ExecContext(ctx, query, followerId, userId)
	return err
}
