package workflow

import (
	"context"
	"fmt"

	"neighbourhood/internal/integrations"

	"github.com/google/uuid"
)

// WorkflowStep defines a single step in a workflow
type WorkflowStep struct {
	Provider integrations.IntegrationType
	Action   string
	Payload  map[string]interface{}
}

// Workflow defines a sequence of steps
type Workflow struct {
	ID    uuid.UUID
	Name  string
	Steps []WorkflowStep
}

// WorkflowEngine executes workflows
// In production, add logging, metrics, distributed tracing, and error handling.
type WorkflowEngine struct {
	// Add logger, metrics, etc. here
}

func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{}
}

// Execute runs the workflow steps in order
func (e *WorkflowEngine) Execute(ctx context.Context, wf Workflow, tokens map[integrations.IntegrationType]*integrations.Token) ([]interface{}, error) {
	var results []interface{}
	for i, step := range wf.Steps {
		provider, err := integrations.GetProvider(step.Provider)
		if err != nil {
			return results, fmt.Errorf("provider %s not found at step %d: %w", step.Provider, i, err)
		}
		token, ok := tokens[step.Provider]
		if !ok {
			return results, fmt.Errorf("token for provider %s not found at step %d", step.Provider, i)
		}
		res, err := provider.Execute(ctx, token, step.Action, step.Payload)
		if err != nil {
			// In production, log error, maybe continue or rollback
			return results, fmt.Errorf("step %d failed: %w", i, err)
		}
		results = append(results, res)
	}
	return results, nil
}
