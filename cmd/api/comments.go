package main

import (
	"errors"
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type commentKey string

const commentCtxKey commentKey = "comment"

type CreateCommentPayload struct {
	PostId  int64  `json:"post_id" validate:"required"`
	Content string `json:"content" validate:"required,max=1000"`
}

type UpdateCommentPayload struct {
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

// CreateComment godoc
//
//	@Summary		Create comment
//	@Description	create a new comment on a post
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateCommentPayload	true	"Create comment payload"
//	@Success		201		{object}	store.Comment
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//
//	@Security		ApiKeyAuth
//	@Router			/posts/{postId}/comments [post]
func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateCommentPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)
	comment := &store.Comment{
		PostId:  payload.PostId,
		UserId:  user.ID,
		Content: payload.Content,
		User:    *user,
	}

	ctx := r.Context()
	if err := app.store.Comments.Create(ctx, comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// GetComments godoc
//
//	@Summary		Get comments
//	@Description	get comments for a post
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			postId	path		int	true	"Post ID"
//	@Success		200		{array}		store.Comment
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//
//	@Router			/posts/{postId}/comments [get]
func (app *application) getCommentsHandler(w http.ResponseWriter, r *http.Request) {
	postId, err := strconv.ParseInt(chi.URLParam(r, "postId"), 10, 64)
	if err != nil {
		app.badRequestErrorResponse(w, r, errors.New("invalid post ID"))
		return
	}

	comments, err := app.store.Comments.GetByPostId(r.Context(), postId)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, comments); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// UpdateComment godoc
//
//	@Summary		Update comment
//	@Description	update an existing comment
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			commentId	path		int						true	"Comment ID"
//	@Param			payload		body		UpdateCommentPayload	true	"Update comment payload"
//	@Success		200			{object}	store.Comment
//	@Failure		400			{object}	error
//	@Failure		404			{object}	error
//	@Failure		500			{object}	error
//
//	@Security		ApiKeyAuth
//	@Router			/comments/{commentId} [patch]
func (app *application) updateCommentHandler(w http.ResponseWriter, r *http.Request) {
	comment := getCommentFromContext(r)
	if comment == nil {
		app.logger.Error("No comment found in context")
		app.notFoundErrorResponse(w, r, errors.New("no comment found in context"))
		return
	}

	var payload struct {
		Content *string `json:"content" validate:"required,max=1000"`
	}
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	if payload.Content != nil {
		comment.Content = *payload.Content
	}

	if err := app.store.Comments.Update(r.Context(), comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// DeleteComment godoc
//
//	@Summary		Delete comment
//	@Description	delete a comment
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			commentId	path	int	true	"Comment ID"
//	@Success		204
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//
//	@Security		ApiKeyAuth
//	@Router			/comments/{commentId} [delete]
func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentId, err := strconv.ParseInt(chi.URLParam(r, "commentId"), 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	user := getUserFromContext(r)
	ctx := r.Context()

	if err := app.store.Comments.Delete(ctx, commentId, user.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getCommentFromContext(r *http.Request) *store.Comment {
	comment, _ := r.Context().Value(commentCtxKey).(*store.Comment)
	return comment
}
