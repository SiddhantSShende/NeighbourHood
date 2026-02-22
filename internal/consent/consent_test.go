package consent

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestGrant_ReturnsConsent(t *testing.T) {
	m := NewManager()
	c, err := m.Grant(context.Background(), uuid.New(), "slack", "integration")
	if err != nil {
		t.Fatalf("Grant error: %v", err)
	}
	if c == nil {
		t.Fatal("Grant returned nil consent")
	}
}
func TestGrant_ConsentHasID(t *testing.T) {
	m := NewManager()
	c, _ := m.Grant(context.Background(), uuid.New(), "slack", "integration")
	if c.ID == (uuid.UUID{}) {
		t.Error("consent ID should be set")
	}
}
func TestGrant_ConsentHasCorrectUserID(t *testing.T) {
	m := NewManager()
	uid := uuid.New()
	c, _ := m.Grant(context.Background(), uid, "slack", "integration")
	if c.UserID != uid {
		t.Errorf("UserID mismatch: want %v, got %v", uid, c.UserID)
	}
}
func TestGrant_ConsentHasCorrectProvider(t *testing.T) {
	m := NewManager()
	c, _ := m.Grant(context.Background(), uuid.New(), "gmail", "marketing")
	if c.Provider != "gmail" {
		t.Errorf("Provider mismatch: want gmail, got %q", c.Provider)
	}
}
func TestGrant_ConsentStatusIsGranted(t *testing.T) {
	m := NewManager()
	c, _ := m.Grant(context.Background(), uuid.New(), "slack", "integration")
	if c.Status != ConsentGranted {
		t.Errorf("status want ConsentGranted, got %q", c.Status)
	}
}
func TestGrant_GrantedAtIsSet(t *testing.T) {
	m := NewManager()
	c, _ := m.Grant(context.Background(), uuid.New(), "slack", "integration")
	if c.GrantedAt == nil {
		t.Error("GrantedAt should be set after Grant")
	}
}
func TestGrant_UniqueIDs(t *testing.T) {
	m := NewManager()
	uid := uuid.New()
	c1, _ := m.Grant(context.Background(), uid, "slack", "integration")
	c2, _ := m.Grant(context.Background(), uid, "gmail", "integration")
	if c1.ID == c2.ID {
		t.Error("two grants should have unique IDs")
	}
}
func TestGrant_PurposeIsPreserved(t *testing.T) {
	m := NewManager()
	c, _ := m.Grant(context.Background(), uuid.New(), "jira", "analytics")
	if c.Purpose != "analytics" {
		t.Errorf("purpose mismatch: want analytics, got %q", c.Purpose)
	}
}

func TestRevoke_NoError(t *testing.T) {
	m := NewManager()
	if err := m.Revoke(context.Background(), uuid.New()); err != nil {
		t.Errorf("Revoke should not error (mock): %v", err)
	}
}
func TestRevoke_KnownConsent_NoError(t *testing.T) {
	m := NewManager()
	c, _ := m.Grant(context.Background(), uuid.New(), "slack", "integration")
	if err := m.Revoke(context.Background(), c.ID); err != nil {
		t.Errorf("Revoke of known consent should not error: %v", err)
	}
}

func TestCheck_MockReturnsTrue(t *testing.T) {
	m := NewManager()
	ok, err := m.Check(context.Background(), uuid.New(), "slack")
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if !ok {
		t.Error("mock Check should return true")
	}
}
func TestCheck_AnyProvider_ReturnsTrue(t *testing.T) {
	m := NewManager()
	for _, p := range []string{"slack", "gmail", "jira", "github", "unknown"} {
		ok, err := m.Check(context.Background(), uuid.New(), p)
		if err != nil {
			t.Errorf("Check(%q) error: %v", p, err)
		}
		if !ok {
			t.Errorf("mock Check(%q) should return true", p)
		}
	}
}

func TestList_EmptyByDefault(t *testing.T) {
	m := NewManager()
	consents, err := m.List(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(consents) != 0 {
		t.Errorf("mock List should return empty slice, got %d items", len(consents))
	}
}

func TestValidateConsent_SlackRequired_MockPasses(t *testing.T) {
	m := NewManager()
	// Check returns true (mock), so ValidateConsent should pass
	if err := m.ValidateConsent(context.Background(), uuid.New(), "slack"); err != nil {
		t.Errorf("ValidateConsent with mock Check=true should pass: %v", err)
	}
}
func TestValidateConsent_GmailRequired_MockPasses(t *testing.T) {
	m := NewManager()
	if err := m.ValidateConsent(context.Background(), uuid.New(), "gmail"); err != nil {
		t.Errorf("ValidateConsent gmail should pass with mock: %v", err)
	}
}
func TestValidateConsent_JiraRequired_MockPasses(t *testing.T) {
	m := NewManager()
	if err := m.ValidateConsent(context.Background(), uuid.New(), "jira"); err != nil {
		t.Errorf("ValidateConsent jira should pass with mock: %v", err)
	}
}
func TestValidateConsent_GitHubNotRequired_Passes(t *testing.T) {
	m := NewManager()
	if err := m.ValidateConsent(context.Background(), uuid.New(), "github"); err != nil {
		t.Errorf("github does not require consent, should pass: %v", err)
	}
}

func TestIntegrationConsentRequired_GitHub(t *testing.T) {
	if IntegrationConsentRequired("github") {
		t.Error("github should NOT require consent")
	}
}
func TestIntegrationConsentRequired_Unknown(t *testing.T) {
	if IntegrationConsentRequired("unknown-xyz") {
		t.Error("unknown should NOT require consent")
	}
}

func TestConsentStatusValues(t *testing.T) {
	if ConsentGranted == "" {
		t.Error("ConsentGranted empty")
	}
	if ConsentRevoked == "" {
		t.Error("ConsentRevoked empty")
	}
	if ConsentPending == "" {
		t.Error("ConsentPending empty")
	}
	if ConsentGranted == ConsentRevoked {
		t.Error("Granted and Revoked must differ")
	}
	if ConsentGranted == ConsentPending {
		t.Error("Granted and Pending must differ")
	}
}
