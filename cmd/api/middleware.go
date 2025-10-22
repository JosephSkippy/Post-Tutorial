package main

import (
	"context"
	"encoding/base64"
	"fmt"
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

			if len(auth_list) != 2 || auth_list[0] != "Bearer" {
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
		users, err := app.store.Users.GetUserbyID(ctx, int64(userID))
		if err != nil {
			app.InvalidUserAuthorization(w, r, fmt.Errorf("invalid token subject"))
			return
		}

		ctx = context.WithValue(ctx, userCtx, users)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}

func getUserCtx(r *http.Request) *store.User {
	user, ok := r.Context().Value(userCtx).(*store.User)

	if !ok {
		panic("expecting user store")
	}
	return user
}
