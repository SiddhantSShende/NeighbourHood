package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Auth      AuthConfig
	Providers ProvidersConfig
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret   string
	GoogleOAuth OAuthConfig
	GitHubOAuth OAuthConfig
}

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Enabled      bool
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port string
	Env  string // development, staging, production
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ProvidersConfig holds all integration provider configurations
type ProvidersConfig struct {
	// Communication & Collaboration
	Slack          ProviderConfig
	MicrosoftTeams ProviderConfig
	Zoom           ProviderConfig
	Discord        ProviderConfig

	// Email & Marketing
	Gmail     ProviderConfig
	SendGrid  ProviderConfig
	Mailchimp ProviderConfig
	Twilio    ProviderConfig

	// Project Management
	Jira    ProviderConfig
	Trello  ProviderConfig
	Asana   ProviderConfig
	Monday  ProviderConfig
	Notion  ProviderConfig
	ClickUp ProviderConfig

	// CRM & Sales
	Salesforce ProviderConfig
	HubSpot    ProviderConfig
	Zendesk    ProviderConfig
	Intercom   ProviderConfig
	Pipedrive  ProviderConfig

	// Development & Code
	GitHub    ProviderConfig
	GitLab    ProviderConfig
	Bitbucket ProviderConfig

	// Storage & Documents
	Dropbox     ProviderConfig
	GoogleDrive ProviderConfig
	OneDrive    ProviderConfig
	Box         ProviderConfig

	// Payment & E-commerce
	Stripe  ProviderConfig
	Shopify ProviderConfig
	PayPal  ProviderConfig
	Square  ProviderConfig

	// Data & Analytics
	Airtable       ProviderConfig
	GoogleSheets   ProviderConfig
	Tableau        ProviderConfig
	MicrosoftExcel ProviderConfig

	// Social Media
	Twitter   ProviderConfig
	LinkedIn  ProviderConfig
	Facebook  ProviderConfig
	Instagram ProviderConfig
}

// ProviderConfig holds generic provider configuration
type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Enabled      bool
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-in-production"),
			GoogleOAuth: OAuthConfig{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
				Enabled:      getEnvBool("GOOGLE_AUTH_ENABLED", true),
			},
			GitHubOAuth: OAuthConfig{
				ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
				ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/auth/github/callback"),
				Enabled:      getEnvBool("GITHUB_AUTH_ENABLED", true),
			},
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "neighbourhood"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Providers: ProvidersConfig{
			// Communication & Collaboration
			Slack:          loadProvider("SLACK"),
			MicrosoftTeams: loadProvider("MICROSOFT_TEAMS"),
			Zoom:           loadProvider("ZOOM"),
			Discord:        loadProvider("DISCORD"),

			// Email & Marketing
			Gmail:     loadProvider("GMAIL"),
			SendGrid:  loadProvider("SENDGRID"),
			Mailchimp: loadProvider("MAILCHIMP"),
			Twilio:    loadProvider("TWILIO"),

			// Project Management
			Jira:    loadProvider("JIRA"),
			Trello:  loadProvider("TRELLO"),
			Asana:   loadProvider("ASANA"),
			Monday:  loadProvider("MONDAY"),
			Notion:  loadProvider("NOTION"),
			ClickUp: loadProvider("CLICKUP"),

			// CRM & Sales
			Salesforce: loadProvider("SALESFORCE"),
			HubSpot:    loadProvider("HUBSPOT"),
			Zendesk:    loadProvider("ZENDESK"),
			Intercom:   loadProvider("INTERCOM"),
			Pipedrive:  loadProvider("PIPEDRIVE"),

			// Development & Code
			GitHub:    loadProvider("GITHUB"),
			GitLab:    loadProvider("GITLAB"),
			Bitbucket: loadProvider("BITBUCKET"),

			// Storage & Documents
			Dropbox:     loadProvider("DROPBOX"),
			GoogleDrive: loadProvider("GOOGLE_DRIVE"),
			OneDrive:    loadProvider("ONEDRIVE"),
			Box:         loadProvider("BOX"),

			// Payment & E-commerce
			Stripe:  loadProvider("STRIPE"),
			Shopify: loadProvider("SHOPIFY"),
			PayPal:  loadProvider("PAYPAL"),
			Square:  loadProvider("SQUARE"),

			// Data & Analytics
			Airtable:       loadProvider("AIRTABLE"),
			GoogleSheets:   loadProvider("GOOGLE_SHEETS"),
			Tableau:        loadProvider("TABLEAU"),
			MicrosoftExcel: loadProvider("MICROSOFT_EXCEL"),

			// Social Media
			Twitter:   loadProvider("TWITTER"),
			LinkedIn:  loadProvider("LINKEDIN"),
			Facebook:  loadProvider("FACEBOOK"),
			Instagram: loadProvider("INSTAGRAM"),
		},
	}
}

// loadProvider loads a provider configuration from environment variables
func loadProvider(prefix string) ProviderConfig {
	return ProviderConfig{
		ClientID:     getEnv(prefix+"_CLIENT_ID", ""),
		ClientSecret: getEnv(prefix+"_CLIENT_SECRET", ""),
		RedirectURL:  getEnv(prefix+"_REDIRECT_URL", ""),
		Enabled:      getEnvBool(prefix+"_ENABLED", true),
	}
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
