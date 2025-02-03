package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"social/internal/store"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestFeed(t *testing.T) {
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

	t.Run("should return feed", func(t *testing.T) {
		mockUserStore := new(store.MockUserStore)
		app.store.Users = mockUserStore
		mockPostsStore := new(store.MockPostStore)
		app.store.Posts = mockPostsStore
		posts := []store.Post{
			{ID: 1, Title: "Test Post 1", Content: "This is a test post", UserId: 1},
			{ID: 2, Title: "Test Post 2", Content: "This is another test post", UserId: 1}}
		expectedPosts := []store.PostWithMetadata{
			{Post: posts[0], CommentsCount: 1},
			{Post: posts[1], CommentsCount: 1},
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(&store.User{ID: 1}, nil).Once()
		mockPostsStore.On("GetUserFeed", mock.Anything, int64(1), mock.Anything).
			Return(expectedPosts, nil).
			Once()

		req, err := http.NewRequest(http.MethodGet, "/v1/users/feed?limit=5&sort=asc", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		var response struct {
			Data []store.PostWithMetadata `json:"data"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if len(response.Data) != 2 {
			t.Errorf("Expected 2 posts, got %d", len(response.Data))
		}

		mockPostsStore.AssertExpectations(t)

	})

	t.Run("should return bad request wrong sort", func(t *testing.T) {
		mockUserStore := new(store.MockUserStore)
		app.store.Users = mockUserStore

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(&store.User{ID: 1}, nil).Once()

		req, err := http.NewRequest(http.MethodGet, "/v1/users/feed?sort=wrong", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rr.Code)
		}
	})

	t.Run("should return internal error from getUserFeed", func(t *testing.T) {
		mockPostsStore := new(store.MockPostStore)
		mockUserStore := new(store.MockUserStore)
		app.store.Users = mockUserStore
		app.store.Posts = mockPostsStore

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(&store.User{ID: 1}, nil).Once()

		mockPostsStore.On("GetUserFeed", mock.Anything, int64(1), mock.Anything).
			Return(nil, errors.New("error test")).
			Once()

		req, err := http.NewRequest(http.MethodGet, "/v1/users/feed?limit=5&sort=asc", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", rr.Code)
		}
	})

}
