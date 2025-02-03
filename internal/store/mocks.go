package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/stretchr/testify/mock"
)

func NewMockStore() Storage {
	return Storage{
		Users:    &MockUserStore{},
		Comments: &MockCommentsStore{},
		Posts:    &MockPostStore{},
		Roles:    &MockRolesStore{},
	}
}

type MockUserStore struct {
	mock.Mock
}

type MockCommentsStore struct {
	mock.Mock
}

type MockPostStore struct {
	mock.Mock
}

type MockRolesStore struct {
	mock.Mock
}

func (m *MockUserStore) Create(ctx context.Context, tx *sql.Tx, u *User) error {
	return nil
}

func (m *MockUserStore) GetById(ctx context.Context, userID int64) (*User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), nil
}

func (m *MockUserStore) GetByEmail(context.Context, string) (*User, error) {
	return &User{
		ID:       1,
		Username: "test",
		Email:    "K7iGd@example.com",
		Password: Password{
			hash: []byte("$2a$10$nfinUN8NKvyEq0hbFkzYYO3UMxRLM5.X3Xp1BAw/G1G14Fdci6DHS"),
			text: &[]string{"test"}[0],
		},
		IsActive: true,
		RoleId:   1,
	}, nil
}

func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return nil
}

func (m *MockUserStore) Activate(ctx context.Context, t string) error {
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *MockUserStore) VerifyPassword(plainPassword, hashedPassword string) (bool, error) {
	args := m.Called(plainPassword, hashedPassword)
	return args.Bool(0), args.Error(1)
}

func (c *MockCommentsStore) GetByPostId(ctx context.Context, postId int64) ([]Comment, error) {
	args := c.Called(ctx, postId)
	return args.Get(0).([]Comment), args.Error(1)
}

func (c *MockCommentsStore) Create(ctx context.Context, comment *Comment) error {
	return nil
}

func (c *MockCommentsStore) Update(ctx context.Context, comment *Comment) error {
	return nil
}

func (c *MockCommentsStore) Delete(ctx context.Context, commentId, userId int64) error {
	args := c.Called(ctx, commentId, userId)
	return args.Error(0)
}

func (c *MockCommentsStore) GetById(ctx context.Context, id int64) (*Comment, error) {
	args := c.Called(ctx, id)
	return args.Get(0).(*Comment), args.Error(1)
}

func (p *MockPostStore) Create(ctx context.Context, post *Post) error {
	args := p.Called(ctx, post)
	return args.Error(0)
}

func (p *MockPostStore) GetById(ctx context.Context, postId int64) (Post, error) {
	args := p.Called(ctx, postId)
	return args.Get(0).(Post), args.Error(1)
}

func (p *MockPostStore) Delete(ctx context.Context, postId int64) error {
	args := p.Called(ctx, postId)
	return args.Error(0)
}

func (p *MockPostStore) Update(ctx context.Context, post *Post) error {
	args := p.Called(ctx, post)
	return args.Error(0)
}

func (p *MockPostStore) GetUserFeed(ctx context.Context, userId int64, fq PaginatedFeedQuery) ([]PostWithMetadata, error) {
	args := p.Called(ctx, userId, fq)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PostWithMetadata), args.Error(1)
}

func (r *MockRolesStore) GetByName(ctx context.Context, name string) (*Role, error) {
	args := r.Called(ctx, name)
	return args.Get(0).(*Role), args.Error(1)
}
