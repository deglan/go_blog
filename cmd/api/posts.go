package main

import (
	"context"
	"errors"
	"net/http"
	"social/internal/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtxKey postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

// CreatePost godoc
//
//	@Summary		Create post
//	@Description	create post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostPayload	true	"Create post payload"
//	@Success		201		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)
	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserId:  user.ID,
	}

	ctx := r.Context()
	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// GetPost godoc
//
//	@Summary		Get post
//	@Description	get post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postId	path		int	true	"Post ID"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//
//	@Security		ApiKeyAuth
//	@Router			/posts/{postId} [get]
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromContext(r)
	if post == nil {
		app.notFoundErrorResponse(w, r, errors.New("post not found in context"))
		return
	}

	comments, err := app.store.Comments.GetByPostId(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// DeletePost godoc
//
//	@Summary		Delete post
//	@Description	delete post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postId	path	int	true	"Post ID"
//	@Success		204
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//
//	@Security		ApiKeyAuth
//	@Router			/posts/{postId} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	postId, err := strconv.ParseInt(chi.URLParam(r, "postId"), 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	ctx := r.Context()

	if err := app.store.Posts.Delete(ctx, postId); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type UpdatePostPayload struct {
	Title   *string  `json:"title" validate:"omitempty,max=100"`
	Content *string  `json:"content" validate:"omitempty,max=1000"`
	Tags    []string `json:"tags" validate:"omitempty"`
}

// UpdatePost godoc
//
//	@Summary		Update post
//	@Description	update post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			postId	path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostPayload	true	"Update post payload"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//
//	@Security		ApiKeyAuth
//	@Router			/posts/{postId} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromContext(r)

	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Tags != nil {
		post.Tags = payload.Tags
	}

	if err := app.store.Posts.Update(r.Context(), post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postIdStr := chi.URLParam(r, "postId")

		postId, err := strconv.ParseInt(postIdStr, 10, 64)
		if err != nil {
			app.logger.Error("Invalid post ID")
			app.badRequestErrorResponse(w, r, errors.New("invalid post ID"))
			return
		}

		ctx := r.Context()
		post, err := app.store.Posts.GetById(ctx, postId)
		if err != nil {
			if errors.Is(err, store.ErrorNotFound) {
				app.logger.Warnf("Post not found for ID: %d", postId)
				app.notFoundErrorResponse(w, r, err)
			} else {
				app.logger.Errorf("Error fetching post: %v", err)
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, postCtxKey, &post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromContext(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtxKey).(*store.Post)
	return post
}
