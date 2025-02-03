package store

import (
	"context"
	"database/sql"
	"time"
)

type Comment struct {
	Id        int64     `json:"id"`
	PostId    int64     `json:"post_id"`
	UserId    int64     `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user"`
}

type CommentsStore struct {
	db *sql.DB
}

func (s *CommentsStore) GetById(ctx context.Context, id int64) (*Comment, error) {
	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, users.username
		FROM comments c
		JOIN users ON c.user_id = users.id
		WHERE c.id = $1
	`

	var c Comment
	c.User = User{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&c.Id,
		&c.PostId,
		&c.UserId,
		&c.Content,
		&c.CreatedAt,
		&c.User.Username,
	)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (s *CommentsStore) GetByPostId(ctx context.Context, postId int64) ([]Comment, error) {
	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, users.username
		FROM comments c
		JOIN users ON c.user_id = users.id
		WHERE c.post_id = $1
		ORDER BY c.created_at DESC
 	`
	rows, err := s.db.QueryContext(ctx, query, postId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []Comment{}

	for rows.Next() {
		var c Comment
		c.User = User{}
		err := rows.Scan(
			&c.Id,
			&c.PostId,
			&c.UserId,
			&c.Content,
			&c.CreatedAt,
			&c.User.Username,
		)

		if err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	return comments, nil
}

func (s *CommentsStore) Create(ctx context.Context, comment *Comment) error {
	query := `
	INSERT INTO comments (post_id, user_id, content) 
	VALUES ($1, $2, $3) 
	RETURNING id, created_at
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.PostId,
		comment.UserId,
		comment.Content,
	).Scan(
		&comment.Id,
		&comment.CreatedAt,
	)

	if err != nil {
		return err
	}
	return nil
}

func (s *CommentsStore) Update(ctx context.Context, comment *Comment) error {
	query := `
		UPDATE comments
		SET content = $1
		WHERE id = $2 AND user_id = $3
		RETURNING updated_at
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var updatedAt time.Time
	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.Content,
		comment.Id,
		comment.UserId,
	).Scan(&updatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *CommentsStore) Delete(ctx context.Context, commentId, userId int64) error {
	query := `
		DELETE FROM comments
		WHERE id = $1 AND user_id = $2
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, commentId, userId)
	return err
}
