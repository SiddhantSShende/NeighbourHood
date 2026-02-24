package integrations

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// IntegrationType is a string representing a supported integration
type IntegrationType string

const (
	// Communication & Collaboration
	IntegrationSlack          IntegrationType = "slack"
	IntegrationMicrosoftTeams IntegrationType = "microsoft_teams"
	IntegrationZoom           IntegrationType = "zoom"
	IntegrationDiscord        IntegrationType = "discord"

	// Email & Marketing
	IntegrationGmail     IntegrationType = "gmail"
	IntegrationSendGrid  IntegrationType = "sendgrid"
	IntegrationMailchimp IntegrationType = "mailchimp"
	IntegrationTwilio    IntegrationType = "twilio"

	// Project Management
	IntegrationJira    IntegrationType = "jira"
	IntegrationTrello  IntegrationType = "trello"
	IntegrationAsana   IntegrationType = "asana"
	IntegrationMonday  IntegrationType = "monday"
	IntegrationNotion  IntegrationType = "notion"
	IntegrationClickUp IntegrationType = "clickup"

	// CRM & Sales
	IntegrationSalesforce IntegrationType = "salesforce"
	IntegrationHubSpot    IntegrationType = "hubspot"
	IntegrationZendesk    IntegrationType = "zendesk"
	IntegrationIntercom   IntegrationType = "intercom"
	IntegrationPipedrive  IntegrationType = "pipedrive"

	// Development & Code
	IntegrationGitHub    IntegrationType = "github"
	IntegrationGitLab    IntegrationType = "gitlab"
	IntegrationBitbucket IntegrationType = "bitbucket"

	// Storage & Documents
	IntegrationDropbox     IntegrationType = "dropbox"
	IntegrationGoogleDrive IntegrationType = "google_drive"
	IntegrationOneDrive    IntegrationType = "onedrive"
	IntegrationBox         IntegrationType = "box"

	// Payment & E-commerce
	IntegrationStripe  IntegrationType = "stripe"
	IntegrationShopify IntegrationType = "shopify"
	IntegrationPayPal  IntegrationType = "paypal"
	IntegrationSquare  IntegrationType = "square"

	// Data & Analytics
	IntegrationAirtable       IntegrationType = "airtable"
	IntegrationGoogleSheets   IntegrationType = "google_sheets"
	IntegrationTableau        IntegrationType = "tableau"
	IntegrationMicrosoftExcel IntegrationType = "microsoft_excel"

	// Social Media
	IntegrationTwitter   IntegrationType = "twitter"
	IntegrationLinkedIn  IntegrationType = "linkedin"
	IntegrationFacebook  IntegrationType = "facebook"
	IntegrationInstagram IntegrationType = "instagram"
)

// Token holds OAuth or API token for a provider
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	Expiry       int64     `json:"expiry,omitempty"` // Unix timestamp for backward compatibility
}

// Provider is a generic interface for all integrations
// Implementations must be stateless and thread-safe.
type Provider interface {
	// Name returns the integration type (e.g., slack, gmail)
	Name() string
	// GetAuthURL returns the OAuth URL for user consent
	GetAuthURL(state string) string
	// ExchangeCode exchanges an OAuth code for a token
	ExchangeCode(ctx context.Context, code string) (*Token, error)
	// Execute performs an action with the provider
	Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error)
}

// Registry of all supported providers
var Providers = map[IntegrationType]Provider{}

// RegisterProvider adds a provider to the registry
// Call this in your main() or init() for each provider
func RegisterProvider(p Provider) {
	Providers[IntegrationType(p.Name())] = p
}

// GetProvider returns a provider by type
func GetProvider(t IntegrationType) (Provider, error) {
	p, ok := Providers[t]
	if !ok {
		return nil, errors.New("provider not found")
	}
	return p, nil
}

// getString safely extracts a required string value from a payload map.
// It returns an error if the key is absent or the value is not a string,
// preventing panics from unsafe type assertions in Execute methods.
func getString(payload map[string]interface{}, key string) (string, error) {
	val, ok := payload[key]
	if !ok {
		return "", fmt.Errorf("missing required field '%s'", key)
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("field '%s' must be a string, got %T", key, val)
	}
	return str, nil
}

