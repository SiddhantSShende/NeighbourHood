package mcp

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSONRPCRequest is a JSON-RPC 2.0 request envelope.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

// JSONRPCResponse is a JSON-RPC 2.0 response envelope.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// writeJSON encodes v as JSON and writes it to w.
// Encoding errors are logged but cannot be surfaced to the client because
// the HTTP status has already been committed.
func writeJSON(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("mcp: failed to encode JSON response: %v", err)
	}
}

// Handler is the HTTP handler for the MCP JSON-RPC endpoint.
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case "tools/list":
		writeJSON(w, JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  map[string]interface{}{"tools": Tools},
			ID:      req.ID,
		})

	case "tools/call":
		var callReq CallToolRequest
		if err := json.Unmarshal(req.Params, &callReq); err != nil {
			writeJSON(w, JSONRPCResponse{
				JSONRPC: "2.0",
				Error:   map[string]interface{}{"code": -32602, "message": "invalid params: " + err.Error()},
				ID:      req.ID,
			})
			return
		}

		result, err := HandleToolCall(r.Context(), callReq)
		if err != nil {
			writeJSON(w, JSONRPCResponse{
				JSONRPC: "2.0",
				Error:   map[string]interface{}{"code": -32603, "message": err.Error()},
				ID:      req.ID,
			})
			return
		}
		writeJSON(w, JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  result,
			ID:      req.ID,
		})

	default:
		writeJSON(w, JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   map[string]interface{}{"code": -32601, "message": "method not found"},
			ID:      req.ID,
		})
	}
}
