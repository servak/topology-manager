package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate <up|down>")
		os.Exit(1)
	}

	command := os.Args[1]
	
	// PostgreSQL DSN を環境変数から取得
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://topology:topology@localhost/topology_manager?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "up":
		if err := migrateUp(db); err != nil {
			fmt.Printf("Migration up failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migration up completed successfully")
	case "down":
		if err := migrateDown(db); err != nil {
			fmt.Printf("Migration down failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migration down completed successfully")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Usage: migrate <up|down>")
		os.Exit(1)
	}
}

func migrateUp(db *sql.DB) error {
	// マイグレーションファイルを読み込み
	migrationDir := "internal/repository/postgres/migrations"
	files, err := ioutil.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// ファイル名でソート
	sort.Strings(migrationFiles)

	// マイグレーションテーブルを作成
	if err := createMigrationTable(db); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// 各マイグレーションファイルを実行
	for _, filename := range migrationFiles {
		if applied, err := isMigrationApplied(db, filename); err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", filename, err)
		} else if applied {
			fmt.Printf("Migration %s already applied, skipping\n", filename)
			continue
		}

		fmt.Printf("Applying migration %s...\n", filename)
		
		filePath := filepath.Join(migrationDir, filename)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// SQLを実行
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// マイグレーション履歴に記録
		if err := recordMigration(db, filename); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		fmt.Printf("Migration %s applied successfully\n", filename)
	}

	return nil
}

func migrateDown(db *sql.DB) error {
	// 簡単なダウンマイグレーション（全テーブル削除）
	fmt.Println("Dropping all tables...")
	
	dropQueries := []string{
		"DROP TABLE IF EXISTS links",
		"DROP TABLE IF EXISTS devices",
		"DROP TABLE IF EXISTS migrations",
		"DROP FUNCTION IF EXISTS update_updated_at_column",
	}

	for _, query := range dropQueries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute drop query: %w", err)
		}
	}

	return nil
}

func createMigrationTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	return err
}

func isMigrationApplied(db *sql.DB, filename string) (bool, error) {
	query := "SELECT COUNT(*) FROM migrations WHERE filename = $1"
	var count int
	err := db.QueryRow(query, filename).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func recordMigration(db *sql.DB, filename string) error {
	query := "INSERT INTO migrations (filename) VALUES ($1)"
	_, err := db.Exec(query, filename)
	return err
}