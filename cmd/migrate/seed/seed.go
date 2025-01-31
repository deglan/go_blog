package main

import (
	"log"
	"social/internal/db"
	"social/internal/env"
	"social/internal/store"
)

func main() {
	cfg := dbConfig{
		addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
		maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
		maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
		maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
	}

	database, err := db.New(
		cfg.addr,
		cfg.maxOpenConns,
		cfg.maxIdleConns,
		cfg.maxIdleTime,
	)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer database.Close()

	storage := store.NewStorage(database)

	log.Println("Seeding data into the database...")
	db.Seed(storage, database)
	log.Println("Seeding completed successfully.")
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}
