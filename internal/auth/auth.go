package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"neighbourhood/internal/config"

	"github.com/golang-jwt/jwt/v5"
	// "neighbourhood/internal/database"
)

// httpClient is a shared HTTP client with a sensible timeout for all outbound
// OAuth calls. The default http.DefaultClient has no timeout and must not be
// used for requests to third-party servers in production.
var httpClient = &http.Client{Timeout: 15 * time.Second}

// LoginRequest is the JSON body expected by LoginHandler.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginHandler handles email/password login requests.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Verify credentials against database.
	// user, err := database.GetUserByEmail(req.Email)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"token":   "mock-jwt-token",
		"message": "Login successful",
	}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// Middleware is a lightweight auth guard used directly on routes.
// Prefer the richer middleware.Auth middleware when building the main mux.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// stateEntry holds a generated OAuth state value and its expiry time.
type stateEntry struct {
	expiresAt time.Time
}

// OAuthHandler manages OAuth authentication flows.
type OAuthHandler struct {
	cfg    *config.Config
	mu     sync.RWMutex          // guards states
	states map[string]stateEntry // CSRF state tokens
}

// NewOAuthHandler creates a new OAuthHandler.
func NewOAuthHandler(cfg *config.Config) *OAuthHandler {
	h := &OAuthHandler{
		cfg:    cfg,
		states: make(map[string]stateEntry),
	}
	go h.cleanupStates() // background goroutine to evict expired state entries
	return h
}

// cleanupStates removes expired state entries every 5 minutes to prevent
// unbounded memory growth when the callback is never called.
func (h *OAuthHandler) cleanupStates() {
	const interval = 5 * time.Minute
	t := time.NewTicker(interval)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		h.mu.Lock()
		for state, entry := range h.states {
			if now.After(entry.expiresAt) {
				delete(h.states, state)
			}
		}
		h.mu.Unlock()
	}
}

// generateState creates a cryptographically secure random state string for
// CSRF protection and stores it with an expiry.
func (h *OAuthHandler) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(b)

	h.mu.Lock()
	h.states[state] = stateEntry{expiresAt: time.Now().Add(10 * time.Minute)}
	h.mu.Unlock()

	return state, nil
}

// validateState returns true and deletes the state if it exists and has not expired.
func (h *OAuthHandler) validateState(state string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	entry, exists := h.states[state]
	if !exists {
		return false
	}
	delete(h.states, state) // consume once — replay protection

	return time.Now().Before(entry.expiresAt)
}

// GoogleLoginHandler initiates the Google OAuth flow.
func (h *OAuthHandler) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Auth.GoogleOAuth.Enabled {
		http.Error(w, "Google authentication is disabled", http.StatusServiceUnavailable)
		return
	}

	state, err := h.generateState()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}

	params := url.Values{}
	params.Set("client_id", h.cfg.Auth.GoogleOAuth.ClientID)
	params.Set("redirect_uri", h.cfg.Auth.GoogleOAuth.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "openid profile email")
	params.Set("state", state)
	authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GoogleCallbackHandler handles the Google OAuth callback.
