package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "neighbourhood/proto/gen/go/auth"
	commonpb "neighbourhood/proto/gen/go/common"
	"neighbourhood/services/auth/internal/usecase"
)

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
}

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	useCase *usecase.AuthUseCase
	logger  Logger
}

func NewAuthHandler(useCase *usecase.AuthUseCase, logger Logger) *AuthHandler {
	return &AuthHandler{
		useCase: useCase,
		logger:  logger,
	}
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	h.logger.Info("Register request received", "email", req.Email)

	user, err := h.useCase.Register(ctx, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		h.logger.Error("Registration failed", "error", err, "email", req.Email)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{
		User: &pb.User{
			Id:            user.ID,
			Email:         user.Email,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			AvatarUrl:     user.AvatarURL,
			EmailVerified: user.EmailVerified,
			CreatedAt:     timestamppb.New(user.CreatedAt),
			UpdatedAt:     timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	h.logger.Info("Login request received", "email", req.Email)

	accessToken, refreshToken, err := h.useCase.Login(ctx, req.Email, req.Password, "", "")
	if err != nil {
		h.logger.Error("Login failed", "error", err, "email", req.Email)

		switch err {
		case usecase.ErrInvalidCredentials:
			return nil, status.Error(codes.Unauthenticated, "Invalid email or password")
		case usecase.ErrAccountLocked:
			return nil, status.Error(codes.PermissionDenied, "Account locked due to too many failed login attempts")
		default:
			return nil, status.Error(codes.Internal, "Login failed")
		}
	}

	// Get user profile
	userID, _ := h.useCase.ValidateToken(ctx, accessToken)
	user, err := h.useCase.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get user profile")
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &pb.User{
			Id:            user.ID,
			Email:         user.Email,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			AvatarUrl:     user.AvatarURL,
			EmailVerified: user.EmailVerified,
			CreatedAt:     timestamppb.New(user.CreatedAt),
			UpdatedAt:     timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	userID, err := h.useCase.ValidateToken(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	h.logger.Info("Refresh token request received")

	accessToken, refreshToken, err := h.useCase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		h.logger.Error("Token refresh failed", "error", err)

		switch err {
		case usecase.ErrInvalidToken:
			return nil, status.Error(codes.Unauthenticated, "Invalid refresh token")
		case usecase.ErrTokenExpired:
			return nil, status.Error(codes.Unauthenticated, "Refresh token expired")
		default:
			return nil, status.Error(codes.Internal, "Token refresh failed")
		}
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *AuthHandler) InitiateOAuth(ctx context.Context, req *pb.OAuthRequest) (*pb.OAuthResponse, error) {
	h.logger.Info("OAuth initiation request received", "provider", req.Provider)

	authURL, state, err := h.useCase.InitiateOAuth(ctx, req.Provider.String(), req.RedirectUri)
	if err != nil {
		h.logger.Error("OAuth initiation failed", "error", err, "provider", req.Provider)
		return &pb.OAuthResponse{
			Success: false,
			Error: &commonpb.Error{
				Code:    "OAUTH_INITIATION_FAILED",
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.OAuthResponse{
		Success: true,
		AuthUrl: authURL,
		State:   state,
	}, nil
}

func (h *AuthHandler) CompleteOAuth(ctx context.Context, req *pb.OAuthCallbackRequest) (*pb.OAuthCallbackResponse, error) {
	h.logger.Info("OAuth completion request received", "provider", req.Provider)

	user, accessToken, refreshToken, isNew, err := h.useCase.CompleteOAuth(ctx, req.Provider, req.Code, req.State)
	if err != nil {
		h.logger.Error("OAuth completion failed", "error", err, "provider", req.Provider)
		return &pb.OAuthCallbackResponse{
			Success: false,
			Error: &commonpb.Error{
				Code:    "OAUTH_COMPLETION_FAILED",
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.OAuthCallbackResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IsNewUser:    isNew,
		User: &pb.User{
			Id:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			AvatarUrl: user.AvatarURL,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (h *AuthHandler) GetUserProfile(ctx context.Context, req *pb.GetUserRequest) (*pb.UserProfile, error) {
	h.logger.Info("Get user profile request received", "user_id", req.UserId)

	user, err := h.useCase.GetUserProfile(ctx, req.UserId)
	if err != nil {
		h.logger.Error("Failed to get user profile", "error", err, "user_id", req.UserId)

		if err == usecase.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, "Failed to get user profile")
	}

	return &pb.UserProfile{
		User: &pb.User{
			Id:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			AvatarUrl: user.AvatarURL,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
		IntegrationCount: 0,
		WorkflowCount:    0,
	}, nil
}

func (h *AuthHandler) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserProfile, error) {
	h.logger.Info("Update user profile request received", "user_id", req.UserId)

	user, err := h.useCase.UpdateUserProfile(ctx, req.UserId, req.FirstName, req.LastName, req.AvatarUrl)
	if err != nil {
		h.logger.Error("Failed to update user profile", "error", err, "user_id", req.UserId)

		if err == usecase.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Error(codes.Internal, "Failed to update user profile")
	}

	return &pb.UserProfile{
		User: &pb.User{
			Id:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			AvatarUrl: user.AvatarURL,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
		IntegrationCount: 0,
		WorkflowCount:    0,
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	h.logger.Info("Logout request received", "user_id", req.UserId)

	if err := h.useCase.Logout(ctx, req.Token); err != nil {
		h.logger.Error("Logout failed", "error", err, "user_id", req.UserId)
		return &pb.LogoutResponse{
			Success: false,
			Error: &commonpb.Error{
				Code:    "LOGOUT_FAILED",
				Message: err.Error(),
			},
		}, nil
	}

	return &pb.LogoutResponse{
		Success: true,
	}, nil
}

func (h *AuthHandler) HealthCheck(ctx context.Context, req *commonpb.HealthCheckRequest) (*commonpb.HealthCheckResponse, error) {
	return &commonpb.HealthCheckResponse{
		Status: commonpb.HealthCheckResponse_SERVING,
	}, nil
}
