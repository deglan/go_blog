package auth

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TestAuthenticator struct{}

const secret = "test_secret"
const testBasicUser = "test_user"
const testBasicPass = "test_password"

var testClaims = jwt.MapClaims{
	"sub": 1,
	"aud": "test_audience",
	"iss": "test_issuer",
	"exp": time.Now().Add(time.Hour).Unix(),
	"iat": time.Now().Unix(),
	"nbf": time.Now().Unix(),
}

func (a *TestAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims)

	tokenString, _ := token.SignedString([]byte(secret))

	return tokenString, nil
}

func (a *TestAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
}

func (a *TestAuthenticator) ValidateBasicAuth(authHeader string) error {
	if authHeader == "" {
		return errors.New("missing auth token")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Basic" {
		return errors.New("invalid auth header format")
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return errors.New("invalid base64 encoding")
	}

	creds := strings.SplitN(string(decoded), ":", 2)
	if len(creds) != 2 || creds[0] != testBasicUser || creds[1] != testBasicPass {
		return errors.New("invalid credentials")
	}

	return nil
}
