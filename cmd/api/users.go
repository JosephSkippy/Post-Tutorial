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

type targetUserKey string

const targetUserCtx targetUserKey = "targetUser"

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int	true	"User ID"
//	@Success		200		{object}	store.User
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{userID} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {

	targetUser := getTargetUserCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, targetUser); err != nil {
		app.InternaServerError(w, r, err)
		return

	}

}

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User followed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {

	targetUser := getTargetUserCtx(r)
	user := getUserCtx(r)

	ctx := r.Context()
	if err := app.store.Follower.FollowUser(ctx, user.ID, targetUser.ID); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

}

// UnfollowUser gdoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User unfollowed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow [put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {

	targetUser := getTargetUserCtx(r)
	user := getUserCtx(r)

	ctx := r.Context()
	if err := app.store.Follower.UnfollowUser(ctx, user.ID, targetUser.ID); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.InternaServerError(w, r, err)
		return
	}
}

func (app *application) GetTargetUserMiddlewareContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		targetUserStr := chi.URLParam(r, "userID")
		if targetUserStr == "" {
			app.StatusBadRequest(w, r, fmt.Errorf("missing userID"))
			return
		}

		targetUserId, err := strconv.ParseInt(targetUserStr, 10, 64)
		if err != nil {
			app.StatusBadRequest(w, r, err)
		}

		ctx := r.Context()
		user, err := app.store.Users.GetUserbyID(ctx, targetUserId)
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

		ctx = context.WithValue(ctx, targetUserCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getTargetUserCtx(r *http.Request) *store.User {
	user, ok := r.Context().Value(targetUserCtx).(*store.User)
	if !ok {
		return nil
	}
	return user
}
