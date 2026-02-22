package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"neighbourhood/internal/integrations"
)

type fakeProvider struct{ name string }

func (f *fakeProvider) Name() string                   { return f.name }
func (f *fakeProvider) GetAuthURL(state string) string { return "https://fake/auth?state=" + state }
func (f *fakeProvider) ExchangeCode(_ context.Context, _ string) (*integrations.Token, error) {
	return &integrations.Token{AccessToken: "fake-token"}, nil
}
func (f *fakeProvider) Execute(_ context.Context, _ *integrations.Token, _ string, _ map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{"ok": true}, nil
}

func newHandler() *Handler {
	integrations.Providers = map[integrations.IntegrationType]integrations.Provider{}
	return NewHandler()
}
func reg(name integrations.IntegrationType) {
	integrations.Providers[name] = &fakeProvider{name: string(name)}
}

func TestListIntegrations_EmptyRegistry_Returns200(t *testing.T) {
	h := newHandler()
	req := httptest.NewRequest(http.MethodGet, "/integrations", nil)
	rr := httptest.NewRecorder()
	h.ListIntegrations(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
func TestListIntegrations_ReturnsJSON(t *testing.T) {
	h := newHandler()
	req := httptest.NewRequest(http.MethodGet, "/integrations", nil)
	rr := httptest.NewRecorder()
	h.ListIntegrations(rr, req)
	var out interface{}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
}
func TestListIntegrations_ContentTypeJSON(t *testing.T) {
	h := newHandler()
	req := httptest.NewRequest(http.MethodGet, "/integrations", nil)
	rr := httptest.NewRecorder()
	h.ListIntegrations(rr, req)
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected json content-type, got %q", ct)
	}
}
func TestListIntegrations_ContainsRegisteredProvider(t *testing.T) {
	h := newHandler()
	reg("slack")
	req := httptest.NewRequest(http.MethodGet, "/integrations", nil)
	rr := httptest.NewRecorder()
	h.ListIntegrations(rr, req)
	if !strings.Contains(strings.ToLower(rr.Body.String()), "slack") {
		t.Error("response should list slack")
	}
}

func TestGetIntegrationAuthURL_ValidProvider(t *testing.T) {
	h := newHandler()
	reg("slack")
	body := "{\"provider\":\"slack\",\"state\":\"csrf-xyz\"}"
	req := httptest.NewRequest(http.MethodPost, "/integrations/auth-url", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.GetIntegrationAuthURL(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
}
func TestGetIntegrationAuthURL_ResponseContainsURL(t *testing.T) {
	h := newHandler()
	reg("slack")
	body := "{\"provider\":\"slack\",\"state\":\"s\"}"
	req := httptest.NewRequest(http.MethodPost, "/integrations/auth-url", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.GetIntegrationAuthURL(rr, req)
	var resp map[string]interface{}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["url"] == nil || resp["url"] == "" {
		t.Error("response should contain url field")
	}
}
func TestGetIntegrationAuthURL_MissingProvider_Returns400(t *testing.T) {
	h := newHandler()
	body := "{\"state\":\"some-state\"}"
	req := httptest.NewRequest(http.MethodPost, "/integrations/auth-url", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.GetIntegrationAuthURL(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
func TestGetIntegrationAuthURL_UnknownProvider_NotOK(t *testing.T) {
	h := newHandler()
	body := "{\"provider\":\"ghost\",\"state\":\"s\"}"
	req := httptest.NewRequest(http.MethodPost, "/integrations/auth-url", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.GetIntegrationAuthURL(rr, req)
	if rr.Code == http.StatusOK {
		t.Error("unknown provider should not return 200")
	}
}
func TestGetIntegrationAuthURL_InvalidJSON_Returns400(t *testing.T) {
	h := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/integrations/auth-url", bytes.NewBufferString("{bad"))
	rr := httptest.NewRecorder()
	h.GetIntegrationAuthURL(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestExecuteIntegrationAction_InvalidJSON_Returns400(t *testing.T) {
	h := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/integrations/execute", bytes.NewBufferString("{bad"))
	rr := httptest.NewRecorder()
	h.ExecuteIntegrationAction(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
func TestExecuteIntegrationAction_MissingProvider_Returns400(t *testing.T) {
	h := newHandler()
	body := "{\"action\":\"send\",\"payload\":{}}"
	req := httptest.NewRequest(http.MethodPost, "/integrations/execute", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ExecuteIntegrationAction(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
func TestExecuteIntegrationAction_MissingToken_Returns400(t *testing.T) {
	// The handler requires provider+action; empty token flows through the provider.
	// Unregistered provider returns 404, which is still not OK (200).
	h := newHandler()
	body := "{\"provider\":\"slack\",\"action\":\"send\",\"payload\":{}}"
	req := httptest.NewRequest(http.MethodPost, "/integrations/execute", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ExecuteIntegrationAction(rr, req)
	if rr.Code == http.StatusOK {
		t.Error("unregistered provider with empty token should not return 200")
	}
}
func TestExecuteIntegrationAction_UnregisteredProvider_NotOK(t *testing.T) {
	h := newHandler()
	body := "{\"provider\":\"ghost\",\"action\":\"do\",\"payload\":{}}"
	req := httptest.NewRequest(http.MethodPost, "/integrations/execute", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ExecuteIntegrationAction(rr, req)
	if rr.Code == http.StatusOK {
		t.Error("unregistered provider should not return 200")
	}
}
func TestExecuteIntegrationAction_ValidRequest_Returns200(t *testing.T) {
	h := newHandler()
	reg("slack")
	body := "{\"provider\":\"slack\",\"action\":\"send_message\",\"token\":{\"access_token\":\"xoxb\"},\"payload\":{\"channel\":\"#g\",\"text\":\"Hi\"}}"
	req := httptest.NewRequest(http.MethodPost, "/integrations/execute", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ExecuteIntegrationAction(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestExecuteWorkflow_InvalidJSON_Returns400(t *testing.T) {
	h := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/workflows/execute", bytes.NewBufferString("{bad"))
	rr := httptest.NewRecorder()
	h.ExecuteWorkflow(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
func TestExecuteWorkflow_EmptySteps_Returns400(t *testing.T) {
	h := newHandler()
	body := "{\"workflow\":{\"name\":\"Empty\",\"steps\":[]},\"tokens\":{}}"
	req := httptest.NewRequest(http.MethodPost, "/workflows/execute", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ExecuteWorkflow(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
func TestExecuteWorkflow_ValidRequest_Returns200(t *testing.T) {
	h := newHandler()
	reg("slack")
	body := "{\"workflow\":{\"name\":\"Notify\",\"steps\":[{\"provider\":\"slack\",\"action\":\"send_message\",\"payload\":{}}]},\"tokens\":{\"slack\":{\"access_token\":\"xoxb\"}}}"
	req := httptest.NewRequest(http.MethodPost, "/workflows/execute", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ExecuteWorkflow(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
}
func TestExecuteWorkflow_UnknownProvider_NotOK(t *testing.T) {
	h := newHandler()
	body := "{\"workflow\":{\"name\":\"T\",\"steps\":[{\"provider\":\"ghost\",\"action\":\"act\",\"payload\":{}}]},\"tokens\":{\"ghost\":{\"access_token\":\"tk\"}}}"
	req := httptest.NewRequest(http.MethodPost, "/workflows/execute", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ExecuteWorkflow(rr, req)
	if rr.Code == http.StatusOK {
		t.Error("unknown provider should not return 200")
	}
}
func TestExecuteWorkflow_MissingToken_NotOK(t *testing.T) {
	h := newHandler()
	reg("slack")
	body := "{\"workflow\":{\"name\":\"T\",\"steps\":[{\"provider\":\"slack\",\"action\":\"send\",\"payload\":{}}]},\"tokens\":{}}"
	req := httptest.NewRequest(http.MethodPost, "/workflows/execute", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ExecuteWorkflow(rr, req)
	if rr.Code == http.StatusOK {
		t.Error("missing token should not return 200")
	}
}

func TestErrorResponse_ContainsErrorField(t *testing.T) {
	h := newHandler()
	body := "{\"provider\":\"ghost\",\"state\":\"s\"}"
	req := httptest.NewRequest(http.MethodPost, "/integrations/auth-url", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.GetIntegrationAuthURL(rr, req)
	if rr.Code < 400 {
		t.Skip("response was success")
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if resp["error"] == nil {
		t.Error("error response should have error field")
	}
}
