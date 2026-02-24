package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"

	"neighbourhood/internal/consent"
	"neighbourhood/internal/integrations"
	"neighbourhood/internal/middleware"
	"neighbourhood/internal/workflow"

	"github.com/google/uuid"
)

// maxRequestBodySize is the maximum number of bytes accepted from an HTTP
// request body. Requests larger than this are rejected with 413.
const maxRequestBodySize = 1 << 20 // 1 MiB

// Handler manages API routes and dependencies
type Handler struct {
	consentManager *consent.Manager
}

// NewHandler creates a new API handler
func NewHandler() *Handler {
	return &Handler{
		consentManager: consent.NewManager(),
	}
}

// GetIntegrationAuthURL returns the OAuth URL for a provider
func (h *Handler) GetIntegrationAuthURL(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Provider string `json:"provider"`
		State    string `json:"state"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider == "" {
		respondError(w, "provider is required", http.StatusBadRequest)
		return
	}

	provider, err := integrations.GetProvider(integrations.IntegrationType(req.Provider))
	if err != nil {
		respondError(w, "provider not found", http.StatusNotFound)
		return
	}

	url := provider.GetAuthURL(req.State)
	respondJSON(w, map[string]string{"url": url}, http.StatusOK)
}

// ExecuteIntegrationAction executes a single integration action
func (h *Handler) ExecuteIntegrationAction(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Provider string                 `json:"provider"`
		Token    integrations.Token     `json:"token"`
		Action   string                 `json:"action"`
		Payload  map[string]interface{} `json:"payload"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Provider == "" || req.Action == "" {
		respondError(w, "provider and action are required", http.StatusBadRequest)
		return
	}

	// Extract authenticated user ID from context (set by Auth middleware).
	// Falls back to a sentinel UUID in dev/demo mode when auth is bypassed.
	userID := extractUserID(r)
	if err := h.consentManager.ValidateConsent(r.Context(), userID, req.Provider); err != nil {
		respondError(w, "consent not granted: "+err.Error(), http.StatusForbidden)
		return
	}

	provider, err := integrations.GetProvider(integrations.IntegrationType(req.Provider))
	if err != nil {
		respondError(w, "provider not found", http.StatusNotFound)
		return
	}

	result, err := provider.Execute(r.Context(), &req.Token, req.Action, req.Payload)
	if err != nil {
		log.Printf("Integration execution error: %v", err)
		respondError(w, "execution failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{"result": result}, http.StatusOK)
}

// ExecuteWorkflow executes a multi-step workflow
func (h *Handler) ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Workflow workflow.Workflow             `json:"workflow"`
		Tokens   map[string]integrations.Token `json:"tokens"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate workflow
	if len(req.Workflow.Steps) == 0 {
		respondError(w, "workflow must have at least one step", http.StatusBadRequest)
		return
	}

	// Extract authenticated user ID from context (set by Auth middleware).
	// Falls back to a sentinel UUID in dev/demo mode when auth is bypassed.
	userID := extractUserID(r)
	for _, step := range req.Workflow.Steps {
		if err := h.consentManager.ValidateConsent(r.Context(), userID, string(step.Provider)); err != nil {
			respondError(w, "consent not granted for "+string(step.Provider)+": "+err.Error(), http.StatusForbidden)
			return
		}
	}

	// Convert tokens to correct type
	tokens := make(map[integrations.IntegrationType]*integrations.Token)
	for k, v := range req.Tokens {
		token := v
		tokens[integrations.IntegrationType(k)] = &token
	}

	engine := workflow.NewWorkflowEngine()
	results, err := engine.Execute(r.Context(), req.Workflow, tokens)
	if err != nil {
		log.Printf("Workflow execution error: %v", err)
		respondError(w, "workflow execution failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{"results": results}, http.StatusOK)
}

// ListIntegrations returns all available integrations, sorted by type for
// deterministic output regardless of map iteration order.
func (h *Handler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	integrationsList := make([]map[string]interface{}, 0, len(integrations.Providers))

	for providerType := range integrations.Providers {
		category, description := getProviderInfo(string(providerType))
		integrationsList = append(integrationsList, map[string]interface{}{
			"type":        string(providerType),
			"name":        formatProviderName(string(providerType)),
			"description": description,
			"category":    category,
		})
	}

	sort.Slice(integrationsList, func(i, j int) bool {
		iType, _ := integrationsList[i]["type"].(string)
		jType, _ := integrationsList[j]["type"].(string)
		return iType < jType
	})

	respondJSON(w, map[string]interface{}{
		"integrations": integrationsList,
		"total":        len(integrationsList),
	}, http.StatusOK)
}

// formatProviderName formats provider type to display name
func formatProviderName(providerType string) string {
	names := map[string]string{
		"slack": "Slack", "microsoft_teams": "Microsoft Teams", "zoom": "Zoom", "discord": "Discord",
		"gmail": "Gmail", "sendgrid": "SendGrid", "mailchimp": "Mailchimp", "twilio": "Twilio",
		"jira": "Jira", "trello": "Trello", "asana": "Asana", "monday": "Monday.com",
		"notion": "Notion", "clickup": "ClickUp",
		"salesforce": "Salesforce", "hubspot": "HubSpot", "zendesk": "Zendesk",
		"intercom": "Intercom", "pipedrive": "Pipedrive",
		"github": "GitHub", "gitlab": "GitLab", "bitbucket": "Bitbucket",
		"dropbox": "Dropbox", "google_drive": "Google Drive", "onedrive": "OneDrive", "box": "Box",
		"stripe": "Stripe", "shopify": "Shopify", "paypal": "PayPal", "square": "Square",
		"airtable": "Airtable", "google_sheets": "Google Sheets", "tableau": "Tableau",
		"microsoft_excel": "Microsoft Excel",
		"twitter":         "Twitter", "linkedin": "LinkedIn", "facebook": "Facebook", "instagram": "Instagram",
	}
	if name, ok := names[providerType]; ok {
		return name
	}
	return providerType
}

// getProviderInfo returns category and description for a provider
func getProviderInfo(providerType string) (string, string) {
	info := map[string]struct {
		category    string
		description string
	}{
		// Communication & Collaboration
		"slack":           {"Communication", "Team collaboration and messaging"},
		"microsoft_teams": {"Communication", "Microsoft Teams workspace collaboration"},
		"zoom":            {"Communication", "Video conferencing and meetings"},
		"discord":         {"Communication", "Voice, video, and text chat platform"},

		// Email & Marketing
		"gmail":     {"Email", "Google email service"},
		"sendgrid":  {"Email", "Email delivery platform"},
		"mailchimp": {"Marketing", "Email marketing and automation"},
		"twilio":    {"Communication", "SMS and voice communication"},

		// Project Management
		"jira":    {"Project Management", "Issue tracking and project management"},
		"trello":  {"Project Management", "Visual project boards"},
		"asana":   {"Project Management", "Team task and project management"},
		"monday":  {"Project Management", "Work operating system"},
		"notion":  {"Productivity", "All-in-one workspace"},
		"clickup": {"Project Management", "Productivity and project management"},

		// CRM & Sales
		"salesforce": {"CRM", "Customer relationship management"},
		"hubspot":    {"CRM", "Marketing, sales, and service platform"},
		"zendesk":    {"Support", "Customer service and support"},
		"intercom":   {"Support", "Customer messaging platform"},
		"pipedrive":  {"CRM", "Sales CRM and pipeline management"},

		// Development & Code
		"github":    {"Development", "Code hosting and version control"},
		"gitlab":    {"Development", "DevOps platform and Git repository"},
		"bitbucket": {"Development", "Git repository management"},

		// Storage & Documents
		"dropbox":      {"Storage", "Cloud file storage and sharing"},
		"google_drive": {"Storage", "Google cloud storage and docs"},
		"onedrive":     {"Storage", "Microsoft cloud storage"},
		"box":          {"Storage", "Enterprise content management"},

		// Payment & E-commerce
		"stripe":  {"Payment", "Online payment processing"},
		"shopify": {"E-commerce", "E-commerce platform"},
		"paypal":  {"Payment", "Digital payment platform"},
		"square":  {"Payment", "Payment and business tools"},

		// Data & Analytics
		"airtable":        {"Database", "Spreadsheet-database hybrid"},
		"google_sheets":   {"Spreadsheet", "Google cloud spreadsheets"},
		"tableau":         {"Analytics", "Data visualization platform"},
		"microsoft_excel": {"Spreadsheet", "Microsoft spreadsheet application"},

		// Social Media
		"twitter":   {"Social Media", "Social networking platform"},
		"linkedin":  {"Social Media", "Professional networking"},
		"facebook":  {"Social Media", "Social networking platform"},
		"instagram": {"Social Media", "Photo and video sharing"},
	}

	if data, ok := info[providerType]; ok {
		return data.category, data.description
	}
	return "Other", "Integration provider"
}

// extractUserID reads the authenticated user's ID from the request context,
// populated by middleware.Auth. If the token has not been validated yet (e.g.
// running without auth in development), it returns a well-known sentinel UUID
// so the service remains functional without panicking.
func extractUserID(r *http.Request) uuid.UUID {
	if raw, ok := r.Context().Value(middleware.ContextKeyUserID).(string); ok && raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			return id
		}
	}
	// Sentinel: used in dev/demo mode when the auth middleware is bypassed.
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

// Helper functions

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func respondError(w http.ResponseWriter, message string, status int) {
	respondJSON(w, map[string]string{"error": message}, status)
}
