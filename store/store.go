package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BuildRecord struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	ImageName  string    `json:"image_name"`
	ImageTag   string    `json:"image_tag"`
	Status     string    `json:"status"` // pending, building, success, failed
	Log        string    `json:"log"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

func Init(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "flowci.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	DB = db
	return migrate()
}

func migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			language TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS build_records (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			image_name TEXT NOT NULL,
			image_tag TEXT NOT NULL DEFAULT 'latest',
			status TEXT NOT NULL DEFAULT 'pending',
			log TEXT NOT NULL DEFAULT '',
			started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			finished_at DATETIME,
			FOREIGN KEY (project_id) REFERENCES projects(id)
		)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
