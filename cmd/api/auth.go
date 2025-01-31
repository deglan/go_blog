package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	mailer "social/internal/mailer"
	"social/internal/store"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

// RegisterUser godoc
//
//	@Summary		Register user
//	@Description	register user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			user	body		RegisterUserPayload	true	"User details"
//	@Success		201		{object}	UserWithToken
//	@Failure		400		{object}	error
//	@Failure		409		{object}	error
//	@Failure		500		{object}	error
//
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	err := user.Password.Set(payload.Password)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	plainToken := uuid.New().String()

	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	ctx := r.Context()

	if err = app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.expiry); err != nil {
		switch err {
		case store.ErrorDuplicatedEmail:
			app.badRequestErrorResponse(w, r, err)
			return
		case store.ErrorDuplicatedUsername:
			app.badRequestErrorResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	userWithToken := &UserWithToken{
		User:  user,
		Token: plainToken,
	}
	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)

	isProdEnv := app.config.env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	err = app.mailer.Send(
		mailer.UserWelcome,
		user.Username,
		user.Email,
		vars,
		!isProdEnv,
	)
	if err != nil {
		app.logger.Errorw("failed to send email", "error", err.Error())
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("failed to delete user", "error", err.Error())
		}
		app.internalServerError(w, r, err)
		return
	}

	if err = app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// CreateToken godoc
//
//	@Summary		Create token
//	@Description	create token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			user	body		CreateUserTokenPayload	true	"User credentials"
//	@Success		201		string		"Token"
//	@Failure		400		{object}	error
//	@Failure		409		{object}	error
//	@Failure		500		{object}	error
//
//	@Router			/authentication/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserTokenPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.unAuthorizedErrorResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
	}

	if !user.Password.VerifyPassword(payload.Password) {
		app.unAuthorizedErrorResponse(w, r, errors.New("invalid password"))
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.iss,
		"aud": app.config.auth.token.aud,
	}
	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err = app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
	}
}
