package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tiago-udemy/internal/auth"
	"tiago-udemy/internal/mailer"
	"tiago-udemy/internal/ratelimiter"
	"tiago-udemy/internal/store"
	"tiago-udemy/internal/store/cache"

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
	cache         cache.CacheStorage
	limiter       ratelimiter.Limiter
}

type config struct {
	addr          string
	dbConfig      dbConfig
	env           string
	apiURL        string
	mail          mailConfig
	frontendURL   string
	authConfig    authConfig
	cacheConfig   cacheConfig
	limiterConfig limiterConfig
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

type cacheConfig struct {
	redis   redisConfig
	enabled bool
}

type redisConfig struct {
	addr string
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

type limiterConfig struct {
	window     time.Duration
	maxRequest int
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(app.RateLimitingMiddleware)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		docsURL := fmt.Sprintf("%s/v1/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))
		r.Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.UserAuthMiddleware)
			r.Post("/", app.createPostHandler)
			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postContextMiddleWare)
				r.With(app.UserPostAuthorizationMiddleware("moderator")).Get("/", app.getPostHandler)
				r.With(app.UserPostAuthorizationMiddleware("admin")).Delete("/", app.deletePostHandler)
				r.With(app.UserPostAuthorizationMiddleware("moderator")).Patch("/", app.updatePostHandler)
			})
		})
		r.Route("/comments", func(r chi.Router) {
			r.Use(app.UserAuthMiddleware)
			r.Post("/", app.createCommentHandler)
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(app.UserAuthMiddleware)
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.GetTargetUserMiddlewareContext)
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
			r.Post("/login", app.authUserHandler)
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

	shutdown := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.logger.Infow("server has stopped", "addr", app.config.addr, "env", app.config.env)

	return nil
}
