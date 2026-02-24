package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"

	"neighbourhood/services/auth/internal/domain"
	"neighbourhood/services/auth/internal/repository/postgres"
)

type RBACUseCase struct {
	rbacRepo domain.RBACRepository
	logger   Logger
}

func NewRBACUseCase(rbacRepo domain.RBACRepository, logger Logger) *RBACUseCase {
	return &RBACUseCase{
		rbacRepo: rbacRepo,
		logger:   logger,
	}
}

// CreateWorkspace creates a new developer workspace with the owner as admin
// Time Complexity: O(1) for workspace creation + O(1) for role assignment = O(1) total
func (uc *RBACUseCase) CreateWorkspace(ctx context.Context, ownerID, name, description string) (*domain.Workspace, error) {
	// Generate unique workspace ID
	workspaceID := uuid.New().String()

	workspace := &domain.Workspace{
		ID:          workspaceID,
		Name:        name,
		OwnerID:     ownerID,
		Description: description,
		Active:      true,
		Plan:        "free",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Settings:    make(map[string]interface{}),
	}

	// Create workspace - O(1)
	if err := uc.rbacRepo.CreateWorkspace(workspace); err != nil {
		uc.logger.Error("Failed to create workspace", "error", err, "owner_id", ownerID)
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Assign owner as admin - O(1)
	ownerRole := &domain.UserRole{
		ID:          uuid.New().String(),
		UserID:      ownerID,
		WorkspaceID: workspaceID,
		Role:        domain.RoleAdmin,
		Permissions: []domain.Permission{}, // Uses default admin permissions
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   ownerID,
	}

	if err := uc.rbacRepo.CreateUserRole(ownerRole); err != nil {
		uc.logger.Error("Failed to assign owner role", "error", err, "workspace_id", workspaceID)
		// Rollback workspace creation
		if delErr := uc.rbacRepo.DeleteWorkspace(workspaceID); delErr != nil {
			uc.logger.Error("Failed to rollback workspace creation", "error", delErr, "workspace_id", workspaceID)
		}
		return nil, fmt.Errorf("failed to assign owner role: %w", err)
	}

	uc.logger.Info("Workspace created successfully", "workspace_id", workspaceID, "owner_id", ownerID)
	return workspace, nil
}

// GenerateAPIKey creates a unique API key for a developer
// Time Complexity: O(1) - single insert with indexed hash
// Space Complexity: O(1) - constant size key
func (uc *RBACUseCase) GenerateAPIKey(ctx context.Context, workspaceID, userID, name string, scopes []string, rateLimit int) (*domain.APIKey, string, error) {
	// Generate secure random key - cryptographically secure
	keyBytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate random key: %w", err)
	}

	// Encode as base64 for URL-safe transmission
	keySecret := base64.URLEncoding.EncodeToString(keyBytes)

	// Create prefix for easy identification
	prefix := fmt.Sprintf("nh_live_pk_%s", keySecret[:12])

	// Full key for user (shown only once)
	fullKey := fmt.Sprintf("%s_%s", prefix, keySecret)

	// Hash for secure storage - O(1) hash operation
	keyHash := postgres.HashAPIKey(fullKey)

	apiKey := &domain.APIKey{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Name:        name,
		KeyPrefix:   prefix,
		KeyHash:     keyHash,
		Active:      true,
		Scopes:      scopes,
		RateLimit:   rateLimit,
		CreatedAt:   time.Now(),
	}

	// Insert with indexed hash - O(1)
	if err := uc.rbacRepo.CreateAPIKey(apiKey); err != nil {
		uc.logger.Error("Failed to create API key", "error", err, "workspace_id", workspaceID)
		return nil, "", fmt.Errorf("failed to create API key: %w", err)
	}

	uc.logger.Info("API key generated", "key_id", apiKey.ID, "workspace_id", workspaceID, "user_id", userID)

	// Return API key object and full key (full key shown only once)
	return apiKey, fullKey, nil
}

