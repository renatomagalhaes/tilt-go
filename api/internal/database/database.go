package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Quote struct {
	ID        int       `json:"id"`
	Quote     string    `json:"quote"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

type DB struct {
	*sql.DB
}

func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{db}, nil
}

func (db *DB) GetRandomQuote() (*Quote, error) {
	var quote Quote
	err := db.QueryRow("SELECT id, quote, author, created_at FROM quotes ORDER BY RAND() LIMIT 1").
		Scan(&quote.ID, &quote.Quote, &quote.Author, &quote.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("error getting random quote: %v", err)
	}
	return &quote, nil
}

func (db *DB) CheckConnection() error {
	return db.Ping()
}
