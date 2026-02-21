package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "neighbourhood/proto/gen/go/common"
	pb "neighbourhood/proto/gen/go/integration"
	"neighbourhood/services/integration/pkg/providers"
)

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type IntegrationService interface {
	ListProviders(ctx context.Context, category string) ([]*providers.Provider, error)
	GetProvider(ctx context.Context, providerType string) (*providers.Provider, error)
	GetAuthURL(ctx context.Context, providerType, userID, redirectURI string, scopes []string) (string, string, error)
	ExchangeCode(ctx context.Context, providerType, userID, code, state string) (string, *providers.Token, error)
	ExecuteAction(ctx context.Context, integrationID, userID, action string, payload map[string]interface{}) (map[string]interface{}, error)
	GetUserIntegrations(ctx context.Context, userID, category string) ([]*providers.UserIntegration, error)
}

type IntegrationHandler struct {
	pb.UnimplementedIntegrationServiceServer
	service IntegrationService
	logger  Logger
}

func NewIntegrationHandler(service IntegrationService, logger Logger) *IntegrationHandler {
	return &IntegrationHandler{
		service: service,
		logger:  logger,
	}
}

// ListProviders returns all available integration providers
func (h *IntegrationHandler) ListProviders(ctx context.Context, req *pb.ListProvidersRequest) (*pb.ListProvidersResponse, error) {
	providersList, err := h.service.ListProviders(ctx, req.Category)
	if err != nil {
		h.logger.Error("Failed to list providers", "error", err)
		return nil, status.Error(codes.Internal, "Failed to list providers")
	}

	var pbProviders []*pb.Provider
	for _, p := range providersList {
		pbProviders = append(pbProviders, &pb.Provider{
			Type:             p.Type,
			Name:             p.Name,
			Description:      p.Description,
			Category:         p.Category,
			IconUrl:          p.IconURL,
			Enabled:          p.Enabled,
			SupportedActions: p.SupportedActions,
			RequiredScopes:   p.RequiredScopes,
			CreatedAt:        timestamppb.Now(),
		})
	}

	return &pb.ListProvidersResponse{
		Providers:      pbProviders,
		TotalAvailable: int32(len(pbProviders)),
	}, nil
}

// GetAuthURL generates OAuth authorization URL for a provider
func (h *IntegrationHandler) GetAuthURL(ctx context.Context, req *pb.GetAuthURLRequest) (*pb.GetAuthURLResponse, error) {
	authURL, state, err := h.service.GetAuthURL(ctx, req.ProviderType, req.UserId, req.RedirectUri, req.Scopes)
	if err != nil {
		h.logger.Error("Failed to generate auth URL", "provider", req.ProviderType, "error", err)
		return &pb.GetAuthURLResponse{
			Success: false,
			Error: &commonpb.Error{
				Code:    "AUTH_URL_ERROR",
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.GetAuthURLResponse{
		Success: true,
		AuthUrl: authURL,
		State:   state,
	}, nil
}

// ExchangeCode exchanges authorization code for access token
func (h *IntegrationHandler) ExchangeCode(ctx context.Context, req *pb.ExchangeCodeRequest) (*pb.ExchangeCodeResponse, error) {
	integrationID, token, err := h.service.ExchangeCode(ctx, req.ProviderType, req.UserId, req.Code, req.State)
	if err != nil {
		h.logger.Error("Code exchange failed", "provider", req.ProviderType, "error", err)
		return &pb.ExchangeCodeResponse{
			Success: false,
			Error: &commonpb.Error{
				Code:    "EXCHANGE_ERROR",
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.ExchangeCodeResponse{
		Success:       true,
		IntegrationId: integrationID,
		Token: &pb.Token{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			TokenType:    token.TokenType,
			ExpiresAt:    timestamppb.New(token.ExpiresAt),
			Scopes:       token.Scopes,
		},
	}, nil
}

// ExecuteAction executes an action on an integration
func (h *IntegrationHandler) ExecuteAction(ctx context.Context, req *pb.ExecuteActionRequest) (*pb.ExecuteActionResponse, error) {
	_, err := h.service.ExecuteAction(ctx, req.IntegrationId, req.UserId, req.Action, req.Payload.AsMap())
	if err != nil {
		h.logger.Error("Action execution failed", "integration", req.IntegrationId, "action", req.Action, "error", err)
		return &pb.ExecuteActionResponse{
			Success: false,
			Error: &commonpb.Error{
				Code:    "EXECUTION_ERROR",
				Message: err.Error(),
			},
		}, nil
	}

	h.logger.Info("Action executed successfully", "integration", req.IntegrationId, "action", req.Action)
	return &pb.ExecuteActionResponse{
		Success:    true,
		ExecutedAt: timestamppb.Now(),
	}, nil
}

// GetUserIntegrations returns all integrations for a user
func (h *IntegrationHandler) GetUserIntegrations(ctx context.Context, req *pb.GetUserIntegrationsRequest) (*pb.GetUserIntegrationsResponse, error) {
	integrations, err := h.service.GetUserIntegrations(ctx, req.UserId, req.Category)
	if err != nil {
		h.logger.Error("Failed to get user integrations", "user", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to get user integrations")
	}

	var pbIntegrations []*pb.UserIntegration
	for _, integration := range integrations {
		pbIntegrations = append(pbIntegrations, &pb.UserIntegration{
			Id:           integration.ID,
			UserId:       integration.UserID,
			ProviderType: integration.ProviderType,
			Status:       integration.Status,
			ConnectedAt:  timestamppb.New(integration.ConnectedAt),
			LastUsed:     timestamppb.New(integration.LastUsed),
		})
	}

	return &pb.GetUserIntegrationsResponse{
		Integrations: pbIntegrations,
	}, nil
}

// HealthCheck returns service health status
func (h *IntegrationHandler) HealthCheck(ctx context.Context, req *commonpb.HealthCheckRequest) (*commonpb.HealthCheckResponse, error) {
	return &commonpb.HealthCheckResponse{
		Status: commonpb.HealthCheckResponse_SERVING,
	}, nil
}
