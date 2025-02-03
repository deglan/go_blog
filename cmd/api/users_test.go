package main

import (
	"net/http"
	"social/internal/store"
	"social/internal/store/cache"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestGetUser(t *testing.T) {
	withRedis := config{
		redisCfg: redisConfig{
			enabled: true,
		},
	}

	app := newTestApplication(t, withRedis)
	mux := app.mount()

	testToken, err := app.authenticator.GenerateToken(nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("should not allow unauthenticated requests", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should allow authenticated requests", func(t *testing.T) {
		mockCacheStore := new(cache.MockUserStore)
		mockUserStore := new(store.MockUserStore)
		app.store.Users = mockUserStore
		app.cacheStore.Users = mockCacheStore

		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
			RoleId:   2,
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil)
		mockCacheStore.On("Get", int64(1)).Return(nil, nil).Twice()
		mockCacheStore.On("Set", mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)

		mockCacheStore.Calls = nil
	})

	t.Run("should hit the cache first and if not exists it sets the user on the cache", func(t *testing.T) {
		mockCacheStore := new(cache.MockUserStore)
		mockUserStore := new(store.MockUserStore)
		app.store.Users = mockUserStore
		app.cacheStore.Users = mockCacheStore

		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
			RoleId:   2,
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil)
		mockCacheStore.On("Get", int64(42)).Return(nil, nil)
		mockCacheStore.On("Get", int64(1)).Return(nil, nil)
		mockCacheStore.On("Set", mock.Anything, mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)

		mockCacheStore.AssertNumberOfCalls(t, "Get", 2)

		mockCacheStore.Calls = nil
	})

	t.Run("should NOT hit the cache if it is not enabled", func(t *testing.T) {
		withRedis := config{
			redisCfg: redisConfig{
				enabled: false,
			},
		}

		app := newTestApplication(t, withRedis)
		mux := app.mount()

		mockCacheStore := new(cache.MockUserStore)
		mockUserStore := new(store.MockUserStore)
		app.cacheStore.Users = mockCacheStore
		app.store.Users = mockUserStore

		user := &store.User{
			ID:       1,
			Username: "testUser",
			Email:    "test@test.com",
			RoleId:   2,
		}

		mockUserStore.On("GetById", mock.Anything, int64(1)).Return(user, nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)

		mockCacheStore.AssertNotCalled(t, "Get")

		mockCacheStore.Calls = nil
	})
}
