package domain

import "time"

// Provider represents a third-party integration provider
type Provider struct {
	ID          string
	Name        string
	Category    string
	Description string
	AuthType    string
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserIntegration represents a user's connected integration
type UserIntegration struct {
	ID           string
	UserID       string
	ProviderID   string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Scopes       []string
	Metadata     map[string]interface{}
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ActionResult represents the result of an integration action
type ActionResult struct {
	Success   bool
	Data      map[string]interface{}
	Error     string
	Timestamp time.Time
}

// ProviderRepository manages provider data
type ProviderRepository interface {
	GetAll() ([]*Provider, error)
	GetByID(id string) (*Provider, error)
	GetByCategory(category string) ([]*Provider, error)
}

// UserIntegrationRepository manages user integration connections
type UserIntegrationRepository interface {
	Create(integration *UserIntegration) error
	GetByID(id string) (*UserIntegration, error)
	GetByUserID(userID string) ([]*UserIntegration, error)
	GetByUserAndProvider(userID, providerID string) (*UserIntegration, error)
	Update(integration *UserIntegration) error
	Delete(id string) error
}