// ValidateAPIKey validates and returns the API key details
// Time Complexity: O(1) - hash index lookup
func (uc *RBACUseCase) ValidateAPIKey(ctx context.Context, apiKey string) (*domain.APIKey, error) {
	// Hash the provided key - O(1)
	keyHash := postgres.HashAPIKey(apiKey)

	// Lookup by hash - O(1) with hash index
	ak, err := uc.rbacRepo.ValidateAPIKey(keyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	// Check if expired
	if ak.ExpiresAt != nil && time.Now().After(*ak.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
	}

	// Check if revoked
	if ak.RevokedAt != nil {
		return nil, fmt.Errorf("API key revoked")
	}

	return ak, nil
}

// CheckPermission verifies if a user has a specific permission in a workspace
// Time Complexity: O(1) average case - indexed lookup + small permission array scan
// Space Complexity: O(1) - constant memory usage
func (uc *RBACUseCase) CheckPermission(ctx context.Context, userID, workspaceID string, permission domain.Permission) (bool, error) {
	// Indexed lookup - O(1) with composite index (user_id, workspace_id)
	hasPermission, err := uc.rbacRepo.HasPermission(userID, workspaceID, permission)
	if err != nil {
		uc.logger.Error("Permission check failed", "error", err, "user_id", userID, "workspace_id", workspaceID)
		return false, err
	}

	return hasPermission, nil
}

// AssignRole assigns or updates a user's role in a workspace
// Time Complexity: O(1) - single indexed operation
func (uc *RBACUseCase) AssignRole(ctx context.Context, adminUserID, targetUserID, workspaceID string, role domain.Role, customPermissions []domain.Permission) error {
	// First check if admin has permission to assign roles
	hasPermission, err := uc.CheckPermission(ctx, adminUserID, workspaceID, domain.PermissionUserWrite)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("insufficient permissions to assign roles")
	}

	// Check if user role already exists
	existingRole, err := uc.rbacRepo.GetUserRole(targetUserID, workspaceID)
	if err != nil {
		uc.logger.Error("Failed to check existing user role", "error", err, "user_id", targetUserID, "workspace_id", workspaceID)
		// Treat as not found and proceed to create
		existingRole = nil
	}

	if existingRole != nil {
		// Update existing role - O(1)
		existingRole.Role = role
		existingRole.Permissions = customPermissions
		existingRole.UpdatedAt = time.Now()

		if err := uc.rbacRepo.UpdateUserRole(existingRole); err != nil {
			uc.logger.Error("Failed to update user role", "error", err, "user_id", targetUserID)
			return fmt.Errorf("failed to update role: %w", err)
		}

		uc.logger.Info("User role updated", "user_id", targetUserID, "workspace_id", workspaceID, "role", role)
	} else {
		// Create new role assignment - O(1)
		userRole := &domain.UserRole{
			ID:          uuid.New().String(),
			UserID:      targetUserID,
			WorkspaceID: workspaceID,
			Role:        role,
			Permissions: customPermissions,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			CreatedBy:   adminUserID,
		}

		if err := uc.rbacRepo.CreateUserRole(userRole); err != nil {
			uc.logger.Error("Failed to create user role", "error", err, "user_id", targetUserID)
			return fmt.Errorf("failed to assign role: %w", err)
		}

		uc.logger.Info("User role assigned", "user_id", targetUserID, "workspace_id", workspaceID, "role", role)
	}

	return nil
}

// RevokeAPIKey revokes an API key
// Time Complexity: O(1) - indexed update
func (uc *RBACUseCase) RevokeAPIKey(ctx context.Context, userID, workspaceID, keyID string) error {
	// Check permission
	hasPermission, err := uc.CheckPermission(ctx, userID, workspaceID, domain.PermissionAPIKeyRevoke)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("insufficient permissions to revoke API keys")
	}

	// Revoke key - O(1)
	if err := uc.rbacRepo.RevokeAPIKey(keyID); err != nil {
		uc.logger.Error("Failed to revoke API key", "error", err, "key_id", keyID)
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	uc.logger.Info("API key revoked", "key_id", keyID, "workspace_id", workspaceID)
	return nil
}

// GetWorkspaceAPIKeys retrieves all API keys for a workspace
// Time Complexity: O(k) where k is number of API keys (typically small)
func (uc *RBACUseCase) GetWorkspaceAPIKeys(ctx context.Context, userID, workspaceID string) ([]*domain.APIKey, error) {
	// Check permission
	hasPermission, err := uc.CheckPermission(ctx, userID, workspaceID, domain.PermissionAPIKeyRead)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to view API keys")
	}

	// Get keys - O(k)
	keys, err := uc.rbacRepo.GetWorkspaceAPIKeys(workspaceID)
	if err != nil {
		uc.logger.Error("Failed to get workspace API keys", "error", err, "workspace_id", workspaceID)
		return nil, err
	}

	return keys, nil
}

// GetUserWorkspaces retrieves all workspaces a user belongs to
// Time Complexity: O(w) where w is number of workspaces (typically small)
func (uc *RBACUseCase) GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error) {
	workspaces, err := uc.rbacRepo.GetUserWorkspaces(userID)
	if err != nil {
		uc.logger.Error("Failed to get user workspaces", "error", err, "user_id", userID)
		return nil, err
	}

	return workspaces, nil
}

// GetWorkspace retrieves workspace details
// Time Complexity: O(1) - primary key lookup
func (uc *RBACUseCase) GetWorkspace(ctx context.Context, userID, workspaceID string) (*domain.Workspace, error) {
	// Check if user has access to workspace
	hasPermission, err := uc.CheckPermission(ctx, userID, workspaceID, domain.PermissionWorkspaceRead)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, fmt.Errorf("no access to workspace")
	}

	workspace, err := uc.rbacRepo.GetWorkspace(workspaceID)
	if err != nil {
		uc.logger.Error("Failed to get workspace", "error", err, "workspace_id", workspaceID)
		return nil, err
	}

	return workspace, nil
}

// RemoveUserFromWorkspace removes a user's access to a workspace
// Time Complexity: O(1) - indexed delete
func (uc *RBACUseCase) RemoveUserFromWorkspace(ctx context.Context, adminUserID, targetUserID, workspaceID string) error {
	// Check permission
	hasPermission, err := uc.CheckPermission(ctx, adminUserID, workspaceID, domain.PermissionUserDelete)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("insufficient permissions to remove users")
	}

	// Prevent removing workspace owner
	workspace, err := uc.rbacRepo.GetWorkspace(workspaceID)
	if err != nil {
		return err
	}
	if workspace.OwnerID == targetUserID {
		return fmt.Errorf("cannot remove workspace owner")
	}

	// Remove user role - O(1)
	if err := uc.rbacRepo.DeleteUserRole(targetUserID, workspaceID); err != nil {
		uc.logger.Error("Failed to remove user from workspace", "error", err, "user_id", targetUserID)
		return fmt.Errorf("failed to remove user: %w", err)
	}

	uc.logger.Info("User removed from workspace", "user_id", targetUserID, "workspace_id", workspaceID)
	return nil
}
