package mcp

import (
	"encoding/json"
	"net/http"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Handle "tools/list"
	if req.Method == "tools/list" {
		json.NewEncoder(w).Encode(JSONRPCResponse{
			JSONRPC: "2.0",
			Result: map[string]interface{}{
				"tools": Tools,
			},
			ID: req.ID,
		})
		return
	}

	// Handle "tools/call"
	if req.Method == "tools/call" {
		var callReq CallToolRequest
		if err := json.Unmarshal(req.Params, &callReq); err != nil {
			// Handle error
		}

		result, _ := HandleToolCall(callReq)

		json.NewEncoder(w).Encode(JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  result,
			ID:      req.ID,
		})
		return
	}

	// Default: Method not found
	json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0",
		Error: map[string]interface{}{
			"code":    -32601,
			"message": "Method not found",
		},
		ID: req.ID,
	})
}
