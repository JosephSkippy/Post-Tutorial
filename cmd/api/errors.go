package main

import (
	"log"
	"net/http"
)

func (app *application) InternaServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Internal Server Error %s path: %s, Error %v", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) RecordNotFound(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Record Not Found Error %s path: %s, Error %v", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusNotFound, err.Error())
}

func (app *application) StatusBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Bad Request Error %s path: %s, Error %v", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusBadRequest, err.Error())
}
