package main

import (
	"tiago-udemy/internal/store"

	"go.uber.org/zap"
)

func newTestApp() *application {
	return &application{
		logger: zap.NewNop().Sugar(), // quiet logger for tests
		store:  store.MockNewStorage(),
		// other fields nil by default
	}
}

// func withChiParam(r *http.Request, key, val string) *http.Request {
// 	rctx := chi.NewRouteContext()
// 	if val != "" {
// 		rctx.URLParams.Add(key, val)
// 	}
// 	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
// }

// runMW runs the middleware with a given userID param and returns:
// - status code
// - whether next was called
// - the user captured from ctx if next ran
// func runMW(t *testing.T, app *application, userIDParam string) (status int, nextCalled bool, captured *store.User) {
// 	t.Helper()

// 	// spy "next" handler
// 	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		nextCalled = true
// 		captured = getTargetUserCtx(r)
// 		w.WriteHeader(http.StatusOK) // sentinel
// 	})

// 	h := app.GetTargetUserMiddlewareContext(next)

// 	path := "/v1/users/"
// 	if userIDParam != "" {
// 		path += userIDParam
// 	}
// 	req := httptest.NewRequest(http.MethodGet, path, nil)
// 	req = withChiParam(req, "userID", userIDParam)

// 	rr := httptest.NewRecorder()
// 	h.ServeHTTP(rr, req)
// 	return rr.Code, nextCalled, captured
// }
