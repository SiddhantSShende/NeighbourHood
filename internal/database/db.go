package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "github.com/lib/pq"
)

// DB is the shared database connection pool.
// Callers should not close it directly; call Close() when the application exits.
var DB *sql.DB

// InitDB opens the PostgreSQL connection pool and verifies connectivity.
// Connection details are read from DATABASE_URL (full DSN) or individual
// DB_HOST / DB_PORT / DB_USER / DB_PASSWORD / DB_NAME / DB_SSLMODE env vars.
func InitDB() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		host := getEnvOrDefault("DB_HOST", "localhost")
		port := getEnvOrDefault("DB_PORT", "5432")
		user := getEnvOrDefault("DB_USER", "postgres")
		password := os.Getenv("DB_PASSWORD") // no default — must be set explicitly
		dbName := getEnvOrDefault("DB_NAME", "neighbourhood")
		sslMode := getEnvOrDefault("DB_SSLMODE", "disable")
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbName, sslMode,
		)
	}

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// Configure connection pool for production workloads.
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)
	DB.SetConnMaxIdleTime(2 * time.Minute)

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

// RunMigrations executes the SQL schema file against the connected database.
// The schema file is resolved relative to the source file location so it works
// regardless of the working directory the binary is launched from.
func RunMigrations() error {
	schemaPath := resolveSchemaPath()

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("error reading schema file (%s): %w", schemaPath, err)
	}

	if _, err = DB.Exec(string(schema)); err != nil {
		return fmt.Errorf("error executing schema: %w", err)
	}

	log.Println("Database migrations run successfully")
	return nil
}

// resolveSchemaPath returns the absolute path to schema.sql.
// It first checks the conventional project-relative path, then falls back to
// a path relative to this source file (useful during development / tests).
func resolveSchemaPath() string {
	// 1. Honour an explicit override via env var.
	if p := os.Getenv("SCHEMA_PATH"); p != "" {
		return p
	}

	// 2. Path relative to the current working directory (typical for production
	//    binaries launched from the project root).
	cwdPath := filepath.Join("internal", "models", "schema.sql")
	if _, err := os.Stat(cwdPath); err == nil {
		return cwdPath
	}

	// 3. Path relative to this source file (works for `go test` and IDE runs).
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		return filepath.Join(filepath.Dir(filename), "..", "models", "schema.sql")
	}

	return cwdPath // best-effort fallback
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
