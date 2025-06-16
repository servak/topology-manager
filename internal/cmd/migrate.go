package cmd

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
	"github.com/servak/topology-manager/internal/config"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [up|down]",
	Short: "Run database migrations",
	Long:  "Run database migrations to set up or tear down the database schema",
	Args:  cobra.ExactArgs(1),
	Run:   runMigrate,
}

func runMigrate(cmd *cobra.Command, args []string) {
	command := args[0]

	// 設定ファイルを読み込み
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// データベース設定を確認
	if cfg.Database.Type != "postgres" {
		log.Fatalf("Migration currently only supports PostgreSQL. Current type: %s", cfg.Database.Type)
	}

	// PostgreSQL DSN を設定ファイルまたは環境変数から取得
	var dsn string
	if dsnFromEnv := os.Getenv("DATABASE_URL"); dsnFromEnv != "" {
		dsn = dsnFromEnv
		if verbose {
			log.Printf("Using DATABASE_URL from environment variable")
		}
	} else {
		// 設定ファイルからDSNを構築
		dsn = cfg.Database.Postgres.BuildDSN()
		if verbose {
			log.Printf("Using database configuration from config file")
		}
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if verbose {
		log.Println("Connected to PostgreSQL")
	}

	switch command {
	case "up":
		if err := migrateUp(db); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migration up completed successfully")
	case "down":
		if err := migrateDown(db); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Migration down completed successfully")
	default:
		log.Fatalf("Unknown command: %s. Use 'up' or 'down'", command)
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
			if verbose {
				log.Printf("Migration %s already applied, skipping", filename)
			}
			continue
		}

		log.Printf("Applying migration %s...", filename)

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

		log.Printf("Migration %s applied successfully", filename)
	}

	return nil
}

func migrateDown(db *sql.DB) error {
	// 簡単なダウンマイグレーション（全テーブル削除）
	log.Println("Dropping all tables...")

	dropQueries := []string{
		"DROP TABLE IF EXISTS links",
		"DROP TABLE IF EXISTS devices",
		"DROP TABLE IF EXISTS migrations",
		"DROP FUNCTION IF EXISTS update_updated_at_column",
	}

	for _, query := range dropQueries {
		if verbose {
			log.Printf("Executing: %s", query)
		}
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
