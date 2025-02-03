package main

import (
	"net/http"
	"net/http/httptest"
	"social/internal/auth"
	mailer "social/internal/mailer"
	"social/internal/ratelimiter"
	"social/internal/store"
	"social/internal/store/cache"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestApplication(t *testing.T, cfg config) *application {
	t.Helper()

	logger := zap.NewNop().Sugar()
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockStore()
	mockMailer := new(mailer.MockMailer)
	testAuth := &auth.TestAuthenticator{}

	if cfg.rateLimiter.RequestsPerTimeFrame == 0 {
		cfg.rateLimiter.RequestsPerTimeFrame = 20
	}
	if cfg.rateLimiter.TimeFrame == 0 {
		cfg.rateLimiter.TimeFrame = 5 * time.Second
	}

	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestsPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	return &application{
		config:        cfg,
		logger:        logger,
		store:         mockStore,
		cacheStore:    mockCacheStore,
		mailer:        mockMailer,
		authenticator: testAuth,
		rateLimiter:   rateLimiter,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}
