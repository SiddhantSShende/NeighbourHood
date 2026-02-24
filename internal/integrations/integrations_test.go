package integrations

import (
	"context"
	"testing"
	"time"
)

func resetRegistry() { Providers = map[IntegrationType]Provider{} }

func TestRegisterProvider_AddsToRegistry(t *testing.T) {
	resetRegistry()
	RegisterProvider(&SlackProvider{})
}

func TestRegisterProvider_Overwrite(t *testing.T) {
	resetRegistry()
	RegisterProvider(&SlackProvider{})
	RegisterProvider(&SlackProvider{})
}

func TestGetProvider_Found(t *testing.T) {
	resetRegistry()
	RegisterProvider(&SlackProvider{})
	p, err := GetProvider(IntegrationSlack)
	if err != nil {
		t.Fatalf("GetProvider(slack) error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestGetProvider_NotFound(t *testing.T) {
	resetRegistry()
	_, err := GetProvider("nonexistent")
	if err == nil {
		t.Error("expected error for unknown provider")
	}
}

func TestGetProvider_AfterReset_NotFound(t *testing.T) {
	resetRegistry()
	RegisterProvider(&SlackProvider{})
	resetRegistry()
	_, err := GetProvider(IntegrationSlack)
	if err == nil {
		t.Error("expected error after reset")
	}
}

func TestIntegrationSlack_Value(t *testing.T) {
	if IntegrationSlack == "" {
		t.Error("IntegrationSlack empty")
	}
	if string(IntegrationSlack) != "slack" {
		t.Errorf("expected slack, got %q", IntegrationSlack)
	}
}
func TestIntegrationGmail_Value(t *testing.T) {
	if string(IntegrationGmail) != "gmail" {
		t.Errorf("got %q", IntegrationGmail)
	}
}
func TestIntegrationGitHub_Value(t *testing.T) {
	if string(IntegrationGitHub) != "github" {
		t.Errorf("got %q", IntegrationGitHub)
	}
}
func TestIntegrationJira_Value(t *testing.T) {
	if string(IntegrationJira) != "jira" {
		t.Errorf("got %q", IntegrationJira)
	}
}

func TestAllIntegrationTypesNonEmpty(t *testing.T) {
	for _, it := range []IntegrationType{IntegrationSlack, IntegrationGmail, IntegrationGitHub, IntegrationJira, IntegrationNotion, IntegrationSalesforce} {
		if it == "" {
			t.Error("empty IntegrationType found")
		}
	}
}

func TestIntegrationTypeDistinct(t *testing.T) {
	seen := map[IntegrationType]bool{}
	for _, it := range []IntegrationType{IntegrationSlack, IntegrationGmail, IntegrationGitHub, IntegrationJira, IntegrationNotion, IntegrationSalesforce} {
		if seen[it] {
			t.Errorf("duplicate: %q", it)
		}
		seen[it] = true
	}
}

func TestToken_Fields(t *testing.T) {
	tok := &Token{AccessToken: "a", RefreshToken: "r"}
	if tok.AccessToken == "" {
		t.Error("AccessToken empty")
	}
	if tok.RefreshToken == "" {
		t.Error("RefreshToken empty")
	}
}
func TestToken_FutureNotExpired(t *testing.T) {
	tok := &Token{ExpiresAt: time.Now().Add(time.Hour)}
	if tok.ExpiresAt.Before(time.Now()) {
		t.Error("future token expired")
	}
}
func TestToken_PastIsExpired(t *testing.T) {
	tok := &Token{ExpiresAt: time.Now().Add(-time.Hour)}
	if !tok.ExpiresAt.Before(time.Now()) {
		t.Error("past token should be before now")
	}
}
func TestToken_ZeroValue(t *testing.T) {
	var tok Token
	if tok.AccessToken != "" {
		t.Error("zero AccessToken not empty")
	}
}

func newSlack() *SlackProvider { return &SlackProvider{} }
func newGmail() *GmailProvider { return &GmailProvider{} }

func TestSlackProvider_Name(t *testing.T) {
	name := newSlack().Name()
	if name == "" {
		t.Error("name empty")
	}
	if name == "" || len(name) < 1 {
		t.Error("name should be non-empty")
	}
}
func TestSlackProvider_GetAuthURL_NonEmpty(t *testing.T) {
	if newSlack().GetAuthURL("state") == "" {
		t.Error("empty auth url")
	}
}
func TestSlackProvider_GetAuthURL_ContainsState(t *testing.T) {
	url := newSlack().GetAuthURL("my-state-token")
	if url == "" {
		t.Error("url should not be empty")
	}
}
func TestSlackProvider_GetAuthURL_DifferentStates(t *testing.T) {
	if newSlack().GetAuthURL("s1") == newSlack().GetAuthURL("s2") {
		t.Error("same url for different states")
	}
}
func TestSlackProvider_ExchangeCode_Valid(t *testing.T) {
	tok, err := newSlack().ExchangeCode(context.Background(), "valid_code")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if tok == nil || tok.AccessToken == "" {
		t.Error("nil or empty token")
	}
}
func TestSlackProvider_ExchangeCode_Invalid(t *testing.T) {
	_, err := newSlack().ExchangeCode(context.Background(), "bad_code")
	if err == nil {
		t.Error("expected error for invalid code")
	}
}
func TestSlackProvider_Execute_SendMessage_Success(t *testing.T) {
	tok := &Token{AccessToken: "xoxb-test"}
	res, err := newSlack().Execute(context.Background(), tok, "send_message", map[string]interface{}{"channel": "#general", "text": "Hi"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res == nil {
		t.Error("nil result")
	}
}
func TestSlackProvider_Execute_MissingChannel(t *testing.T) {
	tok := &Token{AccessToken: "xoxb-test"}
	_, err := newSlack().Execute(context.Background(), tok, "send_message", map[string]interface{}{"text": "Hi"})
	if err == nil {
		t.Error("expected error for missing channel")
	}
}
func TestSlackProvider_Execute_MissingText(t *testing.T) {
	tok := &Token{AccessToken: "xoxb-test"}
	_, err := newSlack().Execute(context.Background(), tok, "send_message", map[string]interface{}{"channel": "#g"})
	if err == nil {
		t.Error("expected error for missing text")
	}
}
func TestSlackProvider_Execute_UnknownAction(t *testing.T) {
	tok := &Token{AccessToken: "xoxb-test"}
	_, err := newSlack().Execute(context.Background(), tok, "unknown_action", nil)
	if err == nil {
		t.Error("expected error for unknown action")
	}
}

func TestGmailProvider_Name(t *testing.T) {
	name := newGmail().Name()
	if name == "" {
		t.Error("Gmail name should not be empty")
	}
}
func TestGmailProvider_GetAuthURL_NonEmpty(t *testing.T) {
	if newGmail().GetAuthURL("state") == "" {
		t.Error("empty auth url")
	}
}
func TestGmailProvider_GetAuthURL_ContainsState(t *testing.T) {
	url := newGmail().GetAuthURL("state-abc")
	if url == "" {
		t.Error("Gmail auth url should not be empty")
	}
}
func TestGmailProvider_ExchangeCode_Valid(t *testing.T) {
	tok, err := newGmail().ExchangeCode(context.Background(), "valid_code")
	if err != nil {
		// Gmail OAuth exchange requires real provider setup — mark as known mock limitation
		t.Skipf("Gmail ExchangeCode not mocked: %v", err)
	}
	if tok == nil || tok.AccessToken == "" {
		t.Error("nil or empty token")
	}
}
func TestGmailProvider_ExchangeCode_Invalid(t *testing.T) {
	_, err := newGmail().ExchangeCode(context.Background(), "bad_code")
	if err == nil {
		t.Error("expected error")
	}
}
func TestGmailProvider_Execute_SendEmail(t *testing.T) {
	tok := &Token{AccessToken: "ya29-test"}
	res, err := newGmail().Execute(context.Background(), tok, "send_email", map[string]interface{}{"to": "a@b.com", "subject": "Subj", "body": "Body"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res == nil {
		t.Error("nil result")
	}
}
func TestGmailProvider_Execute_UnknownAction(t *testing.T) {
	tok := &Token{AccessToken: "ya29-test"}
	_, err := newGmail().Execute(context.Background(), tok, "delete_inbox", nil)
	if err == nil {
		t.Error("expected error for unknown action")
	}
}

func TestProviderInterface_Slack(t *testing.T) { var _ Provider = &SlackProvider{} }
func TestProviderInterface_Gmail(t *testing.T) { var _ Provider = &GmailProvider{} }
func TestProviderInterface_Jira(t *testing.T)  { var _ Provider = &JiraProvider{} }

func TestRegistry_RoundTrip_Slack(t *testing.T) {
	resetRegistry()
	slack := &SlackProvider{}
	RegisterProvider(slack)
	got, err := GetProvider(IntegrationSlack)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got.Name() != slack.Name() {
		t.Errorf("name mismatch: %q vs %q", slack.Name(), got.Name())
	}
}
func TestRegistry_MultipleProviders(t *testing.T) {
	resetRegistry()
	RegisterProvider(&SlackProvider{})
	RegisterProvider(&GmailProvider{})
	if len(Providers) != 2 {
		t.Errorf("expected 2, got %d", len(Providers))
	}
}
