package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	mailer "social/internal/mailer"
	"social/internal/store"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestRegisterUser(t *testing.T) {
	withRedis := config{
		redisCfg: redisConfig{
			enabled: false,
		},
	}
	app := newTestApplication(t, withRedis)
	mux := app.mount()

	t.Run("should create a new user", func(t *testing.T) {
		registerUser := RegisterUserPayload{
			Username: "test",
			Email:    "test@test.com",
			Password: "test",
		}
		payload, err := json.Marshal(registerUser)
		if err != nil {
			t.Fatal(err)
		}

		mockMailer := app.mailer.(*mailer.MockMailer)
		mockMailer.On("Send",
			mailer.UserWelcome,
			"test",
			"test@test.com",
			mock.Anything,
			true,
		).Return(nil)

		req, err := http.NewRequest(http.MethodPost, "/v1/authentication/user", bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		rr := executeRequest(req, mux)

		if rr.Code != http.StatusCreated {
			t.Logf("Unexpected response: %s", rr.Body.String())
		}

		checkResponseCode(t, http.StatusCreated, rr.Code)

		mockMailer.AssertExpectations(t)
	})

	t.Run("should not create a new user", func(t *testing.T) {
		registerUserWrongEmail := RegisterUserPayload{
			Username: "test",
			Email:    "test",
			Password: "test",
		}
		payload, err := json.Marshal(registerUserWrongEmail)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "/v1/authentication/user", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		rr := executeRequest(req, mux)

		t.Logf("Response body: %s", rr.Body.String())

		checkResponseCode(t, http.StatusBadRequest, rr.Code)
	})
}

func TestCreateToken(t *testing.T) {
	withRedis := config{
		redisCfg: redisConfig{
			enabled: false,
		},
	}
	app := newTestApplication(t, withRedis)
	mux := app.mount()

	t.Run("should create a new token", func(t *testing.T) {
		var hashedPassword store.Password
		err := hashedPassword.Set("test")
		if err != nil {
			t.Fatal(err)
		}

		createTokenPayload := CreateUserTokenPayload{
			Email:    "test@test.com",
			Password: "test",
		}
		payload, err := json.Marshal(createTokenPayload)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "/v1/authentication/token", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := executeRequest(req, mux)

		var responseBody map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
		if err != nil {
			t.Fatal("Failed to parse response body:", err)
		}

		t.Logf("Parsed Response: %+v", responseBody)
		checkResponseCode(t, http.StatusCreated, rr.Code)
	})

	t.Run("should not create a new token. Invalid Password", func(t *testing.T) {
		var hashedPassword store.Password
		err := hashedPassword.Set("test1")
		if err != nil {
			t.Fatal(err)
		}

		createTokenPayload := CreateUserTokenPayload{
			Email:    "test@test.com",
			Password: "test1",
		}
		payload, err := json.Marshal(createTokenPayload)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "/v1/authentication/token", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := executeRequest(req, mux)

		var responseBody map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
		if err != nil {
			t.Fatal("Failed to parse response body:", err)
		}

		t.Logf("Parsed Response: %+v", responseBody)
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
		expectedSubstring := "unauthorized"
		if !strings.Contains(rr.Body.String(), expectedSubstring) {
			t.Errorf("Expected response body to contain: %s, but got: %s", expectedSubstring, rr.Body.String())
		}
	})
}
