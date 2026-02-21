package mcp

import (
	"context"
	"encoding/json"
	"neighbourhood/internal/integrations"
)

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type CallToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

var Tools = []Tool{
	{
		Name:        "list_integrations",
		Description: "List all connected integrations for the current user.",
		InputSchema: json.RawMessage(`{"type": "object", "properties": {}}`),
	},
	{
		Name:        "execute_integration_action",
		Description: "Execute an action on a connected integration.",
		InputSchema: json.RawMessage(`{
            "type": "object", 
            "properties": {
                "provider": {"type": "string"},
                "action": {"type": "string"},
                "payload": {"type": "object"}
            },
            "required": ["provider", "action"]
        }`),
	},
}

// HandleToolCall processes the tool execution request
func HandleToolCall(req CallToolRequest) (*CallToolResult, error) {
	if req.Name == "list_integrations" {
		// TODO: distinct logic for listing integrations
		return &CallToolResult{
			Content: []ContentItem{
				{Type: "text", Text: "Slack (Connected), Jira (Disconnected)"},
			},
		}, nil
	}

	if req.Name == "execute_integration_action" {
		providerName, _ := req.Arguments["provider"].(string)
		action, _ := req.Arguments["action"].(string)
		payload, _ := req.Arguments["payload"].(map[string]interface{})

		// Mock lookup of provider
		var provider integrations.Provider
		if providerName == "slack" {
			provider = integrations.NewSlackProvider("foo", "bar", "http://localhost:8080/callback")
		}

		if provider == nil {
			return &CallToolResult{IsError: true, Content: []ContentItem{{Type: "text", Text: "Provider not found"}}}, nil
		}

		// Mock token
		token := &integrations.Token{AccessToken: "mock"}

		result, err := provider.Execute(context.Background(), token, action, payload)
		if err != nil {
			return &CallToolResult{IsError: true, Content: []ContentItem{{Type: "text", Text: err.Error()}}}, nil
		}

		jsonResult, _ := json.Marshal(result)

		return &CallToolResult{
			Content: []ContentItem{
				{Type: "text", Text: string(jsonResult)},
			},
		}, nil
	}

	return &CallToolResult{IsError: true, Content: []ContentItem{{Type: "text", Text: "Tool not found"}}}, nil
}
