package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"neighbourhood/services/auth/internal/domain"
)

// RBACRepository implements domain.RBACRepository with optimized performance
type RBACRepository struct {
	db *sql.DB
}

func NewRBACRepository(db *sql.DB) *RBACRepository {
	return &RBACRepository{db: db}
}

// CreateUserRole - O(1) insert operation
func (r *RBACRepository) CreateUserRole(userRole *domain.UserRole) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	permJSON, err := json.Marshal(userRole.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	query := `
		INSERT INTO user_roles (id, user_id, workspace_id, role, permissions, created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.ExecContext(ctx, query,
		userRole.ID,
		userRole.UserID,
		userRole.WorkspaceID,
		userRole.Role,
		permJSON,
		userRole.CreatedAt,
		userRole.UpdatedAt,
		userRole.CreatedBy,
	)

	return err
}

// HasPermission - O(1) average case using indexed lookup + in-memory check
func (r *RBACRepository) HasPermission(userID, workspaceID string, permission domain.Permission) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use composite index (user_id, workspace_id) for O(1) lookup
	query := `
		SELECT role, permissions 
		FROM user_roles 
		WHERE user_id = $1 AND workspace_id = $2
		LIMIT 1
	`

	var role domain.Role
	var permJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID, workspaceID).Scan(&role, &permJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	// Check custom permissions - O(n) where n is small (< 20 typically)
	var customPerms []domain.Permission
	if len(permJSON) > 0 {
		if err := json.Unmarshal(permJSON, &customPerms); err == nil {
			for _, p := range customPerms {
				if p == permission {
					return true, nil
				}
			}
		}
	}

	// Check default role permissions - O(n) where n is small
	defaultPerms, exists := domain.RolePermissions[role]
	if !exists {
		return false, nil
	}

	for _, p := range defaultPerms {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// GetUserRole - O(1) with composite index
func (r *RBACRepository) GetUserRole(userID, workspaceID string) (*domain.UserRole, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, workspace_id, role, permissions, created_at, updated_at, created_by
		FROM user_roles
		WHERE user_id = $1 AND workspace_id = $2
	`

	var ur domain.UserRole
	var permJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID, workspaceID).Scan(
		&ur.ID,
		&ur.UserID,
		&ur.WorkspaceID,
		&ur.Role,
		&permJSON,
		&ur.CreatedAt,
		&ur.UpdatedAt,
		&ur.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	if len(permJSON) > 0 {
		json.Unmarshal(permJSON, &ur.Permissions)
	}

	return &ur, nil
}

// GetUserRoles - O(k) where k is number of workspaces user belongs to
func (r *RBACRepository) GetUserRoles(userID string) ([]*domain.UserRole, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, workspace_id, role, permissions, created_at, updated_at, created_by
		FROM user_roles
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.UserRole
	for rows.Next() {
		var ur domain.UserRole
		var permJSON []byte

		err := rows.Scan(
			&ur.ID,
			&ur.UserID,
			&ur.WorkspaceID,
			&ur.Role,
			&permJSON,
			&ur.CreatedAt,
			&ur.UpdatedAt,
			&ur.CreatedBy,
		)
		if err != nil {
			return nil, err
		}

		if len(permJSON) > 0 {
			json.Unmarshal(permJSON, &ur.Permissions)
		}

		roles = append(roles, &ur)
	}

	return roles, rows.Err()
}

// UpdateUserRole - O(1) update
func (r *RBACRepository) UpdateUserRole(userRole *domain.UserRole) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	permJSON, err := json.Marshal(userRole.Permissions)
	if err != nil {
		return err
	}

	query := `
		UPDATE user_roles 
		SET role = $1, permissions = $2, updated_at = $3
		WHERE user_id = $4 AND workspace_id = $5
	`

	_, err = r.db.ExecContext(ctx, query,
		userRole.Role,
		permJSON,
		time.Now(),
		userRole.UserID,
		userRole.WorkspaceID,
	)

	return err
}

// DeleteUserRole - O(1) delete
func (r *RBACRepository) DeleteUserRole(userID, workspaceID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM user_roles WHERE user_id = $1 AND workspace_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, workspaceID)
	return err
}

