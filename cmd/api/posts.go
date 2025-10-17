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

type CreatePayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type postKey string

const postCtx postKey = "post"

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePayload

	if err := readJSON(w, r, &payload); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,

		// change after Auth
		UserID: 1,
	}

	ctx := r.Context()
	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

	//return the results
	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.InternaServerError(w, r, err)
	}

}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {

	post := getPostCtx(r)

	ctx := r.Context()
	comments, err := app.store.Comment.GetCommentByID(ctx, post.ID)

	if err != nil {
		app.InternaServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.InternaServerError(w, r, err)
		return
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "postID")
	if idStr == "" {
		app.StatusBadRequest(w, r, fmt.Errorf("missing postID"))
		return
	}

	// 2) Convert to int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		app.StatusBadRequest(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Comment.DeleteCommentByPostID(ctx, id); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

	post, err := app.store.Posts.DeletePost(ctx, id)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			app.RecordNotFound(w, r, err)
			return
		default:
			app.InternaServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.InternaServerError(w, r, err)
		return
	}
}

type UpdatePayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {

	post := getPostCtx(r)

	var payload UpdatePayload
	if err := readJSON(w, r, &payload); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	if payload.Title == nil && payload.Content == nil {
		app.StatusBadRequest(w, r, fmt.Errorf("no fields to update"))
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if err := app.store.Posts.UpdatePost(r.Context(), post); err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			app.RecordNotFound(w, r, fmt.Errorf("version mismatch"))
			return
		default:
			app.InternaServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

}

func (app *application) postContextMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		idStr := chi.URLParam(r, "postID")
		if idStr == "" {
			app.StatusBadRequest(w, r, fmt.Errorf("missing postID"))
			return
		}

		// 2) Convert to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			app.StatusBadRequest(w, r, err)
			return
		}

		ctx := r.Context()
		post, err := app.store.Posts.Get(ctx, id)

		if err != nil {
			switch {
			case errors.Is(err, store.ErrRecordNotFound):
				app.RecordNotFound(w, r, err)
				return
			default:
				app.InternaServerError(w, r, err)
				return
			}
		}

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))

	})

}

func getPostCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
