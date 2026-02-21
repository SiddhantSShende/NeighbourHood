package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"neighbourhood/internal/config"
	// "neighbourhood/internal/database"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual user verification against database
	// user, err := database.GetUserByEmail(req.Email)

	// Mock response for now
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":   "mock-jwt-token",
		"message": "Login successful",
	})
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement actual token validation
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// OAuthHandler manages OAuth authentication
type OAuthHandler struct {
	cfg    *config.Config
	states map[string]time.Time // Store states with expiration (use Redis in production)
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{
		cfg:    cfg,
		states: make(map[string]time.Time),
	}
}

// generateState generates a random state string for CSRF protection
func (h *OAuthHandler) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	h.states[state] = time.Now().Add(10 * time.Minute)
	return state, nil
}

// validateState validates the state parameter
func (h *OAuthHandler) validateState(state string) bool {
	expiry, exists := h.states[state]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		delete(h.states, state)
		return false
	}
	delete(h.states, state)
	return true
}

// GoogleLoginHandler initiates Google OAuth flow
func (h *OAuthHandler) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Auth.GoogleOAuth.Enabled {
		http.Error(w, "Google authentication is disabled", http.StatusServiceUnavailable)
		return
	}

	state, err := h.generateState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		h.cfg.Auth.GoogleOAuth.ClientID,
		h.cfg.Auth.GoogleOAuth.RedirectURL,
		"openid%20profile%20email",
		state,
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GoogleCallbackHandler handles Google OAuth callback
func (h *OAuthHandler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if !h.validateState(state) {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := h.exchangeGoogleCode(code)
	if err != nil {
		http.Error(w, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user info
	userInfo, err := h.getGoogleUserInfo(token)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Create or update user in database
	// user, err := database.CreateOrUpdateGoogleUser(userInfo)

	// Generate JWT token
	jwtToken := h.generateJWT(userInfo)

	// Redirect to frontend with token
	http.Redirect(w, r, fmt.Sprintf("/?token=%s&provider=google", jwtToken), http.StatusTemporaryRedirect)
}

// GitHubLoginHandler initiates GitHub OAuth flow
func (h *OAuthHandler) GitHubLoginHandler(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Auth.GitHubOAuth.Enabled {
		http.Error(w, "GitHub authentication is disabled", http.StatusServiceUnavailable)
		return
	}

	state, err := h.generateState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		h.cfg.Auth.GitHubOAuth.ClientID,
		h.cfg.Auth.GitHubOAuth.RedirectURL,
		"user:email",
		state,
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GitHubCallbackHandler handles GitHub OAuth callback
func (h *OAuthHandler) GitHubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if !h.validateState(state) {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := h.exchangeGitHubCode(code)
	if err != nil {
		http.Error(w, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user info
	userInfo, err := h.getGitHubUserInfo(token)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Create or update user in database
	// user, err := database.CreateOrUpdateGitHubUser(userInfo)

	// Generate JWT token
	jwtToken := h.generateJWT(userInfo)

	// Redirect to frontend with token
	http.Redirect(w, r, fmt.Sprintf("/?token=%s&provider=github", jwtToken), http.StatusTemporaryRedirect)
}

// exchangeGoogleCode exchanges authorization code for access token
func (h *OAuthHandler) exchangeGoogleCode(code string) (string, error) {
	data := fmt.Sprintf(
		"code=%s&client_id=%s&client_secret=%s&redirect_uri=%s&grant_type=authorization_code",
		code,
		h.cfg.Auth.GoogleOAuth.ClientID,
		h.cfg.Auth.GoogleOAuth.ClientSecret,
		h.cfg.Auth.GoogleOAuth.RedirectURL,
	)

	resp, err := http.Post(
		"https://oauth2.googleapis.com/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(data),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("no access token in response")
	}

	return token, nil
}

// getGoogleUserInfo retrieves user information from Google
func (h *OAuthHandler) getGoogleUserInfo(token string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(
		context.Background(),
		"GET",
		"https://www.googleapis.com/oauth2/v2/userinfo",
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

// exchangeGitHubCode exchanges authorization code for access token
func (h *OAuthHandler) exchangeGitHubCode(code string) (string, error) {
	data := fmt.Sprintf(
		"code=%s&client_id=%s&client_secret=%s&redirect_uri=%s",
		code,
		h.cfg.Auth.GitHubOAuth.ClientID,
		h.cfg.Auth.GitHubOAuth.ClientSecret,
		h.cfg.Auth.GitHubOAuth.RedirectURL,
	)

	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		"https://github.com/login/oauth/access_token",
		strings.NewReader(data),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("no access token in response")
	}

	return token, nil
}

// getGitHubUserInfo retrieves user information from GitHub
func (h *OAuthHandler) getGitHubUserInfo(token string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(
		context.Background(),
		"GET",
		"https://api.github.com/user",
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

// generateJWT generates a JWT token for the user (simplified version)
func (h *OAuthHandler) generateJWT(userInfo map[string]interface{}) string {
	// TODO: Implement proper JWT generation with signing
	// For now, return a mock token
	email, _ := userInfo["email"].(string)
	return fmt.Sprintf("jwt-token-for-%s", email)
}
