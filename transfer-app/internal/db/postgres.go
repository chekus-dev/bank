package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"transfer-app/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() error {
	var err error
	DB, err = sql.Open("postgres", config.App.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)
	DB.SetConnMaxIdleTime(2 * time.Minute)

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	log.Println("✅ Database connected")
	return nil
}

func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("❌ Error closing db: %v", err)
		}
	}
}

func BeginTx() (*sql.Tx, error) {
	return DB.Begin()
}
