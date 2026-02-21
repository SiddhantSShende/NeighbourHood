package providers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"neighbourhood/services/integration/internal/config"
)

// Provider defines the interface for all integration providers
type Provider struct {
	Type             string
	Name             string
	Description      string
	Category         string
	IconURL          string
	Enabled          bool
	SupportedActions []string
	RequiredScopes   []string
}

// ProviderInterface defines the methods each provider implementation must have
type ProviderInterface interface {
	ID() string
	Name() string
	Category() string
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*Token, error)
	Execute(ctx context.Context, token *Token, action string, params map[string]interface{}) (interface{}, error)
}

// Token represents an OAuth access token
type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string
	Scopes       []string
}

// UserIntegration represents a user's connected integration
type UserIntegration struct {
	ID           string
	UserID       string
	ProviderType string
	Status       string
	ConnectedAt  time.Time
	LastUsed     time.Time
}

// Registry manages all registered providers
type Registry struct {
	providers map[string]ProviderInterface
	mu        sync.RWMutex
	logger    Logger
}

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
}

// NewRegistry creates a new provider registry
func NewRegistry(configs map[string]config.ProviderConfig, logger Logger) *Registry {
	r := &Registry{
		providers: make(map[string]ProviderInterface),
		logger:    logger,
	}

	// Register enabled providers
	for name, cfg := range configs {
		if !cfg.Enabled {
			continue
		}

		provider := createProvider(name, cfg)
		if provider != nil {
			r.Register(provider)
			logger.Info("Provider registered", "name", name, "category", provider.Category())
		}
	}

	return r
}

// Register adds a provider to the registry
func (r *Registry) Register(p ProviderInterface) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.ID()] = p
}

// Get retrieves a provider by ID
func (r *Registry) Get(id string) (ProviderInterface, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[id]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", id)
	}

	return provider, nil
}

// GetAll returns all registered providers
func (r *Registry) GetAll() []ProviderInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]ProviderInterface, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}

	return providers
}

// Count returns the number of registered providers
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.providers)
}

// createProvider creates a provider instance based on configuration
func createProvider(name string, cfg config.ProviderConfig) ProviderInterface {
	switch name {
	case "slack":
		return NewSlackProvider(cfg)
	case "gmail":
		return NewGmailProvider(cfg)
	case "jira":
		return NewJiraProvider(cfg)
	case "github":
		return NewGitHubProvider(cfg)
	// Add more providers as needed
	default:
		return NewGenericProvider(name, cfg)
	}
}
