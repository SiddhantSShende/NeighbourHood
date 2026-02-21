package consent

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ConsentStatus represents the status of a consent
type ConsentStatus string

const (
	ConsentGranted ConsentStatus = "granted"
	ConsentRevoked ConsentStatus = "revoked"
	ConsentPending ConsentStatus = "pending"
)

// Consent represents a user's consent for data sharing with a third-party
type Consent struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	UserID    uuid.UUID     `json:"user_id" db:"user_id"`
	Provider  string        `json:"provider" db:"provider"`
	Purpose   string        `json:"purpose" db:"purpose"`
	Status    ConsentStatus `json:"status" db:"status"`
	GrantedAt *time.Time    `json:"granted_at,omitempty" db:"granted_at"`
	RevokedAt *time.Time    `json:"revoked_at,omitempty" db:"revoked_at"`
	ExpiresAt *time.Time    `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" db:"updated_at"`
}

// Manager handles consent operations
type Manager struct {
	// In production, add database connection here
}

// NewManager creates a new consent manager
func NewManager() *Manager {
	return &Manager{}
}

// Grant grants consent for a user to share data with a provider
func (m *Manager) Grant(ctx context.Context, userID uuid.UUID, provider, purpose string) (*Consent, error) {
	now := time.Now()
	consent := &Consent{
		ID:        uuid.New(),
		UserID:    userID,
		Provider:  provider,
		Purpose:   purpose,
		Status:    ConsentGranted,
		GrantedAt: &now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// TODO: Store in database
	return consent, nil
}

// Revoke revokes a user's consent
func (m *Manager) Revoke(ctx context.Context, consentID uuid.UUID) error {
	// TODO: Update database
	return nil
}

// Check checks if consent is valid for a user and provider
func (m *Manager) Check(ctx context.Context, userID uuid.UUID, provider string) (bool, error) {
	// TODO: Query database for active consent
	// For now, return true as mock
	return true, nil
}

// List lists all consents for a user
func (m *Manager) List(ctx context.Context, userID uuid.UUID) ([]Consent, error) {
	// TODO: Query database
	return []Consent{}, nil
}

// IntegrationConsentRequired checks if an integration requires consent before execution
func IntegrationConsentRequired(provider string) bool {
	// Define which integrations require explicit consent
	consentRequired := map[string]bool{
		"slack": true,
		"gmail": true,
		"jira":  true,
	}
	return consentRequired[provider]
}

// ValidateConsent validates that a user has granted consent for an integration
func (m *Manager) ValidateConsent(ctx context.Context, userID uuid.UUID, provider string) error {
	if !IntegrationConsentRequired(provider) {
		return nil
	}

	valid, err := m.Check(ctx, userID, provider)
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("consent not granted for this provider")
	}

	return nil
}

// IntegrateFriendConsentSystem integrates with external consent management system
// This is where you'd integrate with your friend's consent management system
func (m *Manager) IntegrateFriendConsentSystem(apiURL, apiKey string) error {
	// TODO: Implement HTTP client to your friend's consent API
	// Example:
	// - Sync consent data
	// - Webhook for consent changes
	// - Real-time consent validation
	return nil
}
