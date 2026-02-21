package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"neighbourhood/services/integration/internal/config"
	"neighbourhood/services/integration/internal/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func New(cfg config.DatabaseConfig) (*PostgresRepository, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &PostgresRepository{db: db}

	if err := repo.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresRepository) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS providers (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		category VARCHAR(100) NOT NULL,
		description TEXT,
		auth_type VARCHAR(50) NOT NULL,
		enabled BOOLEAN DEFAULT true,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS user_integrations (
		id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		provider_id VARCHAR(255) NOT NULL,
		access_token TEXT NOT NULL,
		refresh_token TEXT,
		expires_at TIMESTAMP,
		scopes TEXT[],
		metadata JSONB,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE(user_id, provider_id)
	);

	CREATE INDEX IF NOT EXISTS idx_user_integrations_user_id ON user_integrations(user_id);
	CREATE INDEX IF NOT EXISTS idx_providers_category ON providers(category);
	`

	_, err := r.db.Exec(schema)
	return err
}

// Provider methods

func (r *PostgresRepository) GetAll() ([]*domain.Provider, error) {
	query := `SELECT id, name, category, description, auth_type, enabled, created_at, updated_at 
	          FROM providers WHERE enabled = true ORDER BY category, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*domain.Provider
	for rows.Next() {
		var p domain.Provider
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Description, &p.AuthType, &p.Enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		providers = append(providers, &p)
	}

	return providers, nil
}

func (r *PostgresRepository) GetByID(id string) (*domain.Provider, error) {
	query := `SELECT id, name, category, description, auth_type, enabled, created_at, updated_at 
	          FROM providers WHERE id = $1`

	var p domain.Provider
	err := r.db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Category, &p.Description, &p.AuthType, &p.Enabled, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("provider not found")
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *PostgresRepository) GetByCategory(category string) ([]*domain.Provider, error) {
	query := `SELECT id, name, category, description, auth_type, enabled, created_at, updated_at 
	          FROM providers WHERE category = $1 AND enabled = true ORDER BY name`

	rows, err := r.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*domain.Provider
	for rows.Next() {
		var p domain.Provider
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Description, &p.AuthType, &p.Enabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		providers = append(providers, &p)
	}

	return providers, nil
}
