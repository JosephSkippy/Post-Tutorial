package main

import (
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

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePayload

	if err := readJSON(w, r, &payload); err != nil {
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

	if err := Validate.Struct(payload); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

	//return the results
	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		app.InternaServerError(w, r, err)
	}

}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {

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

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		app.InternaServerError(w, r, err)
		return
	}
}
