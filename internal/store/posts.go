package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserId    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentsCount int `json:"comments_count"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (content, title, user_id, tags) 
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserId,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return err
	}
	return nil
}

func (s *PostStore) GetById(ctx context.Context, postId int64) (Post, error) {
	query := `
		SELECT id, user_id, title, content, created_at, updated_at, tags, version
		FROM posts 
		WHERE id = $1
	`

	var post Post
	err := s.db.QueryRowContext(ctx, query, postId).Scan(
		&post.ID,
		&post.UserId,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return Post{}, ErrorNotFound
		default:
			return Post{}, err
		}
	}

	return post, nil
}

func (s *PostStore) Delete(ctx context.Context, postId int64) error {
	query := `DELETE FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, postId)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts 
		SET content = $1, 
		title = $2, 
		tags = $3, 
		updated_at = now(), -- Poprawiono, usunięto błędną deklarację DEFAULT
		version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING version
	`
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		pq.Array(post.Tags),
		post.ID,
		post.Version,
	).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrorNotFound
		default:
			return err
		}
	}
	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userId int64, fq PaginatedFeedQuery) ([]PostWithMetadata, error) {
	query := `
		select  p.id, p.title, p.user_id, p.content, p.created_at, p.tags, p.updated_at ,p.version
			,COUNT(c.id) AS comments_count
			,u.username
		from public.posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
		WHERE 
   			f.user_id = $1 AND 
    		((p.title ILIKE '%' || $4 || '%') OR (p.content ILIKE '%' || $4 || '%')) AND
    		(p.tags @> $5 OR $5 = '{}') AND
    		((p.created_at >= $6 OR $6 IS NULL) AND (p.created_at <= $7 OR $7 IS NULL))
		GROUP BY p.id, u.username
		ORDER BY p.created_at ` + fq.Sort + `
		LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userId, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags), fq.Since, fq.Until)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []PostWithMetadata

	for rows.Next() {
		var p PostWithMetadata
		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.UserId,
			&p.Content,
			&p.CreatedAt,
			pq.Array(&p.Tags),
			&p.UpdatedAt,
			&p.Version,
			&p.CommentsCount,
			&p.User.Username,
		)

		if err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}

	return feed, nil
}
