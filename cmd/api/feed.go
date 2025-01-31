package main

import (
	"net/http"
	"social/internal/store"
)

// getUserFeedHandler retrieves a paginated list of user feed posts.
//
//	@Summary		Get user feed
//	@Description	Retrieves a paginated list of posts in the user's feed.
//	@Tags			Feed
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int						false	"Limit of posts per page"				default(10)
//	@Param			offset	query		int						false	"Offset for pagination"					default(0)
//	@Param			sort	query		string					false	"Sort order, either 'asc' or 'desc'"	default(desc)
//	@Param			since	query		string					false	"Filter posts created after this date (format: YYYY-MM-DDTHH:MM:SSZ)"
//	@Param			until	query		string					false	"Filter posts created before this date (format: YYYY-MM-DDTHH:MM:SSZ)"
//	@Param			search	query		string					false	"Search term to filter posts by title or content"
//	@Param			tags	query		string					false	"Comma-separated list of tags to filter posts"
//	@Success		200		{array}		store.PostWithMetadata	"List of posts with metadata"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedFeedQuery{
		Limit:  10,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestErrorResponse(w, r, err)
		return
	}

	ctx := r.Context()
	feed, err := app.store.Posts.GetUserFeed(ctx, int64(1), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err = app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
