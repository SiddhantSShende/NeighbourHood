# NeighbourHood Integration Platform

A production-grade third-party integration platform that enables SaaS/PaaS providers to offer their services across multiple platforms with workflow orchestration capabilities.

🌐 **[Live Demo](https://siddhantssshende.github.io/NeighbourHood/)** | 📖 **[API Docs](API_DOCUMENTATION.md)** | 🚀 **[Quick Start](QUICKSTART.md)**

## 🚀 Features

- **Unified Integration Framework**: Single interface for all third-party integrations (Slack, Gmail, Jira, etc.)
- **Workflow Engine**: Orchestrate multi-step workflows across different providers
- **API Gateway**: Central entry point for all integration and workflow requests
- **Consent Management**: Track and validate user consent for data sharing
- **Developer Portal**: Web-based dashboard for managing integrations and workflows
- **Production-Ready**: Includes middleware, logging, error handling, and configuration management

## 📁 Project Structure

```
NeighbourHood/
├── cmdapi/
│   └── main.go                    # Application entry point
├── internal/
│   ├── api/
│   │   └── handlers.go            # HTTP request handlers
│   ├── auth/
│   │   └── auth.go                # Authentication logic
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── consent/
│   │   └── consent.go             # Consent management
│   ├── database/
│   │   └── db.go                  # Database operations
│   ├── integrations/
│   │   ├── integrations.go        # Provider implementations
│   │   └── integrations_model.go  # Integration configurations
│   ├── mcp/
│   │   ├── server.go              # MCP server
│   │   └── tools.go               # MCP tools
│   ├── middleware/
│   │   └── middleware.go          # HTTP middleware
│   ├── models/
│   │   ├── models.go              # Data models
│   │   └── schema.sql             # Database schema
│   └── workflow/
│       └── engine.go              # Workflow orchestration
├── services/                      # Microservices
│   ├── auth/                      # Authentication service
│   └── integration/               # Integration service
├── proto/                         # Protocol Buffers definitions
├── configs/                       # Service configurations
├── docs/                          # GitHub Pages frontend
│   ├── index.html                 # Landing page
│   ├── dashboard.html             # Developer dashboard
│   └── static/                    # CSS & JS assets
├── docker-compose.yml             # Docker orchestration
├── Makefile                       # Build automation
└── go.mod                         # Go dependencies
```

## 🏗️ Architecture

### Clean Code Principles

- **Separation of Concerns**: Each package has a single responsibility
- **Dependency Injection**: Configuration and dependencies are injected
- **Interface-Based Design**: Providers implement a common interface
- **Modular Structure**: Easy to add new integrations and features

### Core Components

1. **Integration Providers**: Modular interfaces for each SaaS platform
2. **API Gateway**: RESTful endpoints for integration management
3. **Workflow Engine**: Multi-step workflow orchestration
4. **Consent Manager**: Track and validate user consent
5. **Middleware Stack**: Logging, CORS, authentication, rate limiting

## 🛠️ Setup

### Prerequisites

- Go 1.20+
- PostgreSQL (optional, runs in offline mode without it)
- Docker & Docker Compose (optional)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd NeighbourHood
```

2. Copy the example environment file:
```bash
cp .env.example .env
```

3. Configure your environment variables in `.env`:
```env
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=neighbourhood
DB_SSL_MODE=disable

# Slack Integration
SLACK_CLIENT_ID=your-slack-client-id
SLACK_CLIENT_SECRET=your-slack-client-secret
SLACK_REDIRECT_URL=http://localhost:8080/callback/slack
SLACK_ENABLED=true

# Gmail Integration
GMAIL_CLIENT_ID=your-gmail-client-id
GMAIL_CLIENT_SECRET=your-gmail-client-secret
GMAIL_REDIRECT_URL=http://localhost:8080/callback/gmail
GMAIL_ENABLED=true

# Jira Integration
JIRA_CLIENT_ID=your-jira-client-id
JIRA_CLIENT_SECRET=your-jira-client-secret
JIRA_REDIRECT_URL=http://localhost:8080/callback/jira
JIRA_ENABLED=true
```

4. Install dependencies:
```bash
go mod download
```

5. Run the server:
```bash
cd cmdapi
go run main.go
```

6. Visit http://localhost:8080

### Docker Setup

```bash
docker-compose up
```

## 📚 API Documentation

### Get Integration Auth URL

```http
POST /api/integration/authurl
Content-Type: application/json

{
  "provider": "slack",
  "state": "random-state-string"
}
```

### Execute Integration Action

```http
POST /api/integration/execute
Content-Type: application/json

{
  "provider": "slack",
  "token": {
    "access_token": "xoxb-xxxx",
    "token_type": "Bearer"
  },
  "action": "send_message",
  "payload": {
    "channel": "#general",
    "text": "Hello World"
  }
}
```

### Execute Workflow

```http
POST /api/workflow/execute
Content-Type: application/json

{
  "workflow": {
    "id": "00000000-0000-0000-0000-000000000000",
    "name": "Email to Slack",
    "steps": [
      {
        "provider": "gmail",
        "action": "send_email",
        "payload": {
          "to": "user@example.com",
          "subject": "Test",
          "body": "Hello"
        }
      },
      {
        "provider": "slack",
        "action": "send_message",
        "payload": {
          "channel": "#alerts",
          "text": "Email sent!"
        }
      }
    ]
  },
  "tokens": {
    "gmail": {
      "access_token": "ya29.xxxx"
    },
    "slack": {
      "access_token": "xoxb-xxxx"
    }
  }
}
```

### List Integrations

```http
GET /api/integrations
```

## 🔌 Adding New Integrations

1. Add the integration type in `integrations.go`:
```go
const (
    IntegrationNewProvider IntegrationType = "newprovider"
)
```

2. Implement the Provider interface:
```go
type NewProvider struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
}

func (p *NewProvider) Name() string { return string(IntegrationNewProvider) }
func (p *NewProvider) GetAuthURL(state string) string { /* ... */ }
func (p *NewProvider) ExchangeCode(ctx context.Context, code string) (*Token, error) { /* ... */ }
func (p *NewProvider) Execute(ctx context.Context, token *Token, action string, payload map[string]interface{}) (interface{}, error) { /* ... */ }
```

3. Register it in `main.go`:
```go
newProvider := integrations.NewNewProvider(clientID, clientSecret, redirectURL)
integrations.RegisterProvider(newProvider)
```

## 🔐 Security Best Practices

- Store secrets in environment variables, never in code
- Use HTTPS in production
- Validate all user inputs
- Implement proper JWT authentication
- Use secure token storage
- Implement rate limiting
- Regular security audits

## 🤝 Integration with Consent Management System

The platform includes built-in consent management that can be integrated with external consent systems:

```go
consentManager := consent.NewManager()
// Integrate with your friend's consent API
consentManager.IntegrateFriendConsentSystem(apiURL, apiKey)
```

## 📝 License

[Private project - All rights reserved.]

## 👥 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.
OR SEND PULL REQUEST

## 📧 Contact

**Email**: siddhant.shende@consultrnr.com  
**GitHub**: [@SiddhantSShende](https://github.com/SiddhantSShende)

---

**Built with ❤️ for developers, by developers**
