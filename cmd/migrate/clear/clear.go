package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"social/internal/db"
	"social/internal/env"
	"time"
)

func main() {
	// Konfiguracja połączenia z bazą danych
	cfg := dbConfig{
		addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
		maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
		maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
		maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
	}

	// Połączenie z bazą danych
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

	// Wykonanie czyszczenia bazy danych
	err = cleanDatabase(database)
	if err != nil {
		log.Fatalf("Failed to clean the database: %v", err)
	}

	log.Println("Database cleaned successfully!")
}

func cleanDatabase(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout())
	defer cancel()

	tables := []string{"comments", "posts", "users"}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to clean table %s: %w", table, err)
		}
		log.Printf("Table %s cleaned successfully", table)
	}

	return nil
}

func dbTimeout() time.Duration {
	return time.Duration(env.GetInt("DB_TIMEOUT", 5)) * time.Second // Timeout w sekundach
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}
