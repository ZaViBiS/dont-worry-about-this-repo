package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ZaViBiS/dont-worry-about-this-repo/internal/config"

	_ "modernc.org/sqlite"
)

type Record struct {
	ID        int64
	Timestamp time.Time
}

func DBInit(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("db init error: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return nil, fmt.Errorf("creating first table error: %w", err)
	}

	return db, nil
}

// Add створює новий запис з монотонно зростаючим id, який гарантує послідовність записів.
// Повертає Record зі згенерованим id та timestamp з бази даних за допомогою RETURNING.
func Add(ctx context.Context, db *sql.DB) (Record, error) {
	var rec Record
	row := db.QueryRowContext(ctx, `INSERT INTO records DEFAULT VALUES RETURNING id, timestamp`)
	if err := row.Scan(&rec.ID, &rec.Timestamp); err != nil {
		return Record{}, fmt.Errorf("insert record: %w", err)
	}

	return rec, nil
}
