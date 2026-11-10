package cache

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"

	_ "modernc.org/sqlite"
)

type Cache struct {
	db *sql.DB
}

func NewCache(dbPath string) (*Cache, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS response_cache (
		hash TEXT PRIMARY KEY,
		response TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("create cache table: %w", err)
	}
	return &Cache{db: db}, nil
}

func (c *Cache) Get(ctx context.Context, reqBody []byte) (string, bool, error) {
	hash := hashBody(reqBody)

	var response string
	query := `SELECT response FROM response_cache WHERE hash = ?`
	err := c.db.QueryRowContext(ctx, query, hash).Scan(&response)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("cache get: %w", err)
	}

	return response, true, nil
}

func (c *Cache) Set(ctx context.Context, reqBody []byte, response string) error {
	hash := hashBody(reqBody)

	query := `INSERT OR REPLACE INTO response_cache (hash, response) VALUES (?, ?)`
	if _, err := c.db.ExecContext(ctx, query, hash, response); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}

	return nil
}

func (c *Cache) Close() error {
	return c.db.Close()
}

func hashBody(body []byte) string {
	hasher := sha256.New()
	hasher.Write(body)
	return hex.EncodeToString(hasher.Sum(nil))
}
