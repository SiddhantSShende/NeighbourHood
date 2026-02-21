package domain

import "time"

// Role represents a role in the RBAC system
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleDeveloper Role = "developer"
	RoleUser      Role = "user"
	RoleViewer    Role = "viewer"
)

// Permission represents a specific permission
type Permission string

const (
	// Integration permissions
	PermissionIntegrationRead    Permission = "integration:read"
	PermissionIntegrationWrite   Permission = "integration:write"
	PermissionIntegrationDelete  Permission = "integration:delete"
	PermissionIntegrationExecute Permission = "integration:execute"

	// API Key permissions
	PermissionAPIKeyRead   Permission = "apikey:read"
	PermissionAPIKeyCreate Permission = "apikey:create"
	PermissionAPIKeyRevoke Permission = "apikey:revoke"

	// User management permissions
	PermissionUserRead   Permission = "user:read"
	PermissionUserWrite  Permission = "user:write"
	PermissionUserDelete Permission = "user:delete"

	// Workspace permissions
	PermissionWorkspaceRead   Permission = "workspace:read"
	PermissionWorkspaceWrite  Permission = "workspace:write"
	PermissionWorkspaceDelete Permission = "workspace:delete"
)

// UserRole represents a user's role within a workspace
type UserRole struct {
	ID          string
	UserID      string
	WorkspaceID string
	Role        Role
	Permissions []Permission
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   string
}

// RolePermission defines default permissions for each role
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionIntegrationRead,
		PermissionIntegrationWrite,
		PermissionIntegrationDelete,
		PermissionIntegrationExecute,
		PermissionAPIKeyRead,
		PermissionAPIKeyCreate,
		PermissionAPIKeyRevoke,
		PermissionUserRead,
		PermissionUserWrite,
		PermissionUserDelete,
		PermissionWorkspaceRead,
		PermissionWorkspaceWrite,
		PermissionWorkspaceDelete,
	},
	RoleDeveloper: {
		PermissionIntegrationRead,
		PermissionIntegrationWrite,
		PermissionIntegrationExecute,
		PermissionAPIKeyRead,
		PermissionAPIKeyCreate,
		PermissionUserRead,
		PermissionWorkspaceRead,
	},
	RoleUser: {
		PermissionIntegrationRead,
		PermissionIntegrationExecute,
		PermissionUserRead,
		PermissionWorkspaceRead,
	},
	RoleViewer: {
		PermissionIntegrationRead,
		PermissionUserRead,
		PermissionWorkspaceRead,
	},
}

// HasPermission checks if a role has a specific permission
func (ur *UserRole) HasPermission(permission Permission) bool {
	// Check custom permissions first
	for _, p := range ur.Permissions {
		if p == permission {
			return true
		}
	}

	// Check default role permissions
	defaultPerms, exists := RolePermissions[ur.Role]
	if !exists {
		return false
	}

	for _, p := range defaultPerms {
		if p == permission {
			return true
		}
	}

	return false
}

// Workspace represents a developer workspace
type Workspace struct {
	ID          string
	Name        string
	OwnerID     string
	Description string
	Active      bool
	Plan        string // free, pro, enterprise
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Settings    map[string]interface{}
}

// APIKey represents a unique API key for developers
type APIKey struct {
	ID          string
	WorkspaceID string
	UserID      string
	Name        string
	KeyPrefix   string // nh_live_ or nh_test_
	KeyHash     string // Hashed version of the full key
	LastUsedAt  *time.Time
	ExpiresAt   *time.Time
	Active      bool
	Scopes      []string
	RateLimit   int
	CreatedAt   time.Time
	RevokedAt   *time.Time
}

// RBACRepository defines the interface for RBAC operations
type RBACRepository interface {
	// UserRole operations
	CreateUserRole(userRole *UserRole) error
	GetUserRole(userID, workspaceID string) (*UserRole, error)
	GetUserRoles(userID string) ([]*UserRole, error)
	UpdateUserRole(userRole *UserRole) error
	DeleteUserRole(userID, workspaceID string) error

	// Permission checks (optimized for O(1) average case)
	HasPermission(userID, workspaceID string, permission Permission) (bool, error)

	// Workspace operations
	CreateWorkspace(workspace *Workspace) error
	GetWorkspace(workspaceID string) (*Workspace, error)
	GetUserWorkspaces(userID string) ([]*Workspace, error)
	UpdateWorkspace(workspace *Workspace) error
	DeleteWorkspace(workspaceID string) error

	// API Key operations
	CreateAPIKey(apiKey *APIKey) error
	GetAPIKey(keyPrefix string) (*APIKey, error)
	GetWorkspaceAPIKeys(workspaceID string) ([]*APIKey, error)
	UpdateAPIKey(apiKey *APIKey) error
	RevokeAPIKey(keyID string) error
	ValidateAPIKey(keyHash string) (*APIKey, error)
}
