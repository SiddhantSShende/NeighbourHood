package postgres

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"neighbourhood/services/auth/internal/config"
	"neighbourhood/services/auth/internal/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

// Make sure PostgresRepository implements both User and OAuth repository interfaces
var (
	_ domain.UserRepository  = (*PostgresRepository)(nil)
	_ domain.OAuthRepository = (*PostgresRepository)(nil)
)

func New(cfg config.DatabaseConfig) (*PostgresRepository, error) {
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, sslMode,
	)

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

	// Initialize schema
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
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255),
		first_name VARCHAR(100),
		last_name VARCHAR(100),
		avatar_url VARCHAR(500),
		email_verified BOOLEAN DEFAULT FALSE,
		active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS oauth_accounts (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		provider VARCHAR(50) NOT NULL,
		provider_id VARCHAR(255) NOT NULL,
		email VARCHAR(255),
		access_token TEXT,
		refresh_token TEXT,
		expires_at TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE(provider, provider_id)
	);

	CREATE INDEX IF NOT EXISTS idx_oauth_accounts_user_id ON oauth_accounts(user_id);
	CREATE INDEX IF NOT EXISTS idx_oauth_accounts_provider ON oauth_accounts(provider, provider_id);
	`

	_, err := r.db.Exec(schema)
	return err
}

// User repository methods

func (r *PostgresRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, avatar_url, email_verified, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.EmailVerified,
		user.Active,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, avatar_url, email_verified, active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.EmailVerified,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) GetByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, avatar_url, email_verified, active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &domain.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.EmailVerified,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, first_name = $3, last_name = $4,
		    avatar_url = $5, email_verified = $6, active = $7, updated_at = $8
		WHERE id = $9
	`

	_, err := r.db.Exec(query,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.EmailVerified,
		user.Active,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Delete(id string) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// OAuth repository methods

func (r *PostgresRepository) CreateOAuth(account *domain.OAuthAccount) error {
	query := `
		INSERT INTO oauth_accounts (id, user_id, provider, provider_id, email, access_token, refresh_token, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(query,
		account.ID,
		account.UserID,
		account.Provider,
		account.ProviderID,
		account.Email,
		account.AccessToken,
		account.RefreshToken,
		account.ExpiresAt,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create OAuth account: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByProviderAndID(provider, providerID string) (*domain.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_id, email, access_token, refresh_token, expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE provider = $1 AND provider_id = $2
	`

	account := &domain.OAuthAccount{}
	err := r.db.QueryRow(query, provider, providerID).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderID,
		&account.Email,
		&account.AccessToken,
		&account.RefreshToken,
		&account.ExpiresAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("OAuth account not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth account: %w", err)
	}

	return account, nil
}

func (r *PostgresRepository) GetByUserID(userID string) ([]*domain.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_id, email, access_token, refresh_token, expires_at, created_at, updated_at
		FROM oauth_accounts
		WHERE user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query OAuth accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.OAuthAccount

	for rows.Next() {
		account := &domain.OAuthAccount{}
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.ProviderID,
			&account.Email,
			&account.AccessToken,
			&account.RefreshToken,
			&account.ExpiresAt,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan OAuth account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *PostgresRepository) UpdateOAuth(account *domain.OAuthAccount) error {
	query := `
		UPDATE oauth_accounts
		SET access_token = $1, refresh_token = $2, expires_at = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(query,
		account.AccessToken,
		account.RefreshToken,
		account.ExpiresAt,
		time.Now(),
		account.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update OAuth account: %w", err)
	}

	return nil
}

func (r *PostgresRepository) DeleteOAuth(id string) error {
	query := `DELETE FROM oauth_accounts WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete OAuth account: %w", err)
	}

	return nil
}
