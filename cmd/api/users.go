package main

import (
	"errors"
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtxKey userKey = "user"

// GETUSER godoc
//
//	@Summary		Get user by ID
//	@Description	get user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userId	path		int	true	"User ID"
//	@Success		200		{object}	store.User
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//
//	@Security		ApiKeyAuth
//	@Router			/users/{userId} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	user, err := app.getUserWithRedis(r.Context(), userID)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.notFoundErrorResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

// FollowUser godoc
//
//	@Summary		Follow user
//	@Description	follow user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userId	path	int	true	"User ID"
//	@Success		204
//	@Failure		400	{object}	error
//	@Failure		409	{object}	error
//	@Failure		500	{object}	error
//
//	@Security		ApiKeyAuth
//	@Router			/users/{userId}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	followerUser := getUserFromContext(r)
	followedId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	err = app.store.Followers.FollowUser(ctx, followerUser.ID, followedId)
	if err != nil {
		switch err {
		case store.ErrorAlreadyExists:
			app.conflictErrorResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// UnfollowUser godoc
//
//	@Summary		Unfollow user
//	@Description	unfollow user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userId	path	int	true	"User ID"
//	@Success		204
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//
//	@Security		ApiKeyAuth
//	@Router			/users/{userId}/unfollow [put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	followerUser := getUserFromContext(r)
	followedId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	err = app.store.Followers.UnfollowUser(ctx, followerUser.ID, followedId)
	if err != nil {
		switch err {
		case store.ErrorAlreadyExists:
			app.conflictErrorResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err = app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// ActivateUser godoc
//
//	@Summary		Activate user
//	@Description	activate user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	if token == "" {
		app.badRequestErrorResponse(w, r, errors.New("empty token"))
		return
	}

	if err := app.store.Users.Activate(r.Context(), token); err != nil {
		switch err {
		case store.ErrorNotFound:
			app.notFoundErrorResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusNoContent, "User activated"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func getUserFromContext(r *http.Request) *store.User {
	return r.Context().Value(userCtxKey).(*store.User)
}
