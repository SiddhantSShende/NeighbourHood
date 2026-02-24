package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"

	"neighbourhood/services/auth/internal/config"
	"neighbourhood/services/auth/internal/domain"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrAccountLocked      = errors.New("account locked due to too many failed login attempts")
)

// Logger interface for dependency injection
type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
}

type AuthUseCase struct {
	userRepo         domain.UserRepository
	oauthRepo        domain.OAuthRepository
	sessionRepo      domain.SessionRepository
	loginAttemptRepo domain.LoginAttemptRepository
	jwtConfig        config.JWTConfig
	oauthConfig      config.OAuthConfig
	securityConfig   config.SecurityConfig
	logger           Logger
	oauthConfigs     map[string]*oauth2.Config
}

func NewAuthUseCase(
	pgRepo interface {
		domain.UserRepository
		domain.OAuthRepository
	},
	redisRepo interface {
		domain.SessionRepository
		domain.LoginAttemptRepository
	},
	jwtConfig config.JWTConfig,
	oauthConfig config.OAuthConfig,
	securityConfig config.SecurityConfig,
	logger Logger,
) *AuthUseCase {
	uc := &AuthUseCase{
		userRepo:         pgRepo,
		oauthRepo:        pgRepo,
		sessionRepo:      redisRepo,
		loginAttemptRepo: redisRepo,
		jwtConfig:        jwtConfig,
		oauthConfig:      oauthConfig,
		securityConfig:   securityConfig,
		logger:           logger,
		oauthConfigs:     make(map[string]*oauth2.Config),
	}

	// Initialize OAuth configs
	uc.initOAuthConfigs()

	return uc
}

func (uc *AuthUseCase) initOAuthConfigs() {
	for provider, cfg := range uc.oauthConfig.Providers {
		if !cfg.Enabled {
			continue
		}

		var endpoint oauth2.Endpoint
		switch provider {
		case "google":
			endpoint = google.Endpoint
		case "github":
			endpoint = github.Endpoint
		case "microsoft":
			endpoint = microsoft.AzureADEndpoint("")
		default:
			continue
		}

		uc.oauthConfigs[provider] = &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
			Endpoint:     endpoint,
		}
	}
}

// Register creates a new user account
func (uc *AuthUseCase) Register(ctx context.Context, email, password, firstName, lastName string) (*domain.User, error) {
	// Check if user exists
	existing, err := uc.userRepo.GetByEmail(email)
	if err == nil && existing != nil {
		return nil, ErrUserExists
	}

	// Validate password
	if err := uc.validatePassword(password); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), uc.securityConfig.BCryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	uc.logger.Info("User registered", "user_id", user.ID, "email", email)

	return user, nil
}

// Login authenticates a user and creates a session
func (uc *AuthUseCase) Login(ctx context.Context, email, password, userAgent, ipAddress string) (string, string, error) {
	// Check if account is locked
	locked, err := uc.loginAttemptRepo.IsLocked(email)
	if err != nil {
		return "", "", err
	}
	if locked {
		return "", "", ErrAccountLocked
	}

	// Get user
	user, err := uc.userRepo.GetByEmail(email)
	if err != nil {
		if recordErr := uc.loginAttemptRepo.Record(email); recordErr != nil {
			uc.logger.Error("Failed to record login attempt", "error", recordErr)
		}
		return "", "", ErrInvalidCredentials
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if recordErr := uc.loginAttemptRepo.Record(email); recordErr != nil {
			uc.logger.Error("Failed to record login attempt", "error", recordErr)
		}
		return "", "", ErrInvalidCredentials
	}

	// Reset login attempts
	if resetErr := uc.loginAttemptRepo.Reset(email); resetErr != nil {
		uc.logger.Error("Failed to reset login attempts", "error", resetErr)
	}

	// Generate tokens
	accessToken, refreshToken, err := uc.generateTokens(user.ID)
	if err != nil {
		return "", "", err
	}

	// Create session
	session := &domain.Session{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(uc.jwtConfig.RefreshTokenExpiry),
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
	}

	if err := uc.sessionRepo.Create(session); err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}

	uc.logger.Info("User logged in", "user_id", user.ID, "email", email)

	return accessToken, refreshToken, nil
}

// ValidateToken validates an access token and returns the user ID
func (uc *AuthUseCase) ValidateToken(ctx context.Context, tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(uc.jwtConfig.Secret), nil
	})

	if err != nil {
		return "", ErrInvalidToken
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return userID, nil
}