func (h *OAuthHandler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if !h.validateState(state) {
		http.Error(w, "invalid or expired state parameter", http.StatusBadRequest)
		return
	}
	if code == "" {
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}

	token, err := h.exchangeGoogleCode(r.Context(), code)
	if err != nil {
		http.Error(w, "failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := h.getGoogleUserInfo(r.Context(), token)
	if err != nil {
		http.Error(w, "failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Create or update user in database.
	// user, err := database.CreateOrUpdateGoogleUser(userInfo)

	jwtToken, err := h.generateJWT(userInfo)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("/?token=%s&provider=google", url.QueryEscape(jwtToken))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// GitHubLoginHandler initiates the GitHub OAuth flow.
func (h *OAuthHandler) GitHubLoginHandler(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Auth.GitHubOAuth.Enabled {
		http.Error(w, "GitHub authentication is disabled", http.StatusServiceUnavailable)
		return
	}

	state, err := h.generateState()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}

	params := url.Values{}
	params.Set("client_id", h.cfg.Auth.GitHubOAuth.ClientID)
	params.Set("redirect_uri", h.cfg.Auth.GitHubOAuth.RedirectURL)
	params.Set("scope", "user:email")
	params.Set("state", state)
	authURL := "https://github.com/login/oauth/authorize?" + params.Encode()

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GitHubCallbackHandler handles the GitHub OAuth callback.
func (h *OAuthHandler) GitHubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if !h.validateState(state) {
		http.Error(w, "invalid or expired state parameter", http.StatusBadRequest)
		return
	}
	if code == "" {
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}

	token, err := h.exchangeGitHubCode(r.Context(), code)
	if err != nil {
		http.Error(w, "failed to exchange code: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := h.getGitHubUserInfo(r.Context(), token)
	if err != nil {
		http.Error(w, "failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Create or update user in database.
	// user, err := database.CreateOrUpdateGitHubUser(userInfo)

	jwtToken, err := h.generateJWT(userInfo)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("/?token=%s&provider=github", url.QueryEscape(jwtToken))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// exchangeGoogleCode exchanges an authorization code for a Google access token.
func (h *OAuthHandler) exchangeGoogleCode(ctx context.Context, code string) (string, error) {
	formData := url.Values{}
	formData.Set("code", code)
	formData.Set("client_id", h.cfg.Auth.GoogleOAuth.ClientID)
	formData.Set("client_secret", h.cfg.Auth.GoogleOAuth.ClientSecret)
	formData.Set("redirect_uri", h.cfg.Auth.GoogleOAuth.RedirectURL)
	formData.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://oauth2.googleapis.com/token",
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("building token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("exchanging code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding token response: %w", err)
	}

	token, ok := result["access_token"].(string)
	if !ok || token == "" {
		return "", fmt.Errorf("no access_token in response")
	}

	return token, nil
}

// getGoogleUserInfo fetches user profile information from Google.
func (h *OAuthHandler) getGoogleUserInfo(ctx context.Context, token string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://www.googleapis.com/oauth2/v2/userinfo",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("building userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned %d", resp.StatusCode)
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("decoding user info: %w", err)
	}

	return userInfo, nil
}

// exchangeGitHubCode exchanges an authorization code for a GitHub access token.
func (h *OAuthHandler) exchangeGitHubCode(ctx context.Context, code string) (string, error) {
	formData := url.Values{}
	formData.Set("code", code)
	formData.Set("client_id", h.cfg.Auth.GitHubOAuth.ClientID)
	formData.Set("client_secret", h.cfg.Auth.GitHubOAuth.ClientSecret)
	formData.Set("redirect_uri", h.cfg.Auth.GitHubOAuth.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("building token request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("exchanging code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding token response: %w", err)
	}

	if errMsg, ok := result["error"].(string); ok && errMsg != "" {
		return "", fmt.Errorf("GitHub token error: %s", errMsg)
	}

	token, ok := result["access_token"].(string)
	if !ok || token == "" {
		return "", fmt.Errorf("no access_token in response")
	}

	return token, nil
}

// getGitHubUserInfo fetches user profile information from GitHub.
func (h *OAuthHandler) getGitHubUserInfo(ctx context.Context, token string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/user",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("building userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user API returned %d", resp.StatusCode)
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("decoding user info: %w", err)
	}

	return userInfo, nil
}

// generateJWT creates a signed HS256 JWT for the authenticated user.
// The token contains standard claims (sub, email, iat, exp) signed with the
// configured JWT secret.
func (h *OAuthHandler) generateJWT(userInfo map[string]interface{}) (string, error) {
	email, _ := userInfo["email"].(string)
	name, _ := userInfo["name"].(string)
	sub, _ := userInfo["id"].(string)
	if sub == "" {
		// GitHub uses a numeric "id" field; fall back to email.
		if id, ok := userInfo["id"].(float64); ok {
			sub = fmt.Sprintf("%.0f", id)
		} else {
			sub = email
		}
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   sub,
		"email": email,
		"name":  name,
		"iat":   now.Unix(),
		"exp":   now.Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.cfg.Auth.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("signing JWT: %w", err)
	}

	return signed, nil
}
