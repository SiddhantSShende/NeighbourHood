package integrations

// IntegrationConfig holds configuration for each integration
type IntegrationConfig struct {
	Type         IntegrationType
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// List of supported integrations (populate from env/config in production)
var SupportedIntegrations = []IntegrationConfig{
	{Type: IntegrationSlack, ClientID: "", ClientSecret: "", RedirectURL: ""},
	{Type: IntegrationGmail, ClientID: "", ClientSecret: "", RedirectURL: ""},
	{Type: IntegrationJira, ClientID: "", ClientSecret: "", RedirectURL: ""},
}
