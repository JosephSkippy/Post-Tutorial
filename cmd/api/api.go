package main

import (
	"fmt"
	"net/http"
	"time"

	"tiago-udemy/internal/auth"
	"tiago-udemy/internal/mailer"
	"tiago-udemy/internal/store"

	"tiago-udemy/docs" // this is required for swagger docs

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.MailClient // this is the mailer interface
	authenticator auth.Authenticator
}

type config struct {
	addr        string
	dbConfig    dbConfig
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
	authConfig  authConfig
}

type mailConfig struct {
	exp            time.Duration
	fromEmail      string
	mailTrapConfig mailTrapConfig
}

type mailTrapConfig struct {
	apiKey string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type authConfig struct {
	basicAuth basicAuth
	jwtAuth   jwtAuth
}

type basicAuth struct {
	username string
	password string
}

type jwtAuth struct {
	secret string
	iss    string
	exp    time.Duration
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		docsURL := fmt.Sprintf("%s/v1/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))
		r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostHandler)
			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postContextMiddleWare)
				r.Get("/", app.getPostHandler)
				r.Delete("/", app.deletePostHandler)
				r.Patch("/", app.updatePostHandler)
			})
		})
		r.Route("/comments", func(r chi.Router) {
			r.Post("/", app.createCommentHandler)
		})

		r.Route("/users", func(r chi.Router) {
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.GetUserMiddlewareContext)
				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)

			})

			r.Group(func(r chi.Router) {
				r.Get("/feed", app.userFeedHandler)
			})
		})

		//public route
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Put("/activate/{token}", app.activateUserHandler)
		})

	})

	return r
}

func (app *application) run(mux http.Handler) error {

	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	fmt.Printf("Server running on %v", app.config.addr)
	return srv.ListenAndServe()
}
