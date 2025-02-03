package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"social/internal/store"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestCreateComment(t *testing.T) {
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

	t.Run("should create a new comment", func(t *testing.T) {
		mockUserStore := new(store.MockUserStore)
		app.store.Users = mockUserStore
		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil).Once()

		createComment := CreateCommentPayload{
			PostId:  1,
			Content: "test",
		}
		payload, err := json.Marshal(createComment)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "/v1/posts/1/comments", bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		t.Logf("response: %s", rr.Body.String())

		checkResponseCode(t, http.StatusCreated, rr.Code)
	})

	t.Run("should not allow unauthenticated requests", func(t *testing.T) {
		createComment := CreateCommentPayload{
			PostId:  1,
			Content: "test",
		}
		payload, err := json.Marshal(createComment)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "/v1/posts/1/comments", bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}

		rr := executeRequest(req, mux)

		t.Logf("response: %s", rr.Body.String())

		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestGetComments(t *testing.T) {
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

	t.Run("should allow authenticated requests", func(t *testing.T) {
		mockUserStore := new(store.MockUserStore)
		app.store.Users = mockUserStore
		mockCommentStore := new(store.MockCommentsStore)
		app.store.Comments = mockCommentStore
		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil).Once()

		mockCommentStore.On("GetByPostId", mock.Anything, int64(1)).Return(
			[]store.Comment{{Id: 1, PostId: 1, Content: "test"}},
			nil).Once()

		req, err := http.NewRequest(http.MethodGet, "/v1/posts/1/comments", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		t.Logf("response: %s", rr.Body.String())

		checkResponseCode(t, http.StatusOK, rr.Code)
		mockCommentStore.AssertExpectations(t)
	})
}

func TestUpdateComment(t *testing.T) {
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

	t.Run("should update comment", func(t *testing.T) {
		mockUserStore := new(store.MockUserStore)
		app.store.Users = mockUserStore
		mockCommentStore := new(store.MockCommentsStore)
		app.store.Comments = mockCommentStore
		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
		}

		comment := &store.Comment{Id: 1, PostId: 1, UserId: 1, Content: "test"}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil).Once()
		mockCommentStore.On("GetById", mock.Anything, int64(1)).Return(comment, nil).Once()

		updateComment := UpdateCommentPayload{
			Content: &[]string{"test1"}[0],
		}
		payload, err := json.Marshal(updateComment)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPatch, "/v1/comments/1", bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		t.Logf("Response Code: %d", rr.Code)
		t.Logf("Response Body: %s", rr.Body.String())

		checkResponseCode(t, http.StatusOK, rr.Code)

		mockCommentStore.AssertExpectations(t)
	})
}

func TestDeleteComment(t *testing.T) {
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

	t.Run("should delete comment", func(t *testing.T) {
		mockUserStore := new(store.MockUserStore)
		app.store.Users = mockUserStore
		mockCommentStore := new(store.MockCommentsStore)
		app.store.Comments = mockCommentStore
		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
		}

		comment := &store.Comment{Id: 1, PostId: 1, UserId: 1, Content: "test"}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil).Once()
		mockCommentStore.On("GetById", mock.Anything, int64(1)).Return(comment, nil).Once()
		mockCommentStore.On("Delete", mock.Anything, int64(1), user.ID).Return(nil).Once()

		req, err := http.NewRequest(http.MethodDelete, "/v1/comments/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req = req.WithContext(context.WithValue(req.Context(), userCtxKey, user))

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusNoContent, rr.Code)

		mockCommentStore.AssertExpectations(t)
	})
}
