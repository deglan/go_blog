package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"social/internal/store"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

func (app *application) BasicAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.logger.Warn("missing auth token")
				app.unAuthorizedBasicErrorResponse(w, r, errors.New("missing auth token"))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.logger.Warn("invalid auth header")
				app.unAuthorizedBasicErrorResponse(w, r, errors.New("invalid auth header. Probably malformed"))
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.logger.Warn("invalid auth header")
				app.unAuthorizedBasicErrorResponse(w, r, err)
				return
			}

			username := app.config.auth.basic.user
			password := app.config.auth.basic.password
			if username == "" || password == "" {
				app.logger.Error("basic auth credentials are not configured")
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != password {
				app.logger.Warn("invalid auth header")
				app.unAuthorizedBasicErrorResponse(w, r, errors.New("wrong username or password"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) AuthTokenMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.logger.Warn("missing auth token")
				app.unAuthorizedErrorResponse(w, r, errors.New("missing auth token"))
				return
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				app.logger.Warn("invalid auth header")
				app.unAuthorizedErrorResponse(w, r, errors.New("invalid auth header. Probably malformed"))
				return
			}

			token := parts[1]

			jwtToken, err := app.authenticator.ValidateToken(token)
			if err != nil {
				app.logger.Warn("invalid token")
				app.unAuthorizedErrorResponse(w, r, err)
				return
			}

			claims, _ := jwtToken.Claims.(jwt.MapClaims)

			userId, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
			if err != nil {
				app.unAuthorizedErrorResponse(w, r, err)
				return
			}

			ctx := r.Context()

			user, err := app.getUserWithRedis(ctx, userId)
			if err != nil {
				app.unAuthorizedErrorResponse(w, r, err)
				return
			}

			ctx = context.WithValue(ctx, userCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		post := getPostFromContext(r)

		if post.UserId == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenErrorResponse(w, r, errors.New("user has no privileges to perform this action"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)

	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *application) getUserWithRedis(ctx context.Context, userId int64) (*store.User, error) {
	if !app.config.redisCfg.enabled {
		return app.store.Users.GetById(ctx, userId)
	}

	user, err := app.cacheStore.Users.Get(ctx, userId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = app.store.Users.GetById(ctx, userId)
		if err != nil {
			return nil, err
		}

		err = app.cacheStore.Users.Set(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (app *application) commentContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commentIdStr := chi.URLParam(r, "commentId")

		commentId, err := strconv.ParseInt(commentIdStr, 10, 64)
		if err != nil {
			app.logger.Error("Invalid comment ID")
			app.badRequestErrorResponse(w, r, errors.New("invalid comment ID"))
			return
		}

		ctx := r.Context()
		comment, err := app.store.Comments.GetByPostId(ctx, commentId)
		if err != nil {
			if errors.Is(err, store.ErrorNotFound) {
				app.logger.Warnf("Comment not found for ID: %d", commentId)
				app.notFoundErrorResponse(w, r, err)
			} else {
				app.logger.Errorf("Error fetching comment: %v", err)
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, commentCtxKey, &comment)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
