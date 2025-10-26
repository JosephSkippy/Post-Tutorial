package main

import (
	"log"
	"tiago-udemy/internal/auth"
	"tiago-udemy/internal/db"
	"tiago-udemy/internal/env"
	"tiago-udemy/internal/mailer"
	"tiago-udemy/internal/ratelimiter"
	"tiago-udemy/internal/store"
	"tiago-udemy/internal/store/cache"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const version = "0.01"

//	@title			For Tiago Udemy Course API
//	@description	API for Social Media Application
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @host						petstore.swagger.io
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
func main() {
	dbConfig := dbConfig{
		addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
		maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
		maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
		maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
	}

	cacheConfig := cacheConfig{
		redis: redisConfig{
			addr: env.GetString("REDDIS_ADDR", "localhost:6379"),
		},
		enabled: env.GetBool("CACHE_ENABLED", false),
	}

	mailConfig := mailConfig{
		exp:       env.GetDuration("MAIL_TOKEN_EXPIRATION", 24*time.Hour), // 1 day
		fromEmail: env.GetString("MAIL_FROM_EMAIL", ""),
		mailTrapConfig: mailTrapConfig{
			apiKey: env.GetString("MAILTRAP_API_KEY", ""),
		},
	}

	authConfig := authConfig{
		basicAuth: basicAuth{
			username: env.GetString("BASIC_USERNAME", "admin1234"),
			password: env.GetString("BASIC_PASSWORD", "admin1234"),
		},
		jwtAuth: jwtAuth{
			iss:    env.GetString("JWT_ISSUER", "tiago-udemy"),
			exp:    env.GetDuration("JWT_EXPIRATION", 1*time.Hour),
			secret: env.GetString("JWT_SECRET", "tiago"),
		},
	}

	limiterConfig := limiterConfig{
		window:     env.GetDuration("JWT_EXPIRATION", 1*time.Hour),
		maxRequest: env.GetInt("DB_MAX_IDLE_CONNS", 200),
	}

	cfg := config{
		addr:          env.GetString("ADDR", ":8080"),
		dbConfig:      dbConfig,
		env:           env.GetString("ENV", "developement"),
		apiURL:        env.GetString("API_URL", "localhost:3000"),
		mail:          mailConfig,
		frontendURL:   env.GetString("FRONTEND_URL", "http://localhost:8080"),
		authConfig:    authConfig,
		cacheConfig:   cacheConfig,
		limiterConfig: limiterConfig,
	}

	//logger
	var logger *zap.SugaredLogger
	if cfg.env == "production" {
		logger = zap.Must(zap.NewProduction()).Sugar()
	} else {
		config := zap.NewDevelopmentConfig()
		config.OutputPaths = []string{"stdout"}
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		devLogger, err := config.Build()
		if err != nil {
			log.Fatalf("Cannot initialize logger: %v", err)
		}
		logger = devLogger.Sugar()
	}
	defer logger.Sync() // flushes buffer, if any

	//database connection
	db, err := db.NewDBConnection(dbConfig.addr, dbConfig.maxOpenConns, dbConfig.maxIdleConns, dbConfig.maxIdleTime)

	if err != nil {
		logger.Fatalf("Cannot connect to database %v", err)
	}
	logger.Info("Database connection established")
	defer db.Close()

	store := store.NewStorage(db)

	//email client

	mailTrapperClient, err := mailer.NewMailTrapClient(mailConfig.mailTrapConfig.apiKey, mailConfig.fromEmail)
	if err != nil {
		logger.Fatalf("Cannot create mailtrap client %v", err)
	}

	// authentication
	authenticator := auth.NewJWTAuthenticator(authConfig.jwtAuth.secret, authConfig.jwtAuth.iss, authConfig.jwtAuth.iss)

	// cache
	var cacheStore cache.CacheStorage
	if cfg.cacheConfig.enabled {
		rdb := cache.NewRedisClient(cfg.cacheConfig.redis.addr, "", 0)
		defer rdb.Close()
		logger.Info("Redis cache client initialized")
		cacheStore = cache.RedisStore(rdb)
	} else {
		logger.Info("Redis cache is disabled")
		cacheStore = cache.NewNoOpStore() // We'll add this next
	}

	// limiter client
	ratelimiterClient := ratelimiter.NewFixedWindowLimiter(limiterConfig.maxRequest, limiterConfig.window)

	app := &application{
		config:        cfg,
		store:         store,
		logger:        logger,
		mailer:        &mailTrapperClient,
		authenticator: authenticator,
		cache:         cacheStore,
		limiter:       ratelimiterClient,
	}
	mux := app.mount()
	log.Fatal(app.run(mux))
}
