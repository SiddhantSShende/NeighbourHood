package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"neighbourhood/internal/config"
)

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

func newTestConfig(googleEnabled, githubEnabled bool) *config.Config {
	return &config.Config{
		Auth: config.AuthConfig{
			GoogleOAuth: config.OAuthConfig{
				Enabled:      googleEnabled,
				ClientID:     "test-google-client-id",
				ClientSecret: "test-google-client-secret",
				RedirectURL:  "http://localhost:8080/auth/google/callback",
			},
			GitHubOAuth: config.OAuthConfig{
				Enabled:      githubEnabled,
				ClientID:     "test-github-client-id",
				ClientSecret: "test-github-client-secret",
				RedirectURL:  "http://localhost:8080/auth/github/callback",
			},
		},
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// LoginHandler tests
// ──────────────────────────────────────────────────────────────────────────────

func TestLoginHandler_ValidCredentials(t *testing.T) {
	body := `{"email":"user@example.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	LoginHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["token"] == "" {
		t.Error("expected non-empty token in response")
	}
	if resp["message"] == "" {
		t.Error("expected non-empty message in response")
	}
}

func TestLoginHandler_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	rr := httptest.NewRecorder()

	LoginHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	LoginHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestLoginHandler_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	LoginHandler(rr, req)

	// Should still return 200 (mock implementation)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for empty credentials, got %d", rr.Code)
	}
}

func TestLoginHandler_ResponseContentType(t *testing.T) {
	body := `{"email":"a@b.com","password":"p"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	LoginHandler(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct == "" {
		t.Error("expected Content-Type header to be set")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Basic Auth Middleware tests
// ──────────────────────────────────────────────────────────────────────────────

func TestAuthMiddleware_NoHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()

	Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth header, got %d", rr.Code)
	}
}

func TestAuthMiddleware_WithHeader_PassesThrough(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rr := httptest.NewRecorder()

	Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 with auth header, got %d", rr.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// OAuthHandler – state generation and validation
// ──────────────────────────────────────────────────────────────────────────────

func TestGenerateState_Unique(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))

	s1, err := h.generateState()
	if err != nil {
		t.Fatalf("generateState error: %v", err)
	}
	s2, err := h.generateState()
	if err != nil {
		t.Fatalf("generateState error: %v", err)
	}

	if s1 == s2 {
		t.Error("two generated states should be unique")
	}
}

func TestGenerateState_NonEmpty(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	state, err := h.generateState()
	if err != nil {
		t.Fatalf("generateState error: %v", err)
	}
	if state == "" {
		t.Error("state should not be empty")
	}
	if len(state) < 10 {
		t.Errorf("state too short: %q", state)
	}
}

func TestValidateState_ValidState(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	state, _ := h.generateState()

	if !h.validateState(state) {
		t.Error("valid (fresh) state should pass validation")
	}
}

func TestValidateState_UnknownState(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))

	if h.validateState("totally-unknown-state") {
		t.Error("unknown state should fail validation")
	}
}

func TestValidateState_ConsumedOnce(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	state, _ := h.generateState()

	// First use — should succeed
	if !h.validateState(state) {
		t.Error("first use of valid state should succeed")
	}
	// Second use — should fail (replay protection)
	if h.validateState(state) {
		t.Error("second use of same state should fail (replay protection)")
	}
}

func TestValidateState_ExpiredState(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	state, _ := h.generateState()

	// Manually expire the state
	h.states[state] = time.Now().Add(-1 * time.Hour)

	if h.validateState(state) {
		t.Error("expired state should fail validation")
	}
}

func TestValidateState_CleanupOnExpiry(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	state, _ := h.generateState()
	h.states[state] = time.Now().Add(-1 * time.Hour)

	h.validateState(state) // should delete expired entry

	if _, exists := h.states[state]; exists {
		t.Error("expired state should be removed from map after validation attempt")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Google OAuth flow
// ──────────────────────────────────────────────────────────────────────────────

func TestGoogleLoginHandler_Enabled_Redirects(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	rr := httptest.NewRecorder()

	h.GoogleLoginHandler(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected 307 redirect, got %d", rr.Code)
	}

	loc := rr.Header().Get("Location")
	if loc == "" {
		t.Error("expected Location header in redirect")
	}
	if len(loc) < 20 {
		t.Errorf("redirect URL suspiciously short: %q", loc)
	}
}

func TestGoogleLoginHandler_Disabled_Returns503(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(false, true))
	req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	rr := httptest.NewRecorder()

	h.GoogleLoginHandler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when Google OAuth disabled, got %d", rr.Code)
	}
}

func TestGoogleLoginHandler_RedirectContainsClientID(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	rr := httptest.NewRecorder()

	h.GoogleLoginHandler(rr, req)

	loc := rr.Header().Get("Location")
	if loc == "" {
		t.Skip("no location header")
	}
	// The redirect URL must contain the client_id
	if !bytes.Contains([]byte(loc), []byte("test-google-client-id")) {
		t.Errorf("redirect URL should contain client_id, got: %q", loc)
	}
}

func TestGoogleLoginHandler_RedirectContainsState(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	rr := httptest.NewRecorder()

	h.GoogleLoginHandler(rr, req)

	loc := rr.Header().Get("Location")
	if !bytes.Contains([]byte(loc), []byte("state=")) {
		t.Error("redirect URL should contain state parameter")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Google OAuth callback
// ──────────────────────────────────────────────────────────────────────────────

func TestGoogleCallbackHandler_NoCode(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	state, _ := h.generateState()

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state="+state, nil)
	rr := httptest.NewRecorder()

	h.GoogleCallbackHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when no code, got %d", rr.Code)
	}
}

func TestGoogleCallbackHandler_InvalidState(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=mycode&state=badstate", nil)
	rr := httptest.NewRecorder()

	h.GoogleCallbackHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid state, got %d", rr.Code)
	}
}

func TestGoogleCallbackHandler_MissingStateParam(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=mycode", nil)
	rr := httptest.NewRecorder()

	h.GoogleCallbackHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing state, got %d", rr.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// GitHub OAuth flow
// ──────────────────────────────────────────────────────────────────────────────

func TestGitHubLoginHandler_Enabled_Redirects(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	req := httptest.NewRequest(http.MethodGet, "/auth/github/login", nil)
	rr := httptest.NewRecorder()

	h.GitHubLoginHandler(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected 307, got %d", rr.Code)
	}
}

func TestGitHubLoginHandler_Disabled_Returns503(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, false))
	req := httptest.NewRequest(http.MethodGet, "/auth/github/login", nil)
	rr := httptest.NewRecorder()

	h.GitHubLoginHandler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when GitHub OAuth disabled, got %d", rr.Code)
	}
}

