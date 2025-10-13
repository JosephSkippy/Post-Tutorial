package main

import (
	"log"
	"tiago-udemy/internal/db"
	"tiago-udemy/internal/env"
	"tiago-udemy/internal/store"
)

const version = "0.01"

func main() {
	dbConfig := dbConfig{
		addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
		maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
		maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
		maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
	}

	cfg := config{
		addr:     env.GetString("ADDR", ":8080"),
		dbConfig: dbConfig,
		env:      env.GetString("ENV", "developement"),
	}

	db, err := db.NewDBConnection(dbConfig.addr, dbConfig.maxOpenConns, dbConfig.maxIdleConns, dbConfig.maxIdleTime)

	if err != nil {
		log.Fatalf("Cannot connect to database %v", err)
	}
	log.Println("Database connection established")
	defer db.Close()

	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}
	mux := app.mount()
	log.Fatal(app.run(mux))
}