// Example: Slack provider implementation (expand for Gmail, Jira, etc.)
type SlackProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewSlackProvider(clientID, clientSecret, redirectURL string) *SlackProvider {
	return &SlackProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

func (p *SlackProvider) Name() string { return string(IntegrationSlack) }
func (p *SlackProvider) GetAuthURL(state string) string {
	// In production, add scopes and state validation
	return fmt.Sprintf("https://slack.com/oauth/v2/authorize?client_id=%s&scope=chat:write&state=%s&redirect_uri=%s", p.ClientID, state, p.RedirectURL)
}
func (p *SlackProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	// TODO: Implement Slack OAuth exchange with HTTP client, handle errors
	// Mock implementation for now
	if code == "valid_code" {
		return &Token{
			AccessToken: "mock-slack-access-token",
			TokenType:   "Bearer",
		}, nil
	}
	return nil, errors.New("invalid authorization code")
}
func (p *SlackProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	// TODO: Implement Slack actions, validate token, handle rate limits
	if action == "send_message" {
		channel, ok := payload["channel"].(string)
		if !ok {
			return nil, errors.New("missing channel")
		}
		text, ok := payload["text"].(string)
		if !ok {
			return nil, errors.New("missing text")
		}
		// TODO: Implement actual Slack API call
		return map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Sent '%s' to %s", text, channel),
		}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// GmailProvider implements Provider interface for Gmail
type GmailProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewGmailProvider(clientID, clientSecret, redirectURL string) *GmailProvider {
	return &GmailProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

func (p *GmailProvider) Name() string { return string(IntegrationGmail) }
func (p *GmailProvider) GetAuthURL(state string) string {
	// Gmail OAuth URL with required scopes
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=https://www.googleapis.com/auth/gmail.send&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *GmailProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	// TODO: Implement Gmail OAuth exchange
	return nil, errors.New("gmail oauth exchange not implemented")
}
func (p *GmailProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "send_email" {
		to, ok := payload["to"].(string)
		if !ok {
			return nil, errors.New("missing 'to' field")
		}
		subject, ok := payload["subject"].(string)
		if !ok {
			return nil, errors.New("missing 'subject' field")
		}
		emailBody, ok := payload["body"].(string)
		if !ok {
			return nil, errors.New("missing 'body' field")
		}
		// TODO: Implement actual Gmail API call
		_ = emailBody // Suppress unused variable warning temporarily
		return map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Email sent to %s with subject '%s'", to, subject),
		}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// JiraProvider implements Provider interface for Jira
type JiraProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewJiraProvider(clientID, clientSecret, redirectURL string) *JiraProvider {
	return &JiraProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

func (p *JiraProvider) Name() string { return string(IntegrationJira) }
func (p *JiraProvider) GetAuthURL(state string) string {
	// Jira OAuth 2.0 URL
	return fmt.Sprintf("https://auth.atlassian.com/authorize?audience=api.atlassian.com&client_id=%s&scope=read:jira-work write:jira-work&redirect_uri=%s&state=%s&response_type=code&prompt=consent", p.ClientID, p.RedirectURL, state)
}
func (p *JiraProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	// TODO: Implement Jira OAuth exchange
	return nil, errors.New("jira oauth exchange not implemented")
}
func (p *JiraProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_issue" {
		project, ok := payload["project"].(string)
		if !ok {
			return nil, errors.New("missing 'project' field")
		}
		summary, ok := payload["summary"].(string)
		if !ok {
			return nil, errors.New("missing 'summary' field")
		}
		issueType, ok := payload["issue_type"].(string)
		if !ok {
			issueType = "Task"
		}
		// TODO: Implement actual Jira API call
		return map[string]string{
			"status":    "success",
			"message":   fmt.Sprintf("Created %s in project %s: %s", issueType, project, summary),
			"issue_key": "DEMO-123",
		}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ========== Communication & Collaboration Providers ==========

// MicrosoftTeamsProvider implements Provider interface for Microsoft Teams
type MicrosoftTeamsProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewMicrosoftTeamsProvider(clientID, clientSecret, redirectURL string) *MicrosoftTeamsProvider {
	return &MicrosoftTeamsProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *MicrosoftTeamsProvider) Name() string { return string(IntegrationMicrosoftTeams) }
func (p *MicrosoftTeamsProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://login.microsoftonline.com/common/oauth2/v2.0/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=https://graph.microsoft.com/User.Read https://graph.microsoft.com/ChannelMessage.Send&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *MicrosoftTeamsProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("microsoft teams oauth exchange not implemented")
}
func (p *MicrosoftTeamsProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "send_message" {
		channel, err := getString(payload, "channel")
		if err != nil {
			return nil, err
		}
		msg, err := getString(payload, "message")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "message": fmt.Sprintf("Sent message '%s' to Teams channel %s", msg, channel)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ZoomProvider implements Provider interface for Zoom
type ZoomProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewZoomProvider(clientID, clientSecret, redirectURL string) *ZoomProvider {
	return &ZoomProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *ZoomProvider) Name() string { return string(IntegrationZoom) }
func (p *ZoomProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://zoom.us/oauth/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *ZoomProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("zoom oauth exchange not implemented")
}
func (p *ZoomProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_meeting" {
		topic, err := getString(payload, "topic")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "meeting_url": "https://zoom.us/j/123456789", "topic": topic}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// DiscordProvider implements Provider interface for Discord
type DiscordProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewDiscordProvider(clientID, clientSecret, redirectURL string) *DiscordProvider {
	return &DiscordProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *DiscordProvider) Name() string { return string(IntegrationDiscord) }
func (p *DiscordProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify webhook.incoming&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *DiscordProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("discord oauth exchange not implemented")
}
func (p *DiscordProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "send_message" {
		channel, err := getString(payload, "channel")
		if err != nil {
			return nil, err
		}
		content, err := getString(payload, "content")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "message": fmt.Sprintf("Sent message '%s' to Discord channel %s", content, channel)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ========== Email & Marketing Providers ==========

// SendGridProvider implements Provider interface for SendGrid
type SendGridProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewSendGridProvider(clientID, clientSecret, redirectURL string) *SendGridProvider {
	return &SendGridProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *SendGridProvider) Name() string { return string(IntegrationSendGrid) }
func (p *SendGridProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://sendgrid.com/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *SendGridProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("sendgrid oauth exchange not implemented")
}
func (p *SendGridProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "send_email" {
		to, err := getString(payload, "to")
		if err != nil {
			return nil, err
		}
		subject, err := getString(payload, "subject")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "message": fmt.Sprintf("Email sent via SendGrid to %s with subject '%s'", to, subject)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// MailchimpProvider implements Provider interface for Mailchimp
type MailchimpProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewMailchimpProvider(clientID, clientSecret, redirectURL string) *MailchimpProvider {
	return &MailchimpProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *MailchimpProvider) Name() string { return string(IntegrationMailchimp) }
func (p *MailchimpProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://login.mailchimp.com/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *MailchimpProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("mailchimp oauth exchange not implemented")
}
func (p *MailchimpProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "add_subscriber" {
		email, err := getString(payload, "email")
		if err != nil {
			return nil, err
		}
		listID, err := getString(payload, "list_id")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "message": fmt.Sprintf("Added %s to list %s", email, listID)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// TwilioProvider implements Provider interface for Twilio
type TwilioProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewTwilioProvider(clientID, clientSecret, redirectURL string) *TwilioProvider {
	return &TwilioProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *TwilioProvider) Name() string { return string(IntegrationTwilio) }
func (p *TwilioProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://www.twilio.com/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *TwilioProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("twilio oauth exchange not implemented")
}
func (p *TwilioProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "send_sms" {
		to, err := getString(payload, "to")
		if err != nil {
			return nil, err
		}
		body, err := getString(payload, "body")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "message": fmt.Sprintf("SMS sent to %s: %s", to, body)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ========== Project Management Providers ==========

// TrelloProvider implements Provider interface for Trello
type TrelloProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewTrelloProvider(clientID, clientSecret, redirectURL string) *TrelloProvider {
	return &TrelloProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *TrelloProvider) Name() string { return string(IntegrationTrello) }
func (p *TrelloProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://trello.com/1/authorize?expiration=never&name=NeighbourHood&scope=read,write&response_type=token&key=%s&callback_method=fragment&return_url=%s&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *TrelloProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("trello oauth exchange not implemented")
}
func (p *TrelloProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_card" {
		listID, err := getString(payload, "list_id")
		if err != nil {
			return nil, err
		}
		name, err := getString(payload, "name")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "card_id": "abc123", "message": fmt.Sprintf("Created card '%s' in list %s", name, listID)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// AsanaProvider implements Provider interface for Asana
type AsanaProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewAsanaProvider(clientID, clientSecret, redirectURL string) *AsanaProvider {
	return &AsanaProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *AsanaProvider) Name() string { return string(IntegrationAsana) }
func (p *AsanaProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://app.asana.com/-/oauth_authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *AsanaProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("asana oauth exchange not implemented")
}
func (p *AsanaProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_task" {
		project, err := getString(payload, "project")
		if err != nil {
			return nil, err
		}
		name, err := getString(payload, "name")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "task_gid": "1234567890", "message": fmt.Sprintf("Created task '%s' in project %s", name, project)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// MondayProvider implements Provider interface for Monday.com
type MondayProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewMondayProvider(clientID, clientSecret, redirectURL string) *MondayProvider {
	return &MondayProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *MondayProvider) Name() string { return string(IntegrationMonday) }
func (p *MondayProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://auth.monday.com/oauth2/authorize?client_id=%s&redirect_uri=%s&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *MondayProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("monday oauth exchange not implemented")
}
func (p *MondayProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_item" {
		board, err := getString(payload, "board_id")
		if err != nil {
			return nil, err
		}
		name, err := getString(payload, "name")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "item_id": "123456", "message": fmt.Sprintf("Created item '%s' in board %s", name, board)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// NotionProvider implements Provider interface for Notion
type NotionProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewNotionProvider(clientID, clientSecret, redirectURL string) *NotionProvider {
	return &NotionProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *NotionProvider) Name() string { return string(IntegrationNotion) }
func (p *NotionProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://api.notion.com/v1/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&owner=user&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *NotionProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("notion oauth exchange not implemented")
}
func (p *NotionProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_page" {
		parent, err := getString(payload, "parent_id")
		if err != nil {
			return nil, err
		}
		title, err := getString(payload, "title")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "page_id": "abc-123", "message": fmt.Sprintf("Created page '%s' in %s", title, parent)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ClickUpProvider implements Provider interface for ClickUp
type ClickUpProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewClickUpProvider(clientID, clientSecret, redirectURL string) *ClickUpProvider {
	return &ClickUpProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *ClickUpProvider) Name() string { return string(IntegrationClickUp) }
func (p *ClickUpProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://app.clickup.com/api?client_id=%s&redirect_uri=%s&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *ClickUpProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("clickup oauth exchange not implemented")
}
func (p *ClickUpProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_task" {
		listID, err := getString(payload, "list_id")
		if err != nil {
			return nil, err
		}
		name, err := getString(payload, "name")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "task_id": "xyz789", "message": fmt.Sprintf("Created task '%s' in list %s", name, listID)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ========== CRM & Sales Providers ==========

// SalesforceProvider implements Provider interface for Salesforce
type SalesforceProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewSalesforceProvider(clientID, clientSecret, redirectURL string) *SalesforceProvider {
	return &SalesforceProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *SalesforceProvider) Name() string { return string(IntegrationSalesforce) }
func (p *SalesforceProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://login.salesforce.com/services/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *SalesforceProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("salesforce oauth exchange not implemented")
}
func (p *SalesforceProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_lead" {
		firstName, err := getString(payload, "first_name")
		if err != nil {
			return nil, err
		}
		lastName, err := getString(payload, "last_name")
		if err != nil {
			return nil, err
		}
		company, err := getString(payload, "company")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "lead_id": "00Q123456", "message": fmt.Sprintf("Created lead for %s %s at %s", firstName, lastName, company)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// HubSpotProvider implements Provider interface for HubSpot
type HubSpotProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewHubSpotProvider(clientID, clientSecret, redirectURL string) *HubSpotProvider {
	return &HubSpotProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *HubSpotProvider) Name() string { return string(IntegrationHubSpot) }
func (p *HubSpotProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://app.hubspot.com/oauth/authorize?client_id=%s&redirect_uri=%s&scope=contacts&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *HubSpotProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("hubspot oauth exchange not implemented")
}
func (p *HubSpotProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_contact" {
		email, err := getString(payload, "email")
		if err != nil {
			return nil, err
		}
		firstName, err := getString(payload, "first_name")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "contact_id": "12345", "message": fmt.Sprintf("Created contact for %s (%s)", firstName, email)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ZendeskProvider implements Provider interface for Zendesk
type ZendeskProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewZendeskProvider(clientID, clientSecret, redirectURL string) *ZendeskProvider {
	return &ZendeskProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *ZendeskProvider) Name() string { return string(IntegrationZendesk) }
func (p *ZendeskProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://your-domain.zendesk.com/oauth/authorizations/new?response_type=code&redirect_uri=%s&client_id=%s&scope=read write&state=%s", p.RedirectURL, p.ClientID, state)
}
func (p *ZendeskProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("zendesk oauth exchange not implemented")
}
func (p *ZendeskProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_ticket" {
		subject, err := getString(payload, "subject")
		if err != nil {
			return nil, err
		}
		desc, err := getString(payload, "description")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "ticket_id": "1234", "message": fmt.Sprintf("Created ticket: %s - %s", subject, desc)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// IntercomProvider implements Provider interface for Intercom
type IntercomProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewIntercomProvider(clientID, clientSecret, redirectURL string) *IntercomProvider {
	return &IntercomProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *IntercomProvider) Name() string { return string(IntegrationIntercom) }
func (p *IntercomProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://app.intercom.com/oauth?client_id=%s&redirect_uri=%s&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *IntercomProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("intercom oauth exchange not implemented")
}
func (p *IntercomProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_user" {
		email, err := getString(payload, "email")
		if err != nil {
			return nil, err
		}
		name, err := getString(payload, "name")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "user_id": "abc123", "message": fmt.Sprintf("Created user %s (%s)", name, email)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// PipedriveProvider implements Provider interface for Pipedrive
type PipedriveProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewPipedriveProvider(clientID, clientSecret, redirectURL string) *PipedriveProvider {
	return &PipedriveProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *PipedriveProvider) Name() string { return string(IntegrationPipedrive) }
func (p *PipedriveProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://oauth.pipedrive.com/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *PipedriveProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("pipedrive oauth exchange not implemented")
}
func (p *PipedriveProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_deal" {
		title, err := getString(payload, "title")
		if err != nil {
			return nil, err
		}
		value := payload["value"] // numeric — kept as interface{}
		return map[string]interface{}{"status": "success", "deal_id": 123, "message": fmt.Sprintf("Created deal '%s' worth %v", title, value)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ========== Development & Code Providers ==========

// GitHubProvider implements Provider interface for GitHub
type GitHubProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewGitHubProvider(clientID, clientSecret, redirectURL string) *GitHubProvider {
	return &GitHubProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *GitHubProvider) Name() string { return string(IntegrationGitHub) }
func (p *GitHubProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=repo user&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *GitHubProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("github oauth exchange not implemented")
}
func (p *GitHubProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_issue" {
		repo, err := getString(payload, "repo")
		if err != nil {
			return nil, err
		}
		title, err := getString(payload, "title")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "issue_number": "42", "message": fmt.Sprintf("Created issue in %s: %s", repo, title)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// GitLabProvider implements Provider interface for GitLab
type GitLabProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewGitLabProvider(clientID, clientSecret, redirectURL string) *GitLabProvider {
	return &GitLabProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *GitLabProvider) Name() string { return string(IntegrationGitLab) }
func (p *GitLabProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://gitlab.com/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s&scope=api", p.ClientID, p.RedirectURL, state)
}
func (p *GitLabProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("gitlab oauth exchange not implemented")
}
func (p *GitLabProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_issue" {
		project, err := getString(payload, "project")
		if err != nil {
			return nil, err
		}
		title, err := getString(payload, "title")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "issue_iid": "123", "message": fmt.Sprintf("Created issue in %s: %s", project, title)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// BitbucketProvider implements Provider interface for Bitbucket
type BitbucketProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewBitbucketProvider(clientID, clientSecret, redirectURL string) *BitbucketProvider {
	return &BitbucketProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *BitbucketProvider) Name() string { return string(IntegrationBitbucket) }
func (p *BitbucketProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://bitbucket.org/site/oauth2/authorize?client_id=%s&response_type=code&state=%s", p.ClientID, state)
}
func (p *BitbucketProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("bitbucket oauth exchange not implemented")
}
func (p *BitbucketProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_pull_request" {
		repo, err := getString(payload, "repo")
		if err != nil {
			return nil, err
		}
		title, err := getString(payload, "title")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "pr_id": "99", "message": fmt.Sprintf("Created PR in %s: %s", repo, title)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ========== Storage & Documents Providers ==========

// DropboxProvider implements Provider interface for Dropbox
type DropboxProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewDropboxProvider(clientID, clientSecret, redirectURL string) *DropboxProvider {
	return &DropboxProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *DropboxProvider) Name() string { return string(IntegrationDropbox) }
func (p *DropboxProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://www.dropbox.com/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *DropboxProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("dropbox oauth exchange not implemented")
}
func (p *DropboxProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "upload_file" {
		path, err := getString(payload, "path")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "file_id": "id:abc123", "message": fmt.Sprintf("Uploaded file to %s", path)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// GoogleDriveProvider implements Provider interface for Google Drive
type GoogleDriveProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewGoogleDriveProvider(clientID, clientSecret, redirectURL string) *GoogleDriveProvider {
	return &GoogleDriveProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *GoogleDriveProvider) Name() string { return string(IntegrationGoogleDrive) }
func (p *GoogleDriveProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=https://www.googleapis.com/auth/drive.file&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *GoogleDriveProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("google drive oauth exchange not implemented")
}
func (p *GoogleDriveProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_file" {
		name, err := getString(payload, "name")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "file_id": "1aBcDeFgHiJkLmN", "message": fmt.Sprintf("Created file '%s'", name)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// OneDriveProvider implements Provider interface for OneDrive
type OneDriveProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewOneDriveProvider(clientID, clientSecret, redirectURL string) *OneDriveProvider {
	return &OneDriveProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *OneDriveProvider) Name() string { return string(IntegrationOneDrive) }
func (p *OneDriveProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://login.microsoftonline.com/common/oauth2/v2.0/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=Files.ReadWrite&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *OneDriveProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("onedrive oauth exchange not implemented")
}
func (p *OneDriveProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "upload_file" {
		fileName, err := getString(payload, "file_name")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "file_id": "abc-123-def", "message": fmt.Sprintf("Uploaded file '%s'", fileName)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// BoxProvider implements Provider interface for Box
type BoxProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewBoxProvider(clientID, clientSecret, redirectURL string) *BoxProvider {
	return &BoxProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *BoxProvider) Name() string { return string(IntegrationBox) }
func (p *BoxProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://account.box.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *BoxProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("box oauth exchange not implemented")
}
func (p *BoxProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "upload_file" {
		folderID, err := getString(payload, "folder_id")
		if err != nil {
			return nil, err
		}
		fileName, err := getString(payload, "file_name")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "file_id": "123456789", "message": fmt.Sprintf("Uploaded '%s' to folder %s", fileName, folderID)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ========== Payment & E-commerce Providers ==========

// StripeProvider implements Provider interface for Stripe
type StripeProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewStripeProvider(clientID, clientSecret, redirectURL string) *StripeProvider {
	return &StripeProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *StripeProvider) Name() string { return string(IntegrationStripe) }
func (p *StripeProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://connect.stripe.com/oauth/authorize?response_type=code&client_id=%s&scope=read_write&redirect_uri=%s&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *StripeProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("stripe oauth exchange not implemented")
}
func (p *StripeProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_payment_intent" {
		amount := payload["amount"] // numeric — kept as interface{}
		currency, err := getString(payload, "currency")
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "success", "payment_intent_id": "pi_123abc", "amount": amount, "currency": currency}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ShopifyProvider implements Provider interface for Shopify
type ShopifyProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewShopifyProvider(clientID, clientSecret, redirectURL string) *ShopifyProvider {
	return &ShopifyProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *ShopifyProvider) Name() string { return string(IntegrationShopify) }
func (p *ShopifyProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://{shop}.myshopify.com/admin/oauth/authorize?client_id=%s&scope=read_products,write_products&redirect_uri=%s&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *ShopifyProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("shopify oauth exchange not implemented")
}
func (p *ShopifyProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_product" {
		title, err := getString(payload, "title")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "product_id": "1234567890", "message": fmt.Sprintf("Created product '%s'", title)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// PayPalProvider implements Provider interface for PayPal
type PayPalProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewPayPalProvider(clientID, clientSecret, redirectURL string) *PayPalProvider {
	return &PayPalProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *PayPalProvider) Name() string { return string(IntegrationPayPal) }
func (p *PayPalProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://www.paypal.com/connect?flowEntry=static&client_id=%s&redirect_uri=%s&scope=openid profile email&response_type=code&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *PayPalProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("paypal oauth exchange not implemented")
}
func (p *PayPalProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_payment" {
		amount := payload["amount"]
		return map[string]interface{}{"status": "success", "payment_id": "PAY-123ABC", "amount": amount}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// SquareProvider implements Provider interface for Square
type SquareProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewSquareProvider(clientID, clientSecret, redirectURL string) *SquareProvider {
	return &SquareProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *SquareProvider) Name() string { return string(IntegrationSquare) }
func (p *SquareProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://connect.squareup.com/oauth2/authorize?client_id=%s&redirect_uri=%s&scope=PAYMENTS_READ PAYMENTS_WRITE&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *SquareProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("square oauth exchange not implemented")
}
func (p *SquareProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_payment" {
		amount := payload["amount"]
		return map[string]interface{}{"status": "success", "payment_id": "sq0abc123", "amount": amount}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ========== Data & Analytics Providers ==========

// AirtableProvider implements Provider interface for Airtable
type AirtableProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewAirtableProvider(clientID, clientSecret, redirectURL string) *AirtableProvider {
	return &AirtableProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *AirtableProvider) Name() string { return string(IntegrationAirtable) }
func (p *AirtableProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://airtable.com/oauth2/v1/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *AirtableProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("airtable oauth exchange not implemented")
}
func (p *AirtableProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "create_record" {
		table, err := getString(payload, "table")
		if err != nil {
			return nil, err
		}
		fields := payload["fields"] // arbitrary object — kept as interface{}
		return map[string]interface{}{"status": "success", "record_id": "recABC123", "message": fmt.Sprintf("Created record in table %s", table), "fields": fields}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// GoogleSheetsProvider implements Provider interface for Google Sheets
type GoogleSheetsProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewGoogleSheetsProvider(clientID, clientSecret, redirectURL string) *GoogleSheetsProvider {
	return &GoogleSheetsProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *GoogleSheetsProvider) Name() string { return string(IntegrationGoogleSheets) }
func (p *GoogleSheetsProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=https://www.googleapis.com/auth/spreadsheets&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *GoogleSheetsProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("google sheets oauth exchange not implemented")
}
func (p *GoogleSheetsProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "append_row" {
		spreadsheetID, err := getString(payload, "spreadsheet_id")
		if err != nil {
			return nil, err
		}
		values := payload["values"] // array — kept as interface{}
		return map[string]interface{}{"status": "success", "spreadsheet_id": spreadsheetID, "message": "Row appended successfully", "values": values}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// TableauProvider implements Provider interface for Tableau
type TableauProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewTableauProvider(clientID, clientSecret, redirectURL string) *TableauProvider {
	return &TableauProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *TableauProvider) Name() string { return string(IntegrationTableau) }
func (p *TableauProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://online.tableau.com/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *TableauProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("tableau oauth exchange not implemented")
}
func (p *TableauProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "refresh_datasource" {
		datasourceID, err := getString(payload, "datasource_id")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "datasource_id": datasourceID, "message": "Datasource refresh initiated"}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// MicrosoftExcelProvider implements Provider interface for Microsoft Excel
type MicrosoftExcelProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewMicrosoftExcelProvider(clientID, clientSecret, redirectURL string) *MicrosoftExcelProvider {
	return &MicrosoftExcelProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *MicrosoftExcelProvider) Name() string { return string(IntegrationMicrosoftExcel) }
func (p *MicrosoftExcelProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://login.microsoftonline.com/common/oauth2/v2.0/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=Files.ReadWrite&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *MicrosoftExcelProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("microsoft excel oauth exchange not implemented")
}
func (p *MicrosoftExcelProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "update_cell" {
		workbookID, err := getString(payload, "workbook_id")
		if err != nil {
			return nil, err
		}
		cell, err := getString(payload, "cell")
		if err != nil {
			return nil, err
		}
		value := payload["value"] // may be any scalar — kept as interface{}
		return map[string]interface{}{"status": "success", "workbook_id": workbookID, "cell": cell, "value": value, "message": "Cell updated successfully"}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// ========== Social Media Providers ==========

// TwitterProvider implements Provider interface for Twitter
type TwitterProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewTwitterProvider(clientID, clientSecret, redirectURL string) *TwitterProvider {
	return &TwitterProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *TwitterProvider) Name() string { return string(IntegrationTwitter) }
func (p *TwitterProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://twitter.com/i/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=tweet.read tweet.write users.read&state=%s&code_challenge=challenge&code_challenge_method=plain", p.ClientID, p.RedirectURL, state)
}
func (p *TwitterProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("twitter oauth exchange not implemented")
}
func (p *TwitterProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "post_tweet" {
		text, err := getString(payload, "text")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "tweet_id": "1234567890", "message": fmt.Sprintf("Posted tweet: %s", text)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// LinkedInProvider implements Provider interface for LinkedIn
type LinkedInProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewLinkedInProvider(clientID, clientSecret, redirectURL string) *LinkedInProvider {
	return &LinkedInProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *LinkedInProvider) Name() string { return string(IntegrationLinkedIn) }
func (p *LinkedInProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://www.linkedin.com/oauth/v2/authorization?response_type=code&client_id=%s&redirect_uri=%s&scope=r_liteprofile w_member_social&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *LinkedInProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("linkedin oauth exchange not implemented")
}
func (p *LinkedInProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "share_post" {
		text, err := getString(payload, "text")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "post_id": "urn:li:share:123", "message": fmt.Sprintf("Posted to LinkedIn: %s", text)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// FacebookProvider implements Provider interface for Facebook
type FacebookProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewFacebookProvider(clientID, clientSecret, redirectURL string) *FacebookProvider {
	return &FacebookProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *FacebookProvider) Name() string { return string(IntegrationFacebook) }
func (p *FacebookProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://www.facebook.com/v12.0/dialog/oauth?client_id=%s&redirect_uri=%s&scope=email,public_profile,pages_manage_posts&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *FacebookProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("facebook oauth exchange not implemented")
}
func (p *FacebookProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "publish_post" {
		message, err := getString(payload, "message")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "post_id": "123456789_987654321", "message": fmt.Sprintf("Published to Facebook: %s", message)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// InstagramProvider implements Provider interface for Instagram
type InstagramProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewInstagramProvider(clientID, clientSecret, redirectURL string) *InstagramProvider {
	return &InstagramProvider{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL}
}

func (p *InstagramProvider) Name() string { return string(IntegrationInstagram) }
func (p *InstagramProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://api.instagram.com/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user_profile,user_media&response_type=code&state=%s", p.ClientID, p.RedirectURL, state)
}
func (p *InstagramProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return nil, errors.New("instagram oauth exchange not implemented")
}
func (p *InstagramProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) {
	if action == "publish_media" {
		imgURL, err := getString(payload, "image_url")
		if err != nil {
			return nil, err
		}
		caption, err := getString(payload, "caption")
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "success", "media_id": "12345_67890", "message": fmt.Sprintf("Published %s to Instagram with caption: %s", imgURL, caption)}, nil
	}
	return nil, fmt.Errorf("unknown action: %s", action)
}

// Add similar structs for GmailProvider, JiraProvider, etc.
// In production, use secure storage for secrets and tokens.
