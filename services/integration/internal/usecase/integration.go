package usecase

import (
	"context"
	"fmt"

	"neighbourhood/services/integration/internal/domain"
	"neighbourhood/services/integration/pkg/providers"
)

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type IntegrationUseCase struct {
	repo     domain.ProviderRepository
	registry *providers.Registry
	logger   Logger
}

func NewIntegrationUseCase(repo domain.ProviderRepository, registry *providers.Registry, logger Logger) *IntegrationUseCase {
	return &IntegrationUseCase{
		repo:     repo,
		registry: registry,
		logger:   logger,
	}
}

func (uc *IntegrationUseCase) ListProviders(ctx context.Context, category string) ([]*providers.Provider, error) {
	// Get all providers from registry
	providersList := make([]*providers.Provider, 0)
	
	for _, p := range uc.registry.GetAll() {
		providersList = append(providersList, &providers.Provider{
			Type:             p.ID(),
			Name:             p.Name(),
			Category:         p.Category(),
			Enabled:          true,
			SupportedActions: []string{},
			RequiredScopes:   []string{},
		})
	}
	
	return providersList, nil
}

func (uc *IntegrationUseCase) GetProvider(ctx context.Context, providerType string) (*providers.Provider, error) {
	p, err := uc.registry.Get(providerType)
	if err != nil {
		return nil, err
	}
	
	return &providers.Provider{
		Type:     p.ID(),
		Name:     p.Name(),
		Category: p.Category(),
		Enabled:  true,
	}, nil
}

func (uc *IntegrationUseCase) GetAuthURL(ctx context.Context, providerType, userID, redirectURI string, scopes []string) (string, string, error) {
	provider, err := uc.registry.Get(providerType)
	if err != nil {
		return "", "", fmt.Errorf("provider not found: %w", err)
	}

	// Generate state for CSRF protection
	state := fmt.Sprintf("%s:%s", providerType, userID)
	return provider.GetAuthURL(state), state, nil
}

func (uc *IntegrationUseCase) ExchangeCode(ctx context.Context, providerType, userID, code, state string) (string, *providers.Token, error) {
	provider, err := uc.registry.Get(providerType)
	if err != nil {
		return "", nil, fmt.Errorf("provider not found: %w", err)
	}

	token, err := provider.ExchangeCode(ctx, code)
	if err != nil {
		uc.logger.Error("Failed to exchange code", "provider", providerType, "error", err)
		return "", nil, err
	}

	// Generate integration ID
	integrationID := fmt.Sprintf("int_%s_%s", providerType, userID)
	
	uc.logger.Info("Code exchanged successfully", "provider", providerType, "user", userID)
	return integrationID, token, nil
}

func (uc *IntegrationUseCase) ExecuteAction(ctx context.Context, integrationID, userID, action string, params map[string]interface{}) (map[string]interface{}, error) {
	// TODO: Extract provider type from integration ID
	// For now, we'll parse from integration ID format: "int_<provider>_<user>"
	providerType := "slack" // Placeholder
	
	provider, err := uc.registry.Get(providerType)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// TODO: Retrieve token from repository using integrationID
	token := &providers.Token{} // Placeholder
	
	result, err := provider.Execute(ctx, token, action, params)
	if err != nil {
		uc.logger.Error("Failed to execute action", "integration", integrationID, "action", action, "error", err)
		return nil, err
	}

	uc.logger.Info("Action executed successfully", "integration", integrationID, "action", action)
	
	// Convert result to map[string]interface{}
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap, nil
	}
	return map[string]interface{}{"result": result}, nil
}

func (uc *IntegrationUseCase) GetUserIntegrations(ctx context.Context, userID, category string) ([]*providers.UserIntegration, error) {
	// TODO: Implement fetching user integrations from repository
	return []*providers.UserIntegration{}, nil
}
