package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"tiago-udemy/internal/mailer"
	"tiago-udemy/internal/store"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Username string `json:"username" validate:"required,min=8,max=20"`
	Password string `json:"password" validate:"required,min=8,max=50"`
}

// registerUserHandler godoc
//
//	@Summary		Register a new user
//	@Description	Register a new user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User registration payload"
//	@Success		201		{object}	string
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.StatusBadRequest(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		Role:     store.Role{Name: "user"},
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.StatusBadRequest(w, r, err)
	}

	ctx := r.Context()

	plainToken := uuid.New().String()

	//store
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	duration := app.config.mail.exp

	// create user and invitation
	if err := app.store.Users.CreateandInvite(ctx, user, hashToken, duration); err != nil {
		if err == store.ErrDuplicateEmail || err == store.ErrDuplicateUsername {
			app.StatusBadRequest(w, r, err)
			return
		}
		app.InternaServerError(w, r, err)
		return
	}
	// send email invitation
	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)
	data := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	status, err := app.mailer.Send(ctx, mailer.UserWelcomeTemplate, user.Email, data)
	if err != nil {
		app.logger.Errorw("error sending welcome email", "error", err)

		// rollback user creation if email fails (SAGA pattern)
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("error deleting user", "error", err)
			return
		}

		app.InternaServerError(w, r, err)
		return
	}

	app.logger.Infow("Email sent", "status code", status)

	if err := app.jsonResponse(w, http.StatusCreated, nil); err != nil {
		app.InternaServerError(w, r, err)
		return
	}
}

// ActivateUser godoc
//
//	@Summary		Activates/Register a user
//	@Description	Activates/Register a user by invitation token
//	@Tags			authentication
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		200		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	plainToken := chi.URLParam(r, "token")

	if plainToken == "" {
		app.StatusBadRequest(w, r, fmt.Errorf("missing tokens"))
		return
	}

	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	ctx := r.Context()
	if err := app.store.Users.Activate(ctx, hashToken); err != nil {
		if errors.Is(err, store.ErrInvalidToken) {
			app.StatusBadRequest(w, r, err)
			return
		}
		app.InternaServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, "User Successfully Activated"); err != nil {
		app.InternaServerError(w, r, err)
		return
	}
}

type UserLoginPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=50"`
}

// authUserHandler godoc
//
//	@Summary		Authenticate the User
//	@Description	Authenticate the user and return credentials token
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		UserLoginPayload	true	"User Login payload"
//	@Success		200		{string}	string				"Token"
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/login [post]
func (app *application) authUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload UserLoginPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.InvalidUserAuthorization(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.InvalidUserAuthorization(w, r, err)
		return
	}
	ctx := r.Context()
	user, err := app.store.Users.GetUserByEmail(ctx, payload.Email)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):

			// change error back to invalid credentials
			app.InvalidUserAuthorization(w, r, fmt.Errorf("no user found"))
			return
		default:
			app.InternaServerError(w, r, err)
			return
		}
	}

	err = user.Password.Compare(payload.Password)
	if err != nil {
		app.InvalidUserAuthorization(w, r, fmt.Errorf("invalid credentials"))
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.authConfig.jwtAuth.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.authConfig.jwtAuth.iss,
		"aud": app.config.authConfig.jwtAuth.iss,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.InternaServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.InternaServerError(w, r, err)
	}
}
