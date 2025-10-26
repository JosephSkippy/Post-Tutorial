package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"tiago-udemy/internal/store"

	"github.com/golang-jwt/jwt/v5"
)

type userKey string

const userCtx userKey = "user"

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			auth := r.Header.Get("Authorization")

			if auth == "" {
				app.InvalidBasicAuthorization(w, r, fmt.Errorf("no auth found"))
				return
			}

			auth_list := strings.SplitN(auth, " ", 2)

			if len(auth_list) != 2 || auth_list[0] != "Basic" {
				app.InvalidBasicAuthorization(w, r, fmt.Errorf("malformed Auth"))
				return
			}

			// decode it
			decoded, err := base64.StdEncoding.DecodeString(auth_list[1])
			if err != nil {
				app.InvalidBasicAuthorization(w, r, err)
				return
			}

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(auth_list) != 2 ||
				creds[0] != app.config.authConfig.basicAuth.username ||
				creds[1] != app.config.authConfig.basicAuth.password {
				app.InvalidBasicAuthorization(w, r, fmt.Errorf("incorrect username/pass"))
				return
			}

			next.ServeHTTP(w, r)

		})
	}

}

func (app *application) UserAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")

		if auth == "" {
			app.InvalidUserAuthorization(w, r, fmt.Errorf("no auth found"))
			return
		}

		auth_list := strings.SplitN(auth, " ", 2)

		if len(auth_list) != 2 || auth_list[0] != "Bearer" {
			app.InvalidUserAuthorization(w, r, fmt.Errorf("malformed Auth"))
			return
		}

		tokens := auth_list[1]

		jwt_tokens, err := app.authenticator.ValidateToken(tokens)
		if err != nil {
			app.InvalidUserAuthorization(w, r, err)
			return
		}

		claims, ok := jwt_tokens.Claims.(jwt.MapClaims)
		if !ok || !jwt_tokens.Valid {
			app.InvalidUserAuthorization(w, r, fmt.Errorf("invalid token claims"))
			return
		}

		// extract userId from claims
		userID, ok := claims["sub"].(float64)
		if !ok {
			app.InvalidUserAuthorization(w, r, fmt.Errorf("invalid token subject"))
			return
		}
		ctx := r.Context()

		//Extract User
		users, err := app.getUserFromCache(ctx, int64(userID))
		if err != nil {
			app.InvalidUserAuthorization(w, r, fmt.Errorf("invalid token subject"))
			return
		}

		ctx = context.WithValue(ctx, userCtx, users)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}

func (app *application) UserPostAuthorizationMiddleware(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := getUserCtx(r)
			post := getPostCtx(r)

			if user.ID == post.UserID {
				// the owner can always proceed
				next.ServeHTTP(w, r)
				return
			}
			ctx := r.Context()
			allow, err := app.store.Role.HasPermission(ctx, requiredRole, user.Role.Level)
			if err != nil {
				app.InternaServerError(w, r, err)
				return
			}
			if !allow {
				app.ForbiddenRequest(w, r, fmt.Errorf("user not allowed to perform this action"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}

}

func getUserCtx(r *http.Request) *store.User {
	user, ok := r.Context().Value(userCtx).(*store.User)

	if !ok {
		panic("expecting user store")
	}
	return user
}

func (app *application) getUserFromCache(ctx context.Context, userID int64) (*store.User, error) {

	if !app.config.cacheConfig.enabled {
		return app.store.Users.GetUserbyID(ctx, int64(userID))
	}

	// Get from cache
	// Try cache first
	user, found, err := app.cache.Users.Get(ctx, userID)
	if err != nil {
		// Log cache error but don't fail the request
		app.logger.Errorw("cache get failed",
			"user_id", userID,
			"error", err,
		)

	} else if found {
		// Cache hit - return immediately
		app.logger.Info("Getting User from Cache")
		return user, nil
	}

	app.logger.Info("Getting User from DB")
	// Cache miss or error - fetch from DB
	user, err = app.store.Users.GetUserbyID(ctx, userID)
	if err != nil {
		return nil, err // DB error is critical
	}

	// Try to populate cache (fire and forget)
	if err := app.cache.Users.Set(ctx, user); err != nil {
		// Log but don't fail - user already got their data
		app.logger.Errorw("cache set failed",
			"user_id", userID,
			"error", err,
		)
	}

	//if found user in cache
	return user, nil

}

func (app *application) RateLimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Extract IP without port
		ip := r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
		//a function that checks a requests ip's limit
		allow, waiting := app.limiter.Allow(ip)
		log.Println("from IP", ip)
		if !allow {
			app.rateLimitExceededResponse(w, r, waiting.String())
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
