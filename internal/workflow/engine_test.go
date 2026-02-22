package workflow

import (
	"context"
	"errors"
	"strings"
	"testing"

	"neighbourhood/internal/integrations"

	"github.com/google/uuid"
)

type fakeProvider struct {
	name       string
	execError  error
	execResult interface{}
}

func (f *fakeProvider) Name() string                   { return f.name }
func (f *fakeProvider) GetAuthURL(state string) string { return "https://fake/auth?state=" + state }
func (f *fakeProvider) ExchangeCode(_ context.Context, _ string) (*integrations.Token, error) {
	return &integrations.Token{AccessToken: "fake-token"}, nil
}
func (f *fakeProvider) Execute(_ context.Context, _ *integrations.Token, _ string, _ map[string]interface{}) (interface{}, error) {
	if f.execError != nil {
		return nil, f.execError
	}
	if f.execResult != nil {
		return f.execResult, nil
	}
	return map[string]interface{}{"ok": true}, nil
}

func setupEngine() *WorkflowEngine {
	integrations.Providers = map[integrations.IntegrationType]integrations.Provider{}
	return NewWorkflowEngine()
}
func reg(name integrations.IntegrationType, p integrations.Provider) {
	integrations.Providers[name] = p
}

func TestNewWorkflowEngine_NotNil(t *testing.T) {
	if NewWorkflowEngine() == nil {
		t.Fatal("NewWorkflowEngine nil")
	}
}

func TestExecute_EmptySteps_NoError(t *testing.T) {
	e := setupEngine()
	wf := Workflow{ID: uuid.New(), Name: "Empty", Steps: []WorkflowStep{}}
	results, err := e.Execute(context.Background(), wf, map[integrations.IntegrationType]*integrations.Token{})
	if err != nil {
		t.Errorf("empty workflow should not error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestExecute_NilSteps_NoError(t *testing.T) {
	e := setupEngine()
	wf := Workflow{ID: uuid.New()}
	_, err := e.Execute(context.Background(), wf, nil)
	if err != nil {
		t.Errorf("nil steps should not error: %v", err)
	}
}

func TestExecute_SingleStep_Success(t *testing.T) {
	e := setupEngine()
	reg("fake-slack", &fakeProvider{name: "fake-slack"})
	wf := Workflow{ID: uuid.New(), Name: "One Step", Steps: []WorkflowStep{{Provider: "fake-slack", Action: "send_message"}}}
	results, err := e.Execute(context.Background(), wf, map[integrations.IntegrationType]*integrations.Token{"fake-slack": {AccessToken: "tk"}})
	if err != nil {
		t.Fatalf("single step error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestExecute_MultipleSteps_AllSucceed(t *testing.T) {
	e := setupEngine()
	reg("fake-A", &fakeProvider{name: "A"})
	reg("fake-B", &fakeProvider{name: "B"})
	wf := Workflow{ID: uuid.New(), Steps: []WorkflowStep{{Provider: "fake-A", Action: "a"}, {Provider: "fake-B", Action: "b"}}}
	tokens := map[integrations.IntegrationType]*integrations.Token{"fake-A": {AccessToken: "ta"}, "fake-B": {AccessToken: "tb"}}
	results, err := e.Execute(context.Background(), wf, tokens)
	if err != nil {
		t.Fatalf("multi-step error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestExecute_ProviderNotRegistered_ReturnsError(t *testing.T) {
	e := setupEngine()
	wf := Workflow{ID: uuid.New(), Steps: []WorkflowStep{{Provider: "ghost", Action: "do"}}}
	_, err := e.Execute(context.Background(), wf, nil)
	if err == nil {
		t.Error("unregistered provider should error")
	}
}

func TestExecute_MissingToken_ReturnsError(t *testing.T) {
	e := setupEngine()
	reg("fake-slack", &fakeProvider{name: "slack"})
	wf := Workflow{ID: uuid.New(), Steps: []WorkflowStep{{Provider: "fake-slack", Action: "send"}}}
	_, err := e.Execute(context.Background(), wf, map[integrations.IntegrationType]*integrations.Token{})
	if err == nil {
		t.Error("missing token should error")
	}
}

func TestExecute_StepFailure_ReturnsError(t *testing.T) {
	e := setupEngine()
	reg("fake-ok", &fakeProvider{name: "ok"})
	reg("fake-fail", &fakeProvider{name: "fail", execError: errors.New("step exploded")})
	wf := Workflow{ID: uuid.New(), Steps: []WorkflowStep{{Provider: "fake-ok", Action: "a"}, {Provider: "fake-fail", Action: "b"}}}
	tokens := map[integrations.IntegrationType]*integrations.Token{"fake-ok": {AccessToken: "t"}, "fake-fail": {AccessToken: "t"}}
	_, err := e.Execute(context.Background(), wf, tokens)
	if err == nil {
		t.Error("step failure should return error")
	}
	if !strings.Contains(err.Error(), "step exploded") {
		t.Errorf("error should contain step message, got: %v", err)
	}
}

func TestExecute_StepFailure_ReturnsPartialResults(t *testing.T) {
	e := setupEngine()
	reg("fake-ok", &fakeProvider{name: "ok", execResult: map[string]interface{}{"done": true}})
	reg("fake-fail", &fakeProvider{name: "fail", execError: errors.New("boom")})
	wf := Workflow{ID: uuid.New(), Steps: []WorkflowStep{{Provider: "fake-ok", Action: "a"}, {Provider: "fake-fail", Action: "b"}}}
	tokens := map[integrations.IntegrationType]*integrations.Token{"fake-ok": {AccessToken: "t"}, "fake-fail": {AccessToken: "t"}}
	results, err := e.Execute(context.Background(), wf, tokens)
	if err == nil {
		t.Error("expected error")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 partial result, got %d", len(results))
	}
}

func TestWorkflowStep_Fields(t *testing.T) {
	s := WorkflowStep{Provider: "slack", Action: "send", Payload: map[string]interface{}{"ch": "#g"}}
	if s.Provider != "slack" {
		t.Error("Provider mismatch")
	}
	if s.Action != "send" {
		t.Error("Action mismatch")
	}
}
func TestWorkflow_Fields(t *testing.T) {
	wf := Workflow{ID: uuid.New(), Name: "My Workflow", Steps: []WorkflowStep{{Provider: "slack", Action: "send"}}}
	if wf.Name != "My Workflow" {
		t.Error("Name mismatch")
	}
	if len(wf.Steps) != 1 {
		t.Error("should have 1 step")
	}
}