// CreateWorkspace - O(1) insert
func (r *RBACRepository) CreateWorkspace(workspace *domain.Workspace) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settingsJSON, _ := json.Marshal(workspace.Settings)

	query := `
		INSERT INTO workspaces (id, name, owner_id, description, active, plan, created_at, updated_at, settings)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		workspace.ID,
		workspace.Name,
		workspace.OwnerID,
		workspace.Description,
		workspace.Active,
		workspace.Plan,
		workspace.CreatedAt,
		workspace.UpdatedAt,
		settingsJSON,
	)

	return err
}

// GetWorkspace - O(1) with primary key index
func (r *RBACRepository) GetWorkspace(workspaceID string) (*domain.Workspace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, name, owner_id, description, active, plan, created_at, updated_at, settings
		FROM workspaces
		WHERE id = $1
	`

	var ws domain.Workspace
	var settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, workspaceID).Scan(
		&ws.ID,
		&ws.Name,
		&ws.OwnerID,
		&ws.Description,
		&ws.Active,
		&ws.Plan,
		&ws.CreatedAt,
		&ws.UpdatedAt,
		&settingsJSON,
	)

	if err != nil {
		return nil, err
	}

	if len(settingsJSON) > 0 {
		json.Unmarshal(settingsJSON, &ws.Settings)
	}

	return &ws, nil
}

// GetUserWorkspaces - O(k) where k is number of user's workspaces
func (r *RBACRepository) GetUserWorkspaces(userID string) ([]*domain.Workspace, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT w.id, w.name, w.owner_id, w.description, w.active, w.plan, w.created_at, w.updated_at, w.settings
		FROM workspaces w
		INNER JOIN user_roles ur ON w.id = ur.workspace_id
		WHERE ur.user_id = $1
		ORDER BY w.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var ws domain.Workspace
		var settingsJSON []byte

		err := rows.Scan(
			&ws.ID,
			&ws.Name,
			&ws.OwnerID,
			&ws.Description,
			&ws.Active,
			&ws.Plan,
			&ws.CreatedAt,
			&ws.UpdatedAt,
			&settingsJSON,
		)
		if err != nil {
			return nil, err
		}

		if len(settingsJSON) > 0 {
			json.Unmarshal(settingsJSON, &ws.Settings)
		}

		workspaces = append(workspaces, &ws)
	}

	return workspaces, rows.Err()
}

// UpdateWorkspace - O(1) update
func (r *RBACRepository) UpdateWorkspace(workspace *domain.Workspace) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	settingsJSON, _ := json.Marshal(workspace.Settings)

	query := `
		UPDATE workspaces 
		SET name = $1, description = $2, active = $3, plan = $4, updated_at = $5, settings = $6
		WHERE id = $7
	`

	_, err := r.db.ExecContext(ctx, query,
		workspace.Name,
		workspace.Description,
		workspace.Active,
		workspace.Plan,
		time.Now(),
		settingsJSON,
		workspace.ID,
	)

	return err
}

// DeleteWorkspace - O(1) delete
func (r *RBACRepository) DeleteWorkspace(workspaceID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM workspaces WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, workspaceID)
	return err
}

