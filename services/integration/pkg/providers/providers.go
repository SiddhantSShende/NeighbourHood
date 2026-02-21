package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"neighbourhood/services/integration/internal/config"
)

// SlackProvider implements Slack integration
type SlackProvider struct {
	config config.ProviderConfig
}

func NewSlackProvider(cfg config.ProviderConfig) *SlackProvider {
	return &SlackProvider{config: cfg}
}

func (p *SlackProvider) ID() string       { return "slack" }
func (p *SlackProvider) Name() string     { return "Slack" }
func (p *SlackProvider) Category() string { return "Communication" }

func (p *SlackProvider) GetAuthURL(state string) string {
	scopes := strings.Join(p.config.Scopes, ",")
	return fmt.Sprintf("https://slack.com/oauth/v2/authorize?client_id=%s&scope=%s&state=%s&redirect_uri=%s",
		p.config.ClientID, scopes, state, url.QueryEscape(p.config.RedirectURL))
}

func (p *SlackProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	data := url.Values{}
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", p.config.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/oauth.v2.access", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: p.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK          bool   `json:"ok"`
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("slack oauth error: %s", result.Error)
	}

	return &Token{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
	}, nil
}

func (p *SlackProvider) Execute(ctx context.Context, token *Token, action string, params map[string]interface{}) (interface{}, error) {
	switch action {
	case "send_message":
		return p.sendMessage(ctx, token, params)
	case "list_channels":
		return p.listChannels(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (p *SlackProvider) sendMessage(ctx context.Context, token *Token, params map[string]interface{}) (interface{}, error) {
	channel, _ := params["channel"].(string)
	text, _ := params["text"].(string)

	if channel == "" || text == "" {
		return nil, fmt.Errorf("channel and text are required")
	}

	payload := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}

	data, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: p.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (p *SlackProvider) listChannels(ctx context.Context, token *Token) (interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://slack.com/api/conversations.list", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{Timeout: p.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

// GmailProvider implements Gmail integration
type GmailProvider struct {
	config config.ProviderConfig
}

func NewGmailProvider(cfg config.ProviderConfig) *GmailProvider {
	return &GmailProvider{config: cfg}
}

func (p *GmailProvider) ID() string       { return "gmail" }
func (p *GmailProvider) Name() string     { return "Gmail" }
func (p *GmailProvider) Category() string { return "Email" }

func (p *GmailProvider) GetAuthURL(state string) string {
	scopes := strings.Join(p.config.Scopes, " ")
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s&access_type=offline",
		p.config.ClientID, url.QueryEscape(p.config.RedirectURL), url.QueryEscape(scopes), state)
}

func (p *GmailProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	data := url.Values{}
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", p.config.RedirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: p.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &Token{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
	}, nil
}

func (p *GmailProvider) Execute(ctx context.Context, token *Token, action string, params map[string]interface{}) (interface{}, error) {
	switch action {
	case "send_email":
		return p.sendEmail(ctx, token, params)
	case "list_messages":
		return p.listMessages(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (p *GmailProvider) sendEmail(ctx context.Context, token *Token, params map[string]interface{}) (interface{}, error) {
	// Simplified email sending - production would use proper Gmail API
	return map[string]interface{}{"status": "sent"}, nil
}

func (p *GmailProvider) listMessages(ctx context.Context, token *Token) (interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://gmail.googleapis.com/gmail/v1/users/me/messages", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{Timeout: p.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

// JiraProvider implements Jira integration
type JiraProvider struct {
	config config.ProviderConfig
}

func NewJiraProvider(cfg config.ProviderConfig) *JiraProvider {
	return &JiraProvider{config: cfg}
}

func (p *JiraProvider) ID() string       { return "jira" }
func (p *JiraProvider) Name() string     { return "Jira" }
func (p *JiraProvider) Category() string { return "Project Management" }

func (p *JiraProvider) GetAuthURL(state string) string {
	scopes := strings.Join(p.config.Scopes, "%20")
	return fmt.Sprintf("https://auth.atlassian.com/authorize?audience=api.atlassian.com&client_id=%s&scope=%s&redirect_uri=%s&state=%s&response_type=code&prompt=consent",
		p.config.ClientID, scopes, url.QueryEscape(p.config.RedirectURL), state)
}

func (p *JiraProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", p.config.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://auth.atlassian.com/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: p.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &Token{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    "Bearer",
	}, nil
}

func (p *JiraProvider) Execute(ctx context.Context, token *Token, action string, params map[string]interface{}) (interface{}, error) {
	switch action {
	case "create_issue":
		return p.createIssue(ctx, token, params)
	case "list_issues":
		return p.listIssues(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (p *JiraProvider) createIssue(ctx context.Context, token *Token, params map[string]interface{}) (interface{}, error) {
	// Simplified - production would use proper Jira API
	return map[string]interface{}{"status": "created"}, nil
}

func (p *JiraProvider) listIssues(ctx context.Context, token *Token) (interface{}, error) {
	// Simplified - production would use proper Jira API
	return map[string]interface{}{"issues": []interface{}{}}, nil
}

// GitHubProvider implements GitHub integration
type GitHubProvider struct {
	config config.ProviderConfig
}

func NewGitHubProvider(cfg config.ProviderConfig) *GitHubProvider {
	return &GitHubProvider{config: cfg}
}

func (p *GitHubProvider) ID() string       { return "github" }
func (p *GitHubProvider) Name() string     { return "GitHub" }
func (p *GitHubProvider) Category() string { return "Development" }

func (p *GitHubProvider) GetAuthURL(state string) string {
	scopes := strings.Join(p.config.Scopes, " ")
	return fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		p.config.ClientID, url.QueryEscape(p.config.RedirectURL), scopes, state)
}

func (p *GitHubProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	data := url.Values{}
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", p.config.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: p.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &Token{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
	}, nil
}

func (p *GitHubProvider) Execute(ctx context.Context, token *Token, action string, params map[string]interface{}) (interface{}, error) {
	switch action {
	case "create_issue":
		return p.createIssue(ctx, token, params)
	case "list_repos":
		return p.listRepos(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (p *GitHubProvider) createIssue(ctx context.Context, token *Token, params map[string]interface{}) (interface{}, error) {
	repo, _ := params["repo"].(string)
	title, _ := params["title"].(string)

	if repo == "" || title == "" {
		return nil, fmt.Errorf("repo and title are required")
	}

	payload := map[string]interface{}{
		"title": title,
		"body":  params["body"],
	}

	data, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://api.github.com/repos/%s/issues", repo), strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: p.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (p *GitHubProvider) listRepos(ctx context.Context, token *Token) (interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/repos", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token.AccessToken)

	client := &http.Client{Timeout: p.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

// GenericProvider provides a generic OAuth implementation
type GenericProvider struct {
	id       string
	name     string
	category string
	config   config.ProviderConfig
}

func NewGenericProvider(name string, cfg config.ProviderConfig) *GenericProvider {
	return &GenericProvider{
		id:       name,
		name:     name,
		category: "Other",
		config:   cfg,
	}
}

func (p *GenericProvider) ID() string       { return p.id }
func (p *GenericProvider) Name() string     { return p.name }
func (p *GenericProvider) Category() string { return p.category }

func (p *GenericProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://oauth.example.com/authorize?client_id=%s&state=%s", p.config.ClientID, state)
}

func (p *GenericProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	return &Token{
		AccessToken: code,
		TokenType:   "Bearer",
	}, nil
}

func (p *GenericProvider) Execute(ctx context.Context, token *Token, action string, params map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{"status": "executed", "provider": p.id, "action": action}, nil
}
