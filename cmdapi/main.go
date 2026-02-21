package main

import (
	"log"
	"net/http"
	"os"

	"neighbourhood/internal/api"
	"neighbourhood/internal/auth"
	"neighbourhood/internal/config"
	"neighbourhood/internal/database"
	"neighbourhood/internal/integrations"
	"neighbourhood/internal/mcp"
	"neighbourhood/internal/middleware"
)

func main() {
	// 0. Load Configuration
	cfg := config.Load()
	log.Printf("Starting NeighbourHood Integration Platform in %s mode", cfg.Server.Env)

	// 1. Set Working Directory to Project Root
	if err := setProjectRoot(); err != nil {
		log.Printf("WARNING: Failed to set project root: %v", err)
	}

	// Verify critical file paths exist
	if _, err := os.Stat("./web/static"); os.IsNotExist(err) {
		wd, _ := os.Getwd()
		log.Fatalf("CRITICAL: ./web/static not found in %s. Check project structure.", wd)
	}
	if _, err := os.Stat("./web/templates/index.html"); os.IsNotExist(err) {
		wd, _ := os.Getwd()
		log.Fatalf("CRITICAL: ./web/templates/index.html not found in %s. Check project structure.", wd)
	}

	// 2. Initialize Database
	if err := database.InitDB(); err != nil {
		log.Printf("WARNING: Failed to initialize database: %v", err)
		log.Println("Server running in OFFLINE mode (No Database). Some features may be limited.")
	} else {
		defer database.DB.Close()

		// Run Migrations
		if err := database.RunMigrations(); err != nil {
			log.Printf("WARNING: Failed to run migrations: %v", err)
		}
	}

	// 3. Register Integration Providers
	registerProviders(cfg)

	// 4. Setup API Handler
	apiHandler := api.NewHandler()

	// 5. Setup OAuth Handler
	oauthHandler := auth.NewOAuthHandler(cfg)

	// 6. Setup Router
	mux := http.NewServeMux()

	// Health Check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Static Files
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Home/Dashboard
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./web/templates/index.html")
	})

	// Auth Routes
	mux.HandleFunc("/auth/login", auth.LoginHandler)

	// OAuth Routes
	mux.HandleFunc("/auth/google/login", oauthHandler.GoogleLoginHandler)
	mux.HandleFunc("/auth/google/callback", oauthHandler.GoogleCallbackHandler)
	mux.HandleFunc("/auth/github/login", oauthHandler.GitHubLoginHandler)
	mux.HandleFunc("/auth/github/callback", oauthHandler.GitHubCallbackHandler)

	// API Gateway routes for integrations and workflows
	mux.HandleFunc("/api/integrations", apiHandler.ListIntegrations)
	mux.HandleFunc("/api/integration/authurl", apiHandler.GetIntegrationAuthURL)
	mux.HandleFunc("/api/integration/execute", apiHandler.ExecuteIntegrationAction)
	mux.HandleFunc("/api/workflow/execute", apiHandler.ExecuteWorkflow)

	// MCP Routes
	mux.HandleFunc("/mcp", mcp.Handler)

	// 6. Apply Global Middleware
	handler := middleware.Chain(mux, middleware.Logger, middleware.CORS)

	// 7. Start Server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("Visit http://localhost:%s to access the developer portal", cfg.Server.Port)

	if err := http.ListenAndServe(":"+cfg.Server.Port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// registerProviders registers all integration providers
func registerProviders(cfg *config.Config) {
	// Communication & Collaboration
	if cfg.Providers.Slack.Enabled {
		integrations.RegisterProvider(integrations.NewSlackProvider(
			cfg.Providers.Slack.ClientID, cfg.Providers.Slack.ClientSecret, cfg.Providers.Slack.RedirectURL))
		log.Println("✓ Registered Slack provider")
	}
	if cfg.Providers.MicrosoftTeams.Enabled {
		integrations.RegisterProvider(integrations.NewMicrosoftTeamsProvider(
			cfg.Providers.MicrosoftTeams.ClientID, cfg.Providers.MicrosoftTeams.ClientSecret, cfg.Providers.MicrosoftTeams.RedirectURL))
		log.Println("✓ Registered Microsoft Teams provider")
	}
	if cfg.Providers.Zoom.Enabled {
		integrations.RegisterProvider(integrations.NewZoomProvider(
			cfg.Providers.Zoom.ClientID, cfg.Providers.Zoom.ClientSecret, cfg.Providers.Zoom.RedirectURL))
		log.Println("✓ Registered Zoom provider")
	}
	if cfg.Providers.Discord.Enabled {
		integrations.RegisterProvider(integrations.NewDiscordProvider(
			cfg.Providers.Discord.ClientID, cfg.Providers.Discord.ClientSecret, cfg.Providers.Discord.RedirectURL))
		log.Println("✓ Registered Discord provider")
	}

	// Email & Marketing
	if cfg.Providers.Gmail.Enabled {
		integrations.RegisterProvider(integrations.NewGmailProvider(
			cfg.Providers.Gmail.ClientID, cfg.Providers.Gmail.ClientSecret, cfg.Providers.Gmail.RedirectURL))
		log.Println("✓ Registered Gmail provider")
	}
	if cfg.Providers.SendGrid.Enabled {
		integrations.RegisterProvider(integrations.NewSendGridProvider(
			cfg.Providers.SendGrid.ClientID, cfg.Providers.SendGrid.ClientSecret, cfg.Providers.SendGrid.RedirectURL))
		log.Println("✓ Registered SendGrid provider")
	}
	if cfg.Providers.Mailchimp.Enabled {
		integrations.RegisterProvider(integrations.NewMailchimpProvider(
			cfg.Providers.Mailchimp.ClientID, cfg.Providers.Mailchimp.ClientSecret, cfg.Providers.Mailchimp.RedirectURL))
		log.Println("✓ Registered Mailchimp provider")
	}
	if cfg.Providers.Twilio.Enabled {
		integrations.RegisterProvider(integrations.NewTwilioProvider(
			cfg.Providers.Twilio.ClientID, cfg.Providers.Twilio.ClientSecret, cfg.Providers.Twilio.RedirectURL))
		log.Println("✓ Registered Twilio provider")
	}

	// Project Management
	if cfg.Providers.Jira.Enabled {
		integrations.RegisterProvider(integrations.NewJiraProvider(
			cfg.Providers.Jira.ClientID, cfg.Providers.Jira.ClientSecret, cfg.Providers.Jira.RedirectURL))
		log.Println("✓ Registered Jira provider")
	}
	if cfg.Providers.Trello.Enabled {
		integrations.RegisterProvider(integrations.NewTrelloProvider(
			cfg.Providers.Trello.ClientID, cfg.Providers.Trello.ClientSecret, cfg.Providers.Trello.RedirectURL))
		log.Println("✓ Registered Trello provider")
	}
	if cfg.Providers.Asana.Enabled {
		integrations.RegisterProvider(integrations.NewAsanaProvider(
			cfg.Providers.Asana.ClientID, cfg.Providers.Asana.ClientSecret, cfg.Providers.Asana.RedirectURL))
		log.Println("✓ Registered Asana provider")
	}
	if cfg.Providers.Monday.Enabled {
		integrations.RegisterProvider(integrations.NewMondayProvider(
			cfg.Providers.Monday.ClientID, cfg.Providers.Monday.ClientSecret, cfg.Providers.Monday.RedirectURL))
		log.Println("✓ Registered Monday.com provider")
	}
	if cfg.Providers.Notion.Enabled {
		integrations.RegisterProvider(integrations.NewNotionProvider(
			cfg.Providers.Notion.ClientID, cfg.Providers.Notion.ClientSecret, cfg.Providers.Notion.RedirectURL))
		log.Println("✓ Registered Notion provider")
	}
	if cfg.Providers.ClickUp.Enabled {
		integrations.RegisterProvider(integrations.NewClickUpProvider(
			cfg.Providers.ClickUp.ClientID, cfg.Providers.ClickUp.ClientSecret, cfg.Providers.ClickUp.RedirectURL))
		log.Println("✓ Registered ClickUp provider")
	}

	// CRM & Sales
	if cfg.Providers.Salesforce.Enabled {
		integrations.RegisterProvider(integrations.NewSalesforceProvider(
			cfg.Providers.Salesforce.ClientID, cfg.Providers.Salesforce.ClientSecret, cfg.Providers.Salesforce.RedirectURL))
		log.Println("✓ Registered Salesforce provider")
	}
	if cfg.Providers.HubSpot.Enabled {
		integrations.RegisterProvider(integrations.NewHubSpotProvider(
			cfg.Providers.HubSpot.ClientID, cfg.Providers.HubSpot.ClientSecret, cfg.Providers.HubSpot.RedirectURL))
		log.Println("✓ Registered HubSpot provider")
	}
	if cfg.Providers.Zendesk.Enabled {
		integrations.RegisterProvider(integrations.NewZendeskProvider(
			cfg.Providers.Zendesk.ClientID, cfg.Providers.Zendesk.ClientSecret, cfg.Providers.Zendesk.RedirectURL))
		log.Println("✓ Registered Zendesk provider")
	}
	if cfg.Providers.Intercom.Enabled {
		integrations.RegisterProvider(integrations.NewIntercomProvider(
			cfg.Providers.Intercom.ClientID, cfg.Providers.Intercom.ClientSecret, cfg.Providers.Intercom.RedirectURL))
		log.Println("✓ Registered Intercom provider")
	}
	if cfg.Providers.Pipedrive.Enabled {
		integrations.RegisterProvider(integrations.NewPipedriveProvider(
			cfg.Providers.Pipedrive.ClientID, cfg.Providers.Pipedrive.ClientSecret, cfg.Providers.Pipedrive.RedirectURL))
		log.Println("✓ Registered Pipedrive provider")
	}

	// Development & Code
	if cfg.Providers.GitHub.Enabled {
		integrations.RegisterProvider(integrations.NewGitHubProvider(
			cfg.Providers.GitHub.ClientID, cfg.Providers.GitHub.ClientSecret, cfg.Providers.GitHub.RedirectURL))
		log.Println("✓ Registered GitHub provider")
	}
	if cfg.Providers.GitLab.Enabled {
		integrations.RegisterProvider(integrations.NewGitLabProvider(
			cfg.Providers.GitLab.ClientID, cfg.Providers.GitLab.ClientSecret, cfg.Providers.GitLab.RedirectURL))
		log.Println("✓ Registered GitLab provider")
	}
	if cfg.Providers.Bitbucket.Enabled {
		integrations.RegisterProvider(integrations.NewBitbucketProvider(
			cfg.Providers.Bitbucket.ClientID, cfg.Providers.Bitbucket.ClientSecret, cfg.Providers.Bitbucket.RedirectURL))
		log.Println("✓ Registered Bitbucket provider")
	}

	// Storage & Documents
	if cfg.Providers.Dropbox.Enabled {
		integrations.RegisterProvider(integrations.NewDropboxProvider(
			cfg.Providers.Dropbox.ClientID, cfg.Providers.Dropbox.ClientSecret, cfg.Providers.Dropbox.RedirectURL))
		log.Println("✓ Registered Dropbox provider")
	}
	if cfg.Providers.GoogleDrive.Enabled {
		integrations.RegisterProvider(integrations.NewGoogleDriveProvider(
			cfg.Providers.GoogleDrive.ClientID, cfg.Providers.GoogleDrive.ClientSecret, cfg.Providers.GoogleDrive.RedirectURL))
		log.Println("✓ Registered Google Drive provider")
	}
	if cfg.Providers.OneDrive.Enabled {
		integrations.RegisterProvider(integrations.NewOneDriveProvider(
			cfg.Providers.OneDrive.ClientID, cfg.Providers.OneDrive.ClientSecret, cfg.Providers.OneDrive.RedirectURL))
		log.Println("✓ Registered OneDrive provider")
	}
	if cfg.Providers.Box.Enabled {
		integrations.RegisterProvider(integrations.NewBoxProvider(
			cfg.Providers.Box.ClientID, cfg.Providers.Box.ClientSecret, cfg.Providers.Box.RedirectURL))
		log.Println("✓ Registered Box provider")
	}

	// Payment & E-commerce
	if cfg.Providers.Stripe.Enabled {
		integrations.RegisterProvider(integrations.NewStripeProvider(
			cfg.Providers.Stripe.ClientID, cfg.Providers.Stripe.ClientSecret, cfg.Providers.Stripe.RedirectURL))
		log.Println("✓ Registered Stripe provider")
	}
	if cfg.Providers.Shopify.Enabled {
		integrations.RegisterProvider(integrations.NewShopifyProvider(
			cfg.Providers.Shopify.ClientID, cfg.Providers.Shopify.ClientSecret, cfg.Providers.Shopify.RedirectURL))
		log.Println("✓ Registered Shopify provider")
	}
	if cfg.Providers.PayPal.Enabled {
		integrations.RegisterProvider(integrations.NewPayPalProvider(
			cfg.Providers.PayPal.ClientID, cfg.Providers.PayPal.ClientSecret, cfg.Providers.PayPal.RedirectURL))
		log.Println("✓ Registered PayPal provider")
	}
	if cfg.Providers.Square.Enabled {
		integrations.RegisterProvider(integrations.NewSquareProvider(
			cfg.Providers.Square.ClientID, cfg.Providers.Square.ClientSecret, cfg.Providers.Square.RedirectURL))
		log.Println("✓ Registered Square provider")
	}

	// Data & Analytics
	if cfg.Providers.Airtable.Enabled {
		integrations.RegisterProvider(integrations.NewAirtableProvider(
			cfg.Providers.Airtable.ClientID, cfg.Providers.Airtable.ClientSecret, cfg.Providers.Airtable.RedirectURL))
		log.Println("✓ Registered Airtable provider")
	}
	if cfg.Providers.GoogleSheets.Enabled {
		integrations.RegisterProvider(integrations.NewGoogleSheetsProvider(
			cfg.Providers.GoogleSheets.ClientID, cfg.Providers.GoogleSheets.ClientSecret, cfg.Providers.GoogleSheets.RedirectURL))
		log.Println("✓ Registered Google Sheets provider")
	}
	if cfg.Providers.Tableau.Enabled {
		integrations.RegisterProvider(integrations.NewTableauProvider(
			cfg.Providers.Tableau.ClientID, cfg.Providers.Tableau.ClientSecret, cfg.Providers.Tableau.RedirectURL))
		log.Println("✓ Registered Tableau provider")
	}
	if cfg.Providers.MicrosoftExcel.Enabled {
		integrations.RegisterProvider(integrations.NewMicrosoftExcelProvider(
			cfg.Providers.MicrosoftExcel.ClientID, cfg.Providers.MicrosoftExcel.ClientSecret, cfg.Providers.MicrosoftExcel.RedirectURL))
		log.Println("✓ Registered Microsoft Excel provider")
	}

	// Social Media
	if cfg.Providers.Twitter.Enabled {
		integrations.RegisterProvider(integrations.NewTwitterProvider(
			cfg.Providers.Twitter.ClientID, cfg.Providers.Twitter.ClientSecret, cfg.Providers.Twitter.RedirectURL))
		log.Println("✓ Registered Twitter provider")
	}
	if cfg.Providers.LinkedIn.Enabled {
		integrations.RegisterProvider(integrations.NewLinkedInProvider(
			cfg.Providers.LinkedIn.ClientID, cfg.Providers.LinkedIn.ClientSecret, cfg.Providers.LinkedIn.RedirectURL))
		log.Println("✓ Registered LinkedIn provider")
	}
	if cfg.Providers.Facebook.Enabled {
		integrations.RegisterProvider(integrations.NewFacebookProvider(
			cfg.Providers.Facebook.ClientID, cfg.Providers.Facebook.ClientSecret, cfg.Providers.Facebook.RedirectURL))
		log.Println("✓ Registered Facebook provider")
	}
	if cfg.Providers.Instagram.Enabled {
		integrations.RegisterProvider(integrations.NewInstagramProvider(
			cfg.Providers.Instagram.ClientID, cfg.Providers.Instagram.ClientSecret, cfg.Providers.Instagram.RedirectURL))
		log.Println("✓ Registered Instagram provider")
	}

	log.Printf("Total providers registered: %d", len(integrations.Providers))
}

// setProjectRoot attempts to find the go.mod file and change the working directory to its location.
func setProjectRoot() error {
	_, err := os.Getwd()
	if err != nil {
		return err
	}

	// Check if go.mod exists in current directory
	if _, err := os.Stat("go.mod"); err == nil {
		return nil // Already at root
	}

	// Check if go.mod exists in parent directory (case for running from cmdapi/)
	if _, err := os.Stat("../go.mod"); err == nil {
		if err := os.Chdir(".."); err != nil {
			return err
		}
		log.Println("Changed working directory to project root")
		return nil
	}

	return os.ErrNotExist
}