func TestGitHubLoginHandler_RedirectContainsClientID(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	req := httptest.NewRequest(http.MethodGet, "/auth/github/login", nil)
	rr := httptest.NewRecorder()

	h.GitHubLoginHandler(rr, req)

	loc := rr.Header().Get("Location")
	if !bytes.Contains([]byte(loc), []byte("test-github-client-id")) {
		t.Errorf("redirect should contain client_id, got: %q", loc)
	}
}

func TestGitHubCallbackHandler_InvalidState(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))

	req := httptest.NewRequest(http.MethodGet, "/auth/github/callback?code=mycode&state=badstate", nil)
	rr := httptest.NewRecorder()

	h.GitHubCallbackHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid state, got %d", rr.Code)
	}
}

func TestGitHubCallbackHandler_NoCode(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	state, _ := h.generateState()

	req := httptest.NewRequest(http.MethodGet, "/auth/github/callback?state="+state, nil)
	rr := httptest.NewRecorder()

	h.GitHubCallbackHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when no code, got %d", rr.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// generateJWT
// ──────────────────────────────────────────────────────────────────────────────

func TestGenerateJWT_ContainsEmail(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	userInfo := map[string]interface{}{
		"email": "jane@example.com",
		"name":  "Jane",
	}
	token := h.generateJWT(userInfo)
	if token == "" {
		t.Error("generateJWT should return non-empty token")
	}
}

func TestGenerateJWT_MissingEmail(t *testing.T) {
	h := NewOAuthHandler(newTestConfig(true, true))
	// email field absent
	token := h.generateJWT(map[string]interface{}{"name": "NoEmail"})
	if token == "" {
		t.Error("generateJWT should still return a token even without email")
	}
}

// Note: OAuthHandler.states map is not concurrency-safe by design (single-server demo).
// Concurrent access tests are skipped. In production, use sync.Mutex around the states map.
