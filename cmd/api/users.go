package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"tiago-udemy/internal/store"

	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {

	user := getUserCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.InternaServerError(w, r, err)
		return

	}

}

type FollowerPayload struct {
	FollowID int64
	UserID   int64 `json:"user_id"`
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserCtx(r)

	// TODO: revert back to get User from Auth instead of ctx
	var payload FollowerPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	// This variable created to avoid confusion
	var follower store.Followers
	follower.FollowerID = user.ID
	follower.UserID = payload.UserID

	ctx := r.Context()
	if err := app.store.Follower.FollowUser(ctx, follower.UserID, follower.FollowerID); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {

	user := getUserCtx(r)

	// TODO: revert back to get User from Auth instead of ctx
	var payload FollowerPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	// This variable created to avoid confusion
	var follower store.Followers
	follower.FollowerID = user.ID
	follower.UserID = payload.UserID

	ctx := r.Context()
	if err := app.store.Follower.UnfollowUser(ctx, follower.UserID, follower.FollowerID); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.InternaServerError(w, r, err)
		return
	}
}

func (app *application) GetUserMiddlewareContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		idStr := chi.URLParam(r, "userID")
		if idStr == "" {
			app.StatusBadRequest(w, r, fmt.Errorf("missing userID"))
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			app.StatusBadRequest(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetUserbyID(ctx, id)

		if err != nil {
			switch {
			case errors.Is(err, store.ErrRecordNotFound):
				app.StatusBadRequest(w, r, err)
				return
			default:
				app.InternaServerError(w, r, err)
				return
			}
		}

		ctx = context.WithValue(ctx, userCtx, user)
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
