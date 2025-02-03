package main

import (
	"net/http"
	"social/internal/store"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestCheckPostOwnership(t *testing.T) {
	withRedis := config{
		redisCfg: redisConfig{
			enabled: false,
		},
	}
	app := newTestApplication(t, withRedis)
	mux := app.mount()

	testToken, err := app.authenticator.GenerateToken(nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("should allow delete post if authenticated user is the owner", func(t *testing.T) {
		mockPostsStore := new(store.MockPostStore)
		mockUserStore := new(store.MockUserStore)
		app.store.Posts = mockPostsStore
		app.store.Users = mockUserStore

		mockPostsStore.ExpectedCalls = nil
		mockUserStore.ExpectedCalls = nil

		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
			RoleId:   2,
		}

		post := &store.Post{
			ID:        1,
			UserId:    1,
			Title:     "Test Post",
			Content:   "This is a test post",
			Tags:      []string{"tag1", "tag2"},
			CreatedAt: "2022-01-01T00:00:00Z",
			UpdatedAt: "2022-01-01T00:00:00Z",
			Version:   1,
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil)
		mockPostsStore.On("GetById", mock.Anything, int64(1)).Return(*post, nil)
		mockPostsStore.On("Delete", mock.Anything, int64(1)).Return(nil)

		req, err := http.NewRequest(http.MethodDelete, "/v1/posts/1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusNoContent, rr.Code)

		mockPostsStore.AssertExpectations(t)
		mockUserStore.AssertExpectations(t)
	})

	t.Run("should allow delete post if authenticated user is the admin", func(t *testing.T) {
		mockPostsStore := new(store.MockPostStore)
		mockUserStore := new(store.MockUserStore)
		mockRoleStore := new(store.MockRolesStore)
		app.store.Posts = mockPostsStore
		app.store.Users = mockUserStore
		app.store.Roles = mockRoleStore

		mockPostsStore.ExpectedCalls = nil
		mockUserStore.ExpectedCalls = nil
		mockRoleStore.ExpectedCalls = nil

		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
			RoleId:   3,
		}

		role := store.Role{
			Id:   3,
			Name: "admin",
		}

		user.Role = role

		post := &store.Post{
			ID:        1,
			UserId:    2,
			Title:     "Test Post",
			Content:   "This is a test post",
			Tags:      []string{"tag1", "tag2"},
			CreatedAt: "2022-01-01T00:00:00Z",
			UpdatedAt: "2022-01-01T00:00:00Z",
			Version:   1,
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil)
		mockPostsStore.On("GetById", mock.Anything, int64(1)).Return(*post, nil)
		mockPostsStore.On("Delete", mock.Anything, int64(1)).Return(nil)
		mockRoleStore.On("GetByName", mock.Anything, mock.Anything).Return(&role, nil)

		req, err := http.NewRequest(http.MethodDelete, "/v1/posts/1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusNoContent, rr.Code)

		mockPostsStore.AssertExpectations(t)
		mockUserStore.AssertExpectations(t)
		mockRoleStore.AssertExpectations(t)
	})

	t.Run("should not allow delete post if authenticated user is not owner and don't have required role", func(t *testing.T) {
		mockPostsStore := new(store.MockPostStore)
		mockUserStore := new(store.MockUserStore)
		mockRoleStore := new(store.MockRolesStore)
		app.store.Posts = mockPostsStore
		app.store.Users = mockUserStore
		app.store.Roles = mockRoleStore

		mockPostsStore.ExpectedCalls = nil
		mockUserStore.ExpectedCalls = nil
		mockRoleStore.ExpectedCalls = nil

		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
			RoleId:   1,
		}

		role := store.Role{
			Id:    1,
			Name:  "user",
			Level: 1,
		}

		requiredRole := store.Role{
			Id:    3,
			Name:  "admin",
			Level: 3,
		}

		user.Role = role

		post := &store.Post{
			ID:        1,
			UserId:    2,
			Title:     "Test Post",
			Content:   "This is a test post",
			Tags:      []string{"tag1", "tag2"},
			CreatedAt: "2022-01-01T00:00:00Z",
			UpdatedAt: "2022-01-01T00:00:00Z",
			Version:   1,
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil)
		mockPostsStore.On("GetById", mock.Anything, int64(1)).Return(*post, nil)
		mockRoleStore.On("GetByName", mock.Anything, mock.Anything).Return(&requiredRole, nil)

		req, err := http.NewRequest(http.MethodDelete, "/v1/posts/1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusForbidden, rr.Code)

		mockPostsStore.AssertExpectations(t)
		mockUserStore.AssertExpectations(t)
		mockRoleStore.AssertExpectations(t)
	})
}

func TestCheckCommentOwnership(t *testing.T) {
	withRedis := config{
		redisCfg: redisConfig{
			enabled: false,
		},
	}
	app := newTestApplication(t, withRedis)
	mux := app.mount()

	testToken, err := app.authenticator.GenerateToken(nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("should allow delete comment if authenticated user is the owner", func(t *testing.T) {
		mockCommentsStore := new(store.MockCommentsStore)
		mockUserStore := new(store.MockUserStore)
		app.store.Comments = mockCommentsStore
		app.store.Users = mockUserStore

		mockCommentsStore.ExpectedCalls = nil
		mockUserStore.ExpectedCalls = nil

		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
			RoleId:   2,
		}

		comment := &store.Comment{
			Id:      1,
			UserId:  1,
			PostId:  1,
			Content: "This is a test comment",
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil)
		mockCommentsStore.On("GetById", mock.Anything, int64(1)).Return(comment, nil)
		mockCommentsStore.On("Delete", mock.Anything, int64(1), int64(1)).Return(nil)

		req, err := http.NewRequest(http.MethodDelete, "/v1/comments/1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusNoContent, rr.Code)

		mockCommentsStore.AssertExpectations(t)
		mockUserStore.AssertExpectations(t)
	})

	t.Run("should allow delete comment if authenticated user is the admin", func(t *testing.T) {
		mockCommentsStore := new(store.MockCommentsStore)
		mockUserStore := new(store.MockUserStore)
		mockRoleStore := new(store.MockRolesStore)
		app.store.Comments = mockCommentsStore
		app.store.Users = mockUserStore
		app.store.Roles = mockRoleStore

		mockCommentsStore.ExpectedCalls = nil
		mockUserStore.ExpectedCalls = nil
		mockRoleStore.ExpectedCalls = nil

		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
			RoleId:   3,
		}

		role := store.Role{
			Id:   3,
			Name: "admin",
		}

		user.Role = role

		comment := &store.Comment{
			Id:      1,
			UserId:  2,
			PostId:  1,
			Content: "This is a test comment",
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil)
		mockCommentsStore.On("GetById", mock.Anything, mock.Anything).Return(comment, nil)
		mockRoleStore.On("GetByName", mock.Anything, "admin").Return(&role, nil)
		mockCommentsStore.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodDelete, "/v1/comments/1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusNoContent, rr.Code)
	})

	t.Run("should not allow delete post if authenticated user is not owner and don't have required role", func(t *testing.T) {
		mockCommentsStore := new(store.MockCommentsStore)
		mockUserStore := new(store.MockUserStore)
		mockRoleStore := new(store.MockRolesStore)

		app.store.Comments = mockCommentsStore
		app.store.Users = mockUserStore
		app.store.Roles = mockRoleStore

		mockCommentsStore.ExpectedCalls = nil
		mockUserStore.ExpectedCalls = nil
		mockRoleStore.ExpectedCalls = nil

		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
			RoleId:   1,
		}

		role := store.Role{
			Id:    1,
			Name:  "user",
			Level: 1,
		}

		requiredRole := store.Role{
			Id:    3,
			Name:  "admin",
			Level: 3,
		}

		user.Role = role

		comment := &store.Comment{
			Id:      1,
			UserId:  2,
			PostId:  1,
			Content: "This is a test comment",
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil)
		mockCommentsStore.On("GetById", mock.Anything, int64(1)).Return(comment, nil)
		mockRoleStore.On("GetByName", mock.Anything, mock.Anything).Return(&requiredRole, nil)

		req, err := http.NewRequest(http.MethodDelete, "/v1/comments/1", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusForbidden, rr.Code)

		mockCommentsStore.AssertExpectations(t)
		mockUserStore.AssertExpectations(t)
		mockRoleStore.AssertExpectations(t)
	})
}
