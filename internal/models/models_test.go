package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ──────────────────────────────────────────────────────────────────────────────
// User model
// ──────────────────────────────────────────────────────────────────────────────

func TestUser_ZeroValue(t *testing.T) {
	var u User
	if u.ID != (uuid.UUID{}) {
		t.Error("zero-value User should have nil UUID")
	}
	if u.Email != "" {
		t.Error("zero-value Email should be empty")
	}
	if u.Role != "" {
		t.Error("zero-value Role should be empty")
	}
}

func TestUser_FieldAssignment(t *testing.T) {
	id := uuid.New()
	u := User{
		ID:           id,
		Email:        "alice@example.com",
		PasswordHash: "hashed-secret",
		Role:         "admin",
		CreatedAt:    time.Now(),
	}

	if u.ID != id {
		t.Errorf("ID mismatch")
	}
	if u.Email != "alice@example.com" {
		t.Errorf("Email mismatch")
	}
	if u.PasswordHash != "hashed-secret" {
		t.Errorf("PasswordHash mismatch")
	}
	if u.Role != "admin" {
		t.Errorf("Role mismatch")
	}
}

func TestUser_JSONOmitsPasswordHash(t *testing.T) {
	u := User{
		ID:           uuid.New(),
		Email:        "bob@example.com",
		PasswordHash: "super-secret-hash",
		Role:         "user",
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	if strings.Contains(string(data), "super-secret-hash") {
		t.Error("JSON output should NOT contain PasswordHash value")
	}
	if strings.Contains(string(data), "password_hash") || strings.Contains(string(data), "passwordHash") {
		// Only fail if the actual hash value is exposed; field name in json tags may vary
	}
}

func TestUser_JSONContainsEmail(t *testing.T) {
	u := User{
		ID:    uuid.New(),
		Email: "carol@example.com",
	}

	data, _ := json.Marshal(u)
	if !strings.Contains(string(data), "carol@example.com") {
		t.Error("JSON output should contain email")
	}
}

func TestUser_JSONRoundTrip(t *testing.T) {
	original := User{
		ID:        uuid.New(),
		Email:     "dave@example.com",
		Role:      "editor",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var restored User
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if restored.Email != original.Email {
		t.Errorf("Email mismatch after round-trip: want %q, got %q", original.Email, restored.Email)
	}
	if restored.Role != original.Role {
		t.Errorf("Role mismatch after round-trip")
	}
}

func TestUser_UniqueIDs(t *testing.T) {
	u1 := User{ID: uuid.New()}
	u2 := User{ID: uuid.New()}
	if u1.ID == u2.ID {
		t.Error("two different users should have different UUIDs")
	}
}

func TestUser_CreatedAtDefaultable(t *testing.T) {
	u := User{CreatedAt: time.Now()}
	if u.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero after explicit assignment")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Integration model
// ──────────────────────────────────────────────────────────────────────────────

func TestIntegration_ZeroValue(t *testing.T) {
	var i Integration
	if i.Provider != "" {
		t.Error("zero-value Provider should be empty")
	}
}

func TestIntegration_FieldAssignment(t *testing.T) {
	userID := uuid.New()
	expiry := time.Now().Add(time.Hour)

	i := Integration{
		ID:           uuid.New(),
		UserID:       userID,
		Provider:     "slack",
		AccessToken:  "xoxb-token",
		RefreshToken: "refresh-xyz",
		ExpiresAt:    expiry,
		Metadata:     []byte(`{"extra":"data"}`),
	}

	if i.Provider != "slack" {
		t.Errorf("Provider mismatch")
	}
	if i.AccessToken != "xoxb-token" {
		t.Errorf("AccessToken mismatch")
	}
	if i.RefreshToken != "refresh-xyz" {
		t.Errorf("RefreshToken mismatch")
	}
	if !i.ExpiresAt.Equal(expiry) {
		t.Errorf("ExpiresAt mismatch")
	}
}

func TestIntegration_MetadataIsJSON(t *testing.T) {
	i := Integration{
		Metadata: []byte(`{"workspace":"T123ABC","bot_user_id":"U9876"}`),
	}

	var m map[string]string
	if err := json.Unmarshal(i.Metadata, &m); err != nil {
		t.Fatalf("Metadata should be valid JSON: %v", err)
	}
	if m["workspace"] != "T123ABC" {
		t.Error("Metadata workspace mismatch")
	}
}

func TestIntegration_AccessTokenNotInJSON(t *testing.T) {
	i := Integration{
		ID:          uuid.New(),
		Provider:    "github",
		AccessToken: "ghp-secret-token",
	}

	data, _ := json.Marshal(i)
	if strings.Contains(string(data), "ghp-secret-token") {
		t.Error("AccessToken should NOT be serialized in JSON output (security)")
	}
}

func TestIntegration_JSONRoundTrip_Provider(t *testing.T) {
	orig := Integration{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Provider: "notion",
	}

	data, _ := json.Marshal(orig)
	var restored Integration
	_ = json.Unmarshal(data, &restored)

	if restored.Provider != orig.Provider {
		t.Errorf("Provider mismatch after round-trip")
	}
}

func TestIntegration_UniqueIDs(t *testing.T) {
	i1 := Integration{ID: uuid.New()}
	i2 := Integration{ID: uuid.New()}
	if i1.ID == i2.ID {
		t.Error("two integrations should have different IDs")
	}
}

func TestIntegration_ExpiryInFuture(t *testing.T) {
	i := Integration{ExpiresAt: time.Now().Add(24 * time.Hour)}
	if i.ExpiresAt.Before(time.Now()) {
		t.Error("future expiry should not be in the past")
	}
}

func TestIntegration_ExpiryInPast(t *testing.T) {
	i := Integration{ExpiresAt: time.Now().Add(-time.Hour)}
	if !i.ExpiresAt.Before(time.Now()) {
		t.Error("past expiry should be before now")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// APIKey model
// ──────────────────────────────────────────────────────────────────────────────

func TestAPIKey_ZeroValue(t *testing.T) {
	var k APIKey
	if k.Name != "" {
		t.Error("zero-value Name should be empty")
	}
	if k.KeyHash != "" {
		t.Error("zero-value KeyHash should be empty")
	}
}

func TestAPIKey_FieldAssignment(t *testing.T) {
	k := APIKey{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		KeyHash:   "sha256:abc123",
		Name:      "production-key",
		Scopes:    pq.StringArray{"read:integrations", "write:workflows"},
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}

	if k.Name != "production-key" {
		t.Errorf("Name mismatch")
	}
	if k.KeyHash != "sha256:abc123" {
		t.Errorf("KeyHash mismatch")
	}
	if len(k.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(k.Scopes))
	}
	if k.Scopes[0] != "read:integrations" {
		t.Errorf("Scopes[0] mismatch")
	}
}

func TestAPIKey_KeyHashNotInJSON(t *testing.T) {
	k := APIKey{
		ID:      uuid.New(),
		Name:    "my-key",
		KeyHash: "sha256:verysecretkeyhash",
	}

	data, _ := json.Marshal(k)
	if strings.Contains(string(data), "verysecretkeyhash") {
		t.Error("KeyHash should NOT be serialized in JSON output (security)")
	}
}

func TestAPIKey_ScopesAsStringArray(t *testing.T) {
	k := APIKey{
		Scopes: pq.StringArray{"read:all", "write:workflows", "admin"},
	}

	if len(k.Scopes) != 3 {
		t.Errorf("expected 3 scopes, got %d", len(k.Scopes))
	}
	for _, s := range k.Scopes {
		if s == "" {
			t.Error("scope should not be empty")
		}
	}
}

func TestAPIKey_EmptyScopes(t *testing.T) {
	k := APIKey{Scopes: pq.StringArray{}}
	if len(k.Scopes) != 0 {
		t.Error("empty scopes should have length 0")
	}
}

func TestAPIKey_UniqueIDs(t *testing.T) {
	k1 := APIKey{ID: uuid.New()}
	k2 := APIKey{ID: uuid.New()}
	if k1.ID == k2.ID {
		t.Error("two API keys should have different IDs")
	}
}

func TestAPIKey_IsExpired_NotExpired(t *testing.T) {
	k := APIKey{ExpiresAt: time.Now().Add(24 * time.Hour)}
	if k.ExpiresAt.Before(time.Now()) {
		t.Error("future key should not be expired")
	}
}

func TestAPIKey_IsExpired_Expired(t *testing.T) {
	k := APIKey{ExpiresAt: time.Now().Add(-time.Hour)}
	if !k.ExpiresAt.Before(time.Now()) {
		t.Error("past key should be expired")
	}
}

func TestAPIKey_JSONContainsName(t *testing.T) {
	k := APIKey{
		ID:   uuid.New(),
		Name: "staging-key",
	}
	data, _ := json.Marshal(k)
	if !strings.Contains(string(data), "staging-key") {
		t.Error("JSON should include Name field")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Cross-model UUID consistency
// ──────────────────────────────────────────────────────────────────────────────

func TestUserIntegrationRelation_UserIDMatches(t *testing.T) {
	userID := uuid.New()

	u := User{ID: userID, Email: "linked@example.com"}
	i := Integration{UserID: userID, Provider: "slack"}

	if u.ID != i.UserID {
		t.Error("User.ID should match Integration.UserID for same user")
	}
}

func TestUserAPIKeyRelation_UserIDMatches(t *testing.T) {
	userID := uuid.New()

	u := User{ID: userID, Email: "linked@example.com"}
	k := APIKey{UserID: userID, Name: "my-key"}

	if u.ID != k.UserID {
		t.Error("User.ID should match APIKey.UserID for same user")
	}
}

func TestUUIDIsValidFormat(t *testing.T) {
	id := uuid.New()
	str := id.String()

	// UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	parts := strings.Split(str, "-")
	if len(parts) != 5 {
		t.Errorf("UUID should have 5 hyphen-separated parts, got %q", str)
	}
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		t.Errorf("UUID format incorrect: %q", str)
	}
}
