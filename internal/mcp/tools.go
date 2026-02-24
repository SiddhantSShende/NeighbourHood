package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"neighbourhood/internal/integrations"
)

// Tool describes a callable MCP tool exposed to AI clients.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// CallToolRequest carries the tool invocation from the client.
type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// CallToolResult is returned after executing a tool.
type CallToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ContentItem is a single piece of content in a tool result.
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Tools is the static list of tools advertised to MCP clients.
var Tools = []Tool{
	{
		Name:        "list_integrations",
		Description: "List all registered integration providers.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
	},
	{
		Name:        "execute_integration_action",
		Description: "Execute an action on a registered integration provider.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"provider": {"type": "string", "description": "Integration provider name (e.g. slack, github)"},
				"action":   {"type": "string", "description": "Action to perform"},
				"payload":  {"type": "object", "description": "Action-specific parameters"},
				"token":    {"type": "object", "description": "Access token object"}
			},
			"required": ["provider", "action"]
		}`),
	},
}

// HandleToolCall dispatches a tool call to the appropriate handler.
// ctx is propagated to downstream provider Execute calls.
func HandleToolCall(ctx context.Context, req CallToolRequest) (*CallToolResult, error) {
	switch req.Name {
	case "list_integrations":
		names := make([]string, 0, len(integrations.Providers))
		for k := range integrations.Providers {
			names = append(names, string(k))
		}
		sort.Strings(names)
		out, _ := json.Marshal(names)
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: string(out)}},
		}, nil

	case "execute_integration_action":
		providerName, _ := req.Arguments["provider"].(string)
		action, _ := req.Arguments["action"].(string)
		payload, _ := req.Arguments["payload"].(map[string]interface{})
		if payload == nil {
			payload = map[string]interface{}{}
		}

		if providerName == "" || action == "" {
			return &CallToolResult{
				IsError: true,
				Content: []ContentItem{{Type: "text", Text: "provider and action are required"}},
			}, nil
		}

		provider, err := integrations.GetProvider(integrations.IntegrationType(providerName))
		if err != nil {
			return &CallToolResult{
				IsError: true,
				Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("provider '%s' not found", providerName)}},
			}, nil
		}

		// Build token from arguments if provided.
		var token integrations.Token
		if rawToken, ok := req.Arguments["token"].(map[string]interface{}); ok {
			if at, ok := rawToken["access_token"].(string); ok {
				token.AccessToken = at
			}
		}

		result, err := provider.Execute(ctx, &token, action, payload)
		if err != nil {
			return &CallToolResult{
				IsError: true,
				Content: []ContentItem{{Type: "text", Text: err.Error()}},
			}, nil
		}

		jsonResult, _ := json.Marshal(result)
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: string(jsonResult)}},
		}, nil

	default:
		return &CallToolResult{
			IsError: true,
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("tool '%s' not found", req.Name)}},
		}, nil
	}
}
