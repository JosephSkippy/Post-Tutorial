package main

import (
	"log"
	"tiago-udemy/internal/db"
	"tiago-udemy/internal/env"
	"tiago-udemy/internal/mailer"
	"tiago-udemy/internal/store"
	"time"

	"go.uber.org/zap"
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

	mailConfig := mailConfig{
		exp:       env.GetDuration("MAIL_TOKEN_EXPIRATION", 24*time.Hour), // 1 day
		fromEmail: env.GetString("MAIL_FROM_EMAIL", "hsa@noreply.com"),
		mailTrapConfig: mailTrapConfig{
			apiKey: env.GetString("MAILTRAP_API_KEY", ""),
		},
	}

	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		dbConfig:    dbConfig,
		env:         env.GetString("ENV", "developement"),
		apiURL:      env.GetString("API_URL", "localhost:3000"),
		mail:        mailConfig,
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:8080"),
	}

	//logger

	logger := zap.Must(zap.NewProduction()).Sugar()
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

	mailTrapperClient, err := mailer.NewMailTrapClient("123", "123")
	if err != nil {
		logger.Fatalf("Cannot create mailtrap client %v", err)
	}

	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
		mailer: &mailTrapperClient,
	}
	mux := app.mount()
	log.Fatal(app.run(mux))
}
