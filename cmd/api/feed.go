package main

import (
	"errors"
	"net/http"
	"tiago-udemy/internal/store"
)

func (app *application) userFeedHandler(w http.ResponseWriter, r *http.Request) {

	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
		Tags:   []string{},
		Search: "",
	}
	fq, err := fq.Parse(w, r)
	if err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	ctx := r.Context()

	feeds, err := app.store.Posts.GetFeed(ctx, int64(1), fq)
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

	if err := app.jsonResponse(w, http.StatusOK, feeds); err != nil {
		app.InternaServerError(w, r, err)
		return
	}

}