// RefreshToken generates a new access token using a refresh token
func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// Get session by refresh token
	session, err := uc.sessionRepo.GetByRefreshToken(refreshToken)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	// Check if session expired
	if time.Now().After(session.ExpiresAt) {
		uc.sessionRepo.Delete(session.ID)
		return "", "", ErrTokenExpired
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := uc.generateTokens(session.UserID)
	if err != nil {
		return "", "", err
	}

	// Update session
	session.AccessToken = accessToken
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(uc.jwtConfig.RefreshTokenExpiry)

	// Delete old session and create new one
	if delErr := uc.sessionRepo.Delete(session.ID); delErr != nil {
		uc.logger.Error("Failed to delete old session", "error", delErr, "session_id", session.ID)
	}
	session.ID = uuid.New().String()
	session.CreatedAt = time.Now()

	if err := uc.sessionRepo.Create(session); err != nil {
		return "", "", fmt.Errorf("failed to update session: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// Logout invalidates a user's session
func (uc *AuthUseCase) Logout(ctx context.Context, accessToken string) error {
	session, err := uc.sessionRepo.GetByAccessToken(accessToken)
	if err != nil {
		return err
	}

	if err := uc.sessionRepo.Delete(session.ID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	uc.logger.Info("User logged out", "user_id", session.UserID)

	return nil
}

// InitiateOAuth starts the OAuth flow and returns the authorization URL
func (uc *AuthUseCase) InitiateOAuth(ctx context.Context, provider, redirectUri string) (string, string, error) {
	oauthConfig, exists := uc.oauthConfigs[provider]
	if !exists {
		return "", "", fmt.Errorf("provider %s not configured", provider)
	}

	// Generate random state for CSRF protection
	state := uuid.New().String()

	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return url, state, nil
}

// CompleteOAuth completes the OAuth flow and creates or links a user account
func (uc *AuthUseCase) CompleteOAuth(ctx context.Context, provider, code, state string) (*domain.User, string, string, bool, error) {
	oauthConfig, exists := uc.oauthConfigs[provider]
	if !exists {
		return nil, "", "", false, fmt.Errorf("provider %s not configured", provider)
	}

	// Exchange code for token
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", "", false, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from provider (simplified - would need provider-specific implementation)
	providerID := uuid.New().String()             // Placeholder
	email := fmt.Sprintf("user@%s.com", provider) // Placeholder

	// Check if OAuth account exists
	oauthAccount, err := uc.oauthRepo.GetByProviderAndID(provider, providerID)

	var user *domain.User
	isNewUser := false

	if err != nil {
		// OAuth account doesn't exist, check if user exists by email
		user, err = uc.userRepo.GetByEmail(email)
		if err != nil {
			// Create new user
			isNewUser = true
			user = &domain.User{
				ID:        uuid.New().String(),
				Email:     email,
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := uc.userRepo.Create(user); err != nil {
				return nil, "", "", false, fmt.Errorf("failed to create user: %w", err)
			}
		}

		// Link OAuth account
		oauthAccount = &domain.OAuthAccount{
			ID:           uuid.New().String(),
			UserID:       user.ID,
			Provider:     provider,
			ProviderID:   providerID,
			Email:        email,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    token.Expiry,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := uc.oauthRepo.CreateOAuth(oauthAccount); err != nil {
			return nil, "", "", false, fmt.Errorf("failed to create OAuth account: %w", err)
		}
	} else {
		// OAuth account exists, get user
		user, err = uc.userRepo.GetByID(oauthAccount.UserID)
		if err != nil {
			return nil, "", "", false, ErrUserNotFound
		}

		// Update OAuth token
		oauthAccount.AccessToken = token.AccessToken
		oauthAccount.RefreshToken = token.RefreshToken
		oauthAccount.ExpiresAt = token.Expiry
		oauthAccount.UpdatedAt = time.Now()
		if updateErr := uc.oauthRepo.UpdateOAuth(oauthAccount); updateErr != nil {
			uc.logger.Error("Failed to update OAuth token", "error", updateErr, "provider", provider)
		}
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := uc.generateTokens(user.ID)
	if err != nil {
		return nil, "", "", false, err
	}

	// Create session (using state as a basic user agent placeholder)
	session := &domain.Session{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(uc.jwtConfig.RefreshTokenExpiry),
		UserAgent:    state, // Using state temporarily
		IPAddress:    "",
		CreatedAt:    time.Now(),
	}

	if err := uc.sessionRepo.Create(session); err != nil {
		return nil, "", "", false, fmt.Errorf("failed to create session: %w", err)
	}

	uc.logger.Info("OAuth login successful", "user_id", user.ID, "provider", provider, "is_new", isNewUser)

	return user, accessToken, refreshToken, isNewUser, nil
}

// GetUserProfile retrieves a user's profile
func (uc *AuthUseCase) GetUserProfile(ctx context.Context, userID string) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// UpdateUserProfile updates a user's profile
func (uc *AuthUseCase) UpdateUserProfile(ctx context.Context, userID, firstName, lastName, avatarURL string) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.AvatarURL = avatarURL
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// Helper functions

func (uc *AuthUseCase) generateTokens(userID string) (string, string, error) {
	// Generate access token
	accessClaims := jwt.MapClaims{
		"sub":  userID,
		"iss":  uc.jwtConfig.Issuer,
		"aud":  uc.jwtConfig.Audience,
		"exp":  time.Now().Add(uc.jwtConfig.AccessTokenExpiry).Unix(),
		"iat":  time.Now().Unix(),
		"type": "access",
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(uc.jwtConfig.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := jwt.MapClaims{
		"sub":  userID,
		"iss":  uc.jwtConfig.Issuer,
		"aud":  uc.jwtConfig.Audience,
		"exp":  time.Now().Add(uc.jwtConfig.RefreshTokenExpiry).Unix(),
		"iat":  time.Now().Unix(),
		"type": "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(uc.jwtConfig.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, nil
}

func (uc *AuthUseCase) validatePassword(password string) error {
	if len(password) < uc.securityConfig.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters", uc.securityConfig.PasswordMinLength)
	}

	// Add more validation based on security config
	// This is a simplified version

	return nil
}