// CreateAPIKey - O(1) insert
func (r *RBACRepository) CreateAPIKey(apiKey *domain.APIKey) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scopesJSON, _ := json.Marshal(apiKey.Scopes)

	query := `
		INSERT INTO api_keys (id, workspace_id, user_id, name, key_prefix, key_hash, expires_at, active, scopes, rate_limit, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		apiKey.ID,
		apiKey.WorkspaceID,
		apiKey.UserID,
		apiKey.Name,
		apiKey.KeyPrefix,
		apiKey.KeyHash,
		apiKey.ExpiresAt,
		apiKey.Active,
		scopesJSON,
		apiKey.RateLimit,
		apiKey.CreatedAt,
	)

	return err
}

// ValidateAPIKey - O(1) with hash index lookup
func (r *RBACRepository) ValidateAPIKey(keyHash string) (*domain.APIKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Hash lookup - O(1) with proper index
	query := `
		SELECT id, workspace_id, user_id, name, key_prefix, key_hash, last_used_at, expires_at, active, scopes, rate_limit, created_at, revoked_at
		FROM api_keys
		WHERE key_hash = $1 AND active = true
	`

	var ak domain.APIKey
	var scopesJSON []byte
	var lastUsed, expires, revoked sql.NullTime

	err := r.db.QueryRowContext(ctx, query, keyHash).Scan(
		&ak.ID,
		&ak.WorkspaceID,
		&ak.UserID,
		&ak.Name,
		&ak.KeyPrefix,
		&ak.KeyHash,
		&lastUsed,
		&expires,
		&ak.Active,
		&scopesJSON,
		&ak.RateLimit,
		&ak.CreatedAt,
		&revoked,
	)

	if err != nil {
		return nil, err
	}

	if lastUsed.Valid {
		ak.LastUsedAt = &lastUsed.Time
	}
	if expires.Valid {
		ak.ExpiresAt = &expires.Time
		// Check expiration
		if time.Now().After(*ak.ExpiresAt) {
			return nil, fmt.Errorf("API key expired")
		}
	}
	if revoked.Valid {
		ak.RevokedAt = &revoked.Time
	}

	if len(scopesJSON) > 0 {
		json.Unmarshal(scopesJSON, &ak.Scopes)
	}

	// Update last used timestamp asynchronously (fire and forget for performance)
	go r.updateLastUsed(ak.ID)

	return &ak, nil
}

func (r *RBACRepository) updateLastUsed(keyID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	query := `UPDATE api_keys SET last_used_at = $1 WHERE id = $2`
	r.db.ExecContext(ctx, query, time.Now(), keyID)
}

// GetAPIKey - O(1) lookup by prefix
func (r *RBACRepository) GetAPIKey(keyPrefix string) (*domain.APIKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, workspace_id, user_id, name, key_prefix, key_hash, last_used_at, expires_at, active, scopes, rate_limit, created_at, revoked_at
		FROM api_keys
		WHERE key_prefix = $1
	`

	var ak domain.APIKey
	var scopesJSON []byte
	var lastUsed, expires, revoked sql.NullTime

	err := r.db.QueryRowContext(ctx, query, keyPrefix).Scan(
		&ak.ID,
		&ak.WorkspaceID,
		&ak.UserID,
		&ak.Name,
		&ak.KeyPrefix,
		&ak.KeyHash,
		&lastUsed,
		&expires,
		&ak.Active,
		&scopesJSON,
		&ak.RateLimit,
		&ak.CreatedAt,
		&revoked,
	)

	if err != nil {
		return nil, err
	}

	if lastUsed.Valid {
		ak.LastUsedAt = &lastUsed.Time
	}
	if expires.Valid {
		ak.ExpiresAt = &expires.Time
	}
	if revoked.Valid {
		ak.RevokedAt = &revoked.Time
	}

	if len(scopesJSON) > 0 {
		json.Unmarshal(scopesJSON, &ak.Scopes)
	}

	return &ak, nil
}

// GetWorkspaceAPIKeys - O(k) where k is number of workspace's API keys
func (r *RBACRepository) GetWorkspaceAPIKeys(workspaceID string) ([]*domain.APIKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, workspace_id, user_id, name, key_prefix, key_hash, last_used_at, expires_at, active, scopes, rate_limit, created_at, revoked_at
		FROM api_keys
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apiKeys []*domain.APIKey
	for rows.Next() {
		var ak domain.APIKey
		var scopesJSON []byte
		var lastUsed, expires, revoked sql.NullTime

		err := rows.Scan(
			&ak.ID,
			&ak.WorkspaceID,
			&ak.UserID,
			&ak.Name,
			&ak.KeyPrefix,
			&ak.KeyHash,
			&lastUsed,
			&expires,
			&ak.Active,
			&scopesJSON,
			&ak.RateLimit,
			&ak.CreatedAt,
			&revoked,
		)
		if err != nil {
			return nil, err
		}

		if lastUsed.Valid {
			ak.LastUsedAt = &lastUsed.Time
		}
		if expires.Valid {
			ak.ExpiresAt = &expires.Time
		}
		if revoked.Valid {
			ak.RevokedAt = &revoked.Time
		}

		if len(scopesJSON) > 0 {
			json.Unmarshal(scopesJSON, &ak.Scopes)
		}

		apiKeys = append(apiKeys, &ak)
	}

	return apiKeys, rows.Err()
}

// UpdateAPIKey - O(1) update
func (r *RBACRepository) UpdateAPIKey(apiKey *domain.APIKey) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	scopesJSON, _ := json.Marshal(apiKey.Scopes)

	query := `
		UPDATE api_keys 
		SET name = $1, active = $2, scopes = $3, rate_limit = $4
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query,
		apiKey.Name,
		apiKey.Active,
		scopesJSON,
		apiKey.RateLimit,
		apiKey.ID,
	)

	return err
}

// RevokeAPIKey - O(1) soft delete
func (r *RBACRepository) RevokeAPIKey(keyID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE api_keys SET active = false, revoked_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), keyID)
	return err
}

// HashAPIKey creates a SHA-256 hash of the API key for secure storage
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
