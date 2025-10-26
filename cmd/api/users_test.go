package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tiago-udemy/internal/store"

	"github.com/go-chi/chi/v5"
)

func TestGetUserHandler_OK(t *testing.T) {
	app := newTestApp()

	want := &store.User{ID: 42, Email: "demo@example.com"}

	req := httptest.NewRequest(http.MethodGet, "/v1/users/42", nil)
	req = req.WithContext(context.WithValue(req.Context(), targetUserCtx, want))

	rr := httptest.NewRecorder()
	app.getUserHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d; body=%s", rr.Code, rr.Body.String())
	}

	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("want JSON Content-Type, got %q", ct)
	}

	var response struct {
		Data *store.User `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("invalid JSON: %v; body=%s", err, rr.Body.String())
	}

	if response.Data == nil {
		t.Fatal("got nil user in response data")
	}

	if response.Data.ID != want.ID || response.Data.Email != want.Email {
		t.Fatalf("wrong user: got ID=%d Email=%q, want ID=%d Email=%q",
			response.Data.ID, response.Data.Email, want.ID, want.Email)
	}
}

func TestGetTargetUserMiddlewareContext(t *testing.T) {
	app := newTestApp()

	t.Run("userID missing → should return 400, NOT call next handler", func(t *testing.T) {
		// Step 1: Create a "next handler" that should NOT be called
		nextCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			t.Error("next handler should not be called when userID is missing")
		})

		// fake a request that does have path
		req := httptest.NewRequest(http.MethodGet, "/v1/users/", nil)

		// call the middleware
		handler := app.GetTargetUserMiddlewareContext(nextHandler)

		rctx := chi.NewRouteContext()
		// Don't add userID param - we want it missing!
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Step 5: Create response recorder
		rr := httptest.NewRecorder()

		// Step 6: Call the handler
		handler.ServeHTTP(rr, req)

		// Step 7: Assert the results
		if rr.Code != http.StatusBadRequest {
			t.Errorf("want 400, got %d; body=%s", rr.Code, rr.Body.String())
		}

		if nextCalled {
			t.Error("next handler was called but shouldn't have been")
		}

	})

	t.Run("invalid userId → should return 400, NOT call next handler", func(t *testing.T) {
		// Step 1: Create a "next handler" that should NOT be called
		nextCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			t.Error("next handler should not be called when userID is missing")
		})

		// fake a request that does have path
		req := httptest.NewRequest(http.MethodGet, "/v1/users/abcd123", nil)

		// call the middleware
		handler := app.GetTargetUserMiddlewareContext(nextHandler)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("userID", "abcd123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Step 5: Create response recorder
		rr := httptest.NewRecorder()

		// Step 6: Call the handler
		handler.ServeHTTP(rr, req)

		// Step 7: Assert the results
		if rr.Code != http.StatusBadRequest {
			t.Errorf("want 400, got %d; body=%s", rr.Code, rr.Body.String())
		}

		if nextCalled {
			t.Error("next handler was called but shouldn't have been")
		}

	})

	t.Run("userId → should pass", func(t *testing.T) {

		// Spy to capture what's in context
		var capturedUser *store.User
		nextCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			capturedUser = getTargetUserCtx(r) // Retrieve from context
			w.WriteHeader(http.StatusOK)
		})

		// fake a request that does have path
		req := httptest.NewRequest(http.MethodGet, "/v1/users/123", nil)

		// call the middleware
		handler := app.GetTargetUserMiddlewareContext(nextHandler)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("userID", "123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Step 5: Create response recorder
		rr := httptest.NewRecorder()

		// Step 6: Call the handler
		handler.ServeHTTP(rr, req)

		// Assertions
		if rr.Code != http.StatusOK {
			t.Errorf("want 200, got %d; body=%s", rr.Code, rr.Body.String())
		}

		if !nextCalled {
			t.Error("next handler should have been called")
		}

		if capturedUser == nil {
			t.Fatal("user was not added to context")
		}

		if capturedUser.ID != 123 {
			t.Errorf("wrong user ID: got %d, want 123", capturedUser.ID)
		}

	})

}

// Table driven test

// func TestGetTargetUserMiddlewareContextTableStyle(t *testing.T) {
// 	app := newTestApp()

// 	type tc struct {
// 		name        string
// 		userIDParam string
// 		setupMock   func(m store.Storage) // <-- use your mock's API here
// 		wantStatus  int
// 		wantNext    bool
// 		wantUserID  int64 // 0 = don't assert
// 	}

// 	tests := []tc{
// 		{
// 			name:        "missing userID -> 400, next not called",
// 			userIDParam: "",
// 			setupMock: func(m store.Storage) {
// 				// No calls expected
// 			},
// 			wantStatus: http.StatusBadRequest,
// 			wantNext:   false,
// 		},
// 		{
// 			name:        "invalid userID -> 400, next not called",
// 			userIDParam: "abc",
// 			setupMock: func(m store.Storage) {
// 				// No calls expected
// 			},
// 			wantStatus: http.StatusBadRequest,
// 			wantNext:   false,
// 		},
// 		{
// 			name:        "not found -> 404, next not called",
// 			userIDParam: "7",
// 			setupMock: func(m store.Storage) {
// 				// STYLE A: if your mock exposes Users().OnGetUserbyID(fn)
// 				// m.Users().OnGetUserbyID(func(_ context.Context, _ int64) (*store.User, error) {
// 				// 	 return nil, store.ErrRecordNotFound
// 				// })

// 				// STYLE B: if your mock exposes a struct you can replace:
// 				// m.Users = &store.UsersMock{
// 				// 	 GetUserbyIDFunc: func(ctx context.Context, id int64) (*store.User, error) {
// 				//     return nil, store.ErrRecordNotFound
// 				//   },
// 				// }
// 			},
// 			wantStatus: http.StatusNotFound, // use 400 if you haven't switched yet
// 			wantNext:   false,
// 		},
// 		{
// 			name:        "store error -> 500, next not called",
// 			userIDParam: "9",
// 			setupMock: func(m store.Storage) {
// 				// A or B as above; return generic error
// 				// m.Users().OnGetUserbyID(func(ctx context.Context, id int64) (*store.User, error) {
// 				// 	 return nil, errors.New("db exploded")
// 				// })
// 			},
// 			wantStatus: http.StatusInternalServerError,
// 			wantNext:   false,
// 		},
// 		{
// 			name:        "success -> injects user and calls next",
// 			userIDParam: "42",
// 			setupMock: func(m store.Storage) {
// 				// A or B as above; return a user
// 				// m.Users().OnGetUserbyID(func(ctx context.Context, id int64) (*store.User, error) {
// 				//   return &store.User{ID: 42, Email: "demo@example.com"}, nil
// 				// })
// 			},
// 			wantStatus: http.StatusOK,
// 			wantNext:   true,
// 			wantUserID: 42,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// fresh mock storage for each test to avoid cross-test leakage
// 			app.store = store.MockNewStorage()
// 			tt.setupMock(app.store)

// 			gotStatus, gotNext, gotUser := runMW(t, app, tt.userIDParam)

// 			if gotStatus != tt.wantStatus {
// 				t.Fatalf("status: got %d, want %d", gotStatus, tt.wantStatus)
// 			}
// 			if gotNext != tt.wantNext {
// 				t.Fatalf("next called: got %v, want %v", gotNext, tt.wantNext)
// 			}
// 			if tt.wantUserID != 0 {
// 				if gotUser == nil {
// 					t.Fatalf("want user in ctx, got nil")
// 				}
// 				if gotUser.ID != tt.wantUserID {
// 					t.Fatalf("user ID: got %d, want %d", gotUser.ID, tt.wantUserID)
// 				}
// 			}
// 		})
// 	}
// }
