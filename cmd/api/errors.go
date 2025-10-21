package main

import (
	"net/http"
)

func (app *application) InternaServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusInternalServerError, "the server encountered  a problem")
}

func (app *application) RecordNotFound(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("Record Not Found", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusNotFound, err.Error())
}

func (app *application) StatusBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("Bad Request", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) InvalidBasicAuthorization(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("invalid authorization", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	writeJSONError(w, http.StatusUnauthorized, "invalid authorization")
}

func (app *application) InvalidUserAuthorization(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("invalid authorization", "method", r.Method, "path", r.URL.Path, "error", err.Error())

	writeJSONError(w, http.StatusUnauthorized, "invalid authorization")
}
