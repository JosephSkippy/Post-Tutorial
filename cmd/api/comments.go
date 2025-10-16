package main

import (
	"net/http"
	"tiago-udemy/internal/store"
)

type createCommentPayload struct {
	PostID  int64  `json:"post_id" validate:"required"`
	Comment string `json:"comment" validate:"required,max=100"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var payload createCommentPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	comment := &store.Comment{
		PostID:  payload.PostID,
		UserID:  1,
		Comment: payload.Comment,
	}

	ctx := r.Context()

	if err := app.store.Comment.Create(ctx, comment); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusAccepted, comment); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

}
