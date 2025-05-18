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

func (db *DB) GetRandomQuotes(limit int) ([]Quote, error) {
	query := "SELECT id, quote, author, created_at FROM quotes ORDER BY RAND() LIMIT ?"
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying random quotes: %v", err)
	}
	defer rows.Close()

	var quotes []Quote
	for rows.Next() {
		var quote Quote
		if err := rows.Scan(&quote.ID, &quote.Quote, &quote.Author, &quote.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning quote row: %v", err)
		}
		quotes = append(quotes, quote)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating quote rows: %v", err)
	}

	return quotes, nil
}
