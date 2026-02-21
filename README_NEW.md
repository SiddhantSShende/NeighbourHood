# NeighbourHood - Enterprise Workflow Automation Platform

> **Production-grade B2B2C workflow automation platform** enabling developers to build powerful integration workflows for their end users.

[![License](https://img.shields.io/badge/license-Proprietary-red.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Architecture](https://img.shields.io/badge/architecture-microservices-green.svg)](./ARCHITECTURE.md)

## 🎯 What is NeighbourHood?

NeighbourHood is a **multi-tenant workflow automation platform** that enables SaaS developers to offer integration capabilities to their customers without building integration infrastructure themselves.

### The Problem We Solve

Building integrations is hard:
- 45+ different OAuth implementations
- Rate limiting and retry logic
- Token refresh and credential storage
- Workflow orchestration
- Multi-tenant isolation
- Compliance (GDPR, SOC2, HIPAA)

### Our Solution

**You build workflows. Your users connect accounts. We handle everything else.**

```
Developers (You)          →  Build workflows with our APIs
Your Customers            →  Authorize their accounts (OAuth)
NeighbourHood Platform    →  Execute workflows securely
```

## 🏗️ Architecture: Microservices-First

```
┌──────────────────────────────────────────────────────────────────┐
│                    API Gateway (Port 8080)                       │
│                         REST/HTTP                                │
└──────────────────────────────────────────────────────────────────┘
                               │
                ┌──────────────┴──────────────┐
                │         gRPC Layer          │
                └──────────────┬──────────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
┌───────▼───────┐    ┌────────▼────────┐    ┌───────▼────────┐
│  Auth Service │    │ Integration Svc │    │ Workflow Svc   │
│   Port 50051  │    │   Port 50052    │    │  Port 50053    │
└───────┬───────┘    └────────┬────────┘    └───────┬────────┘
        │                     │                      │
┌───────▼───────┐    ┌────────▼────────┐    ┌───────▼────────┐
│  Consent Svc  │    │Notification Svc │    │  Monitoring    │
│   Port 50054  │    │   Port 50055    │   │   (Prometheus)  │
└───────────────┘    └─────────────────┘    └────────────────┘

Database Layer:
├── PostgreSQL (Auth, Integration, Workflow, Consent DBs)
├── Redis (Sessions, Cache, Rate Limiting)
└── MongoDB (Notifications, Logs)

Observability:
├── Prometheus (Metrics) - Port 9090
├── Jaeger (Tracing) - Port 16where686
└── Grafana (Dashboards) - Port 3000
```

## ⚡ Quick Start

### Prerequisites

- **Go 1.21+**
- **Docker & Docker Compose**
- **Protocol Buffers Compiler** (`protoc`)
- **Make**

### Installation

```bash
# 1. Install tools
make install-tools

# 2. Start infrastructure (PostgreSQL, Redis, Jaeger, etc.)
docker-compose up -d

# 3. Generate protobuf code
make proto

# 4. Build all services
make build-all

# 5. Run services
make run-all           # In separate terminals, or:
make run-auth &        # Auth service
make run-integration & # Integration service
# ... etc
```

### Your First Workflow (For Developers)

#### Step 1: Create a Developer Workspace

```bash
curl -X POST http://localhost:8080/api/v1/workspaces \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "email": "dev@acme.com",
    "plan": "professional"
  }'

# Response:
# {
#   "workspace_id": "ws_abc123",
#   "api_key": "nhood_sk_live_...",
#   "webhook_secret": "whsec_..."
# }
```

#### Step 2: Create a Workflow

```bash
curl -X POST http://localhost:8080/api/v1/workspaces/ws_abc123/workflows \
  -H "Authorization: Bearer nhood_sk_live_..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Welcome New Customer",
    "trigger": {
      "type": "webhook",
      "path": "/new-customer"
    },
    "steps": [
      {
        "provider": "slack",
        "action": "send_message",
        "config": {
          "channel": "#sales",
          "text": "New customer: {{customer.name}}"
        }
      },
      {
        "provider": "salesforce",
        "action": "create_lead",
        "config": {
          "first_name": "{{customer.first_name}}",
          "last_name": "{{customer.last_name}}",
          "email": "{{customer.email}}"
        }
      }
    ]
  }'
```

#### Step 3: Your Customer Authorizes Integrations

Provide your customers with authentication links:

```javascript
// Your application code
const authUrl = await fetch('http://localhost:8080/api/v1/integrations/auth-url', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer nhood_sk_live_...',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    provider: 'slack',
    user_id: 'user_xyz789',  // Your customer's ID
    workspace_id: 'ws_abc123'
  })
}).then(r => r.json());

// Redirect your customer to authUrl.url
// After OAuth, workflow is ready to execute!
```

#### Step 4: Trigger Workflows

```bash
# Via webhook (triggered by your app)
curl -X POST http://localhost:8080/webhooks/ws_abc123/new-customer \
  -H "Content-Type: application/json" \
  -d '{
    "customer": {
      "name": "John Doe",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john@example.com"
    }
  }'

# Workflow executes automatically using customer's connected accounts!
```

## 📦 Supported Integrations (45+)

### ✅ Communication & Collaboration (9)
- **Slack** - Messages, channels, users, files
- **Microsoft Teams** - Messages, channels, meetings
- **Discord** - Messages, servers, roles
- **Zoom** - Meetings, webinars, recordings
- **Google Meet** - Meetings, recordings
- **Telegram** - Messages, bots
- **WhatsApp Business** - Messages via Twilio
- **Cisco Webex** - Meetings, teams
- **Mattermost** - Messages, channels

### 📧 Email & Marketing (8)
- **Gmail** - Send emails, read, labels, filters
- **SendGrid** - Transactional emails, templates
- **Mailchimp** - Campaigns, audiences, automations
- **Twilio** - SMS, voice, WhatsApp
- **Postmark** - Transactional email
- **AWS SES** - Email sending service
- **Customer.io** - Marketing automation
- **Braze** - Customer engagement

### 📊 Project Management (9)
- **Jira** - Issues, projects, workflows
- **Trello** - Boards, cards, lists
- **Asana** - Tasks, projects, teams
- **Monday.com** - Boards, items, columns
- **ClickUp** - Tasks, docs, goals, dashboards
- **Notion** - Pages, databases, blocks
- **Basecamp** - Projects, to-dos, messages
- **Linear** - Issues, projects, cycles
- **Height** - Tasks, lists, workspaces

### 💼 CRM & Sales (7)
- **Salesforce** - Leads, contacts, opportunities
- **HubSpot** - CRM, marketing, sales hub
- **Pipedrive** - Deals, activities, contacts
- **Zendesk** - Tickets, users, organizations
- **Intercom** - Messages, users, conversations
- **Freshdesk** - Tickets, contacts
- **Close.com** - Leads, calls, emails

### 🛠️ Development Tools (3)
- **GitHub** - Repos, issues, PRs, actions
- **GitLab** - Projects, MRs, CI/CD
- **Bitbucket** - Repositories, PRs, pipelines

### ☁️ Cloud Storage (4)
- **Google Drive** - Files, folders, permissions
- **Dropbox** - Files and folders
- **OneDrive** - Files, drive items
- **Box** - Files, folders, collaboration

### 💳 Payment & E-commerce (4)
- **Stripe** - Payments, subscriptions, customers
- **PayPal** - Payments, invoices
- **Shopify** - Products, orders, customers
- **Square** - Payments, inventory, customers

### 📈 Data & Analytics (5)
- **Google Sheets** - Spreadsheets, ranges, formulas
- **Airtable** - Bases, tables, records
- **Tableau** - Workbooks, views, data sources
- **Excel Online** - Workbooks, worksheets
- **Metabase** - Dashboards, questions

Each integration supports:
- ✅ OAuth 2.0 authentication
- ✅ Automatic token refresh
- ✅ Rate limiting and retry logic
- ✅ Webhook support (where available)
- ✅ Comprehensive error handling

## 🎨 Developer Experience

### Multi-Tenancy: Workspaces

Every developer gets an isolated workspace:

```go
type Workspace struct {
    ID          string
    Name        string
    APIKey      string      // For server-side calls
    WebhookURL  string      // Where we send events
    Plan        string      // free, professional, enterprise
    Settings    WorkspaceSettings
    CreatedAt   time.Time
}

type WorkspaceSettings struct {
    AllowClientIntegrations  bool
    MaxWorkflows             int
    MaxExecutionsPerMonth    int
    RetentionDays            int
    WebhookSecret            string
    CustomDomains            []string
}
```

### Client Authorization Flow

**For your end users** (your customers):

```typescript
// Frontend: Get auth URL
const response = await fetch('/api/integrations/slack/connect', {
  method: 'POST'
});
const { authUrl } = await response.json();

// Redirect user
window.location.href = authUrl;

// After OAuth callback (backend):
app.get('/integrations/callback', async (req, res) => {
  const { code, state } = req.query;
  
  await neighbourhoodAPI.integrations.exchangeCode({
    code,
    state,
    userId: req.session.userId
  });
  
  res.redirect('/dashboard?success=true');
});
```

### Workflow Definition Language

Simple JSON-based DSL:

```json
{
  "name": "Process Support Ticket",
  "description": "Auto-triage and assign support tickets",
  "trigger": {
    "type": "webhook",
    "path": "/support-ticket"
  },
  "steps": [
    {
      "id": "classify",
      "type": "condition",
      "condition": "{{ticket.priority}} == 'high'",
      "then": ["create_jira"],
      "else": ["send_to_zendesk"]
    },
    {
      "id": "create_jira",
      "provider": "jira",
      "action": "create_issue",
      "config": {
        "project": "SUP",
        "issuetype": "Bug",
        "summary": "{{ticket.title}}",
        "description": "{{ticket.description}}",
        "priority": "High"
      },
      "retry": { "max_attempts": 3 }
    },
    {
      "id": "send_to_zendesk",
      "provider": "zendesk",
      "action": "create_ticket",
      "config": {
        "subject": "{{ticket.title}}",
        "description": "{{ticket.description}}",
        "priority": "normal"
      }
    },
    {
      "id": "notify_team",
      "provider": "slack",
      "action": "send_message",
      "config": {
        "channel": "#support",
        "text": "New {{ticket.priority}} priority ticket: {{ticket.title}}"
      },
      "depends_on": ["classify"]
    }
  ]
}
```

## 🔐 Security & Compliance

### Authentication & Authorization

- **Workspace API Keys**: Server-to-server authentication
- **JWT Tokens**: User session management (15min access, 7d refresh)
- **OAuth 2.0**: Integration authorization
- **PKCE**: Enhanced security for public clients
- **Rate Limiting**: Redis-based with 5000 req/hour per workspace

### Data Protection

- **Encryption at Rest**: AES-256 for credentials
- **Encryption in Transit**: TLS 1.3
- **Token Storage**: Encrypted OAuth tokens in PostgreSQL
- **Secrets Management**: HashiCorp Vault or AWS Secrets Manager
- **Audit Logs**: Complete audit trail with 90-day retention

### Compliance

- **GDPR**: Data portability & right to deletion APIs
- **SOC 2 Type II**: Security controls documentation
- **HIPAA**: Healthcare data handling (PHI encryption)
- **PCI DSS**: Payment data compliance (Stripe, PayPal)

## 📊 Monitoring & Observability

### Metrics (Prometheus)

```promql
# Request rate
rate(grpc_server_handled_total[5m])

# Error rate
rate(grpc_server_handled_total{grpc_code!="OK"}[5m])

# Latency P95
histogram_quantile(0.95, grpc_server_handling_seconds_bucket)

# Integration success rate
sum(rate(integration_action_total{status="success"}[5m])) /
sum(rate(integration_action_total[5m]))
```

### Distributed Tracing (Jaeger)

Every request gets a trace ID:

```
Trace: workflow_execution_abc123
├─ API Gateway: POST /workflows/execute (10ms)
├─ Auth Service: ValidateToken (5ms)
├─ Integration Service: GetAuthURL (150ms)
│  ├─ PostgreSQL: SELECT user_integrations (8ms)
│  └─ Redis: GET oauth_token:user_123 (2ms)
└─ Workflow Service: Execute (2.5s)
   ├─ Slack API: POST /chat.postMessage (800ms)
   └─ Salesforce API: POST /sobjects/Lead (1.2s)
```

### Pre-built Grafana Dashboards

- **Service Health**: CPU, memory, error rates
- **Workflow Executions**: Success/failure rates, duration
- **Integration Performance**: API latency by provider
- **Business Metrics**: Active workspaces, executions/day

## 🚀 Deployment

### Development (Docker Compose)

```bash
docker-compose up -d
```

Includes:
- PostgreSQL (auth, integration, workflow DBs)
- Redis (6 databases for different services)
- MongoDB (notifications)
- Jaeger (tracing)
- Prometheus (metrics)
- Grafana (dashboards)

### Production (Kubernetes)

```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Secrets & Config
kubectl create secret generic neighbourhood-secrets \
  --from-literal=jwt-secret=$JWT_SECRET \
  --from-literal=db-password=$DB_PASSWORD

# Deploy services
kubectl apply -f k8s/

# Autoscaling
kubectl autoscale deployment integration-service \
  --cpu-percent=70 \
  --min=3 \
  --max=10
```

### Environment Configuration

```bash
# Required
ENVIRONMENT=production
JWT_SECRET=your-super-secret-key
DATABASE_URL=postgresql://user:pass@host/db

# Optional
LOG_LEVEL=info
REDIS_URL=redis://localhost:6379
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
PROMETHEUS_PORT=9090
```

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [API Reference](./API_DOCUMENTATION.md) | Complete REST API documentation |
| [Architecture Guide](./ARCHITECTURE.md) | System design deep dive |
| [Microservices Guide](./MICROSERVICES_GUIDE.md) | Service-by-service breakdown |
| [Deployment Guide](./DEPLOYMENT.md) | Production deployment |
| [Integration Guides](./docs/integrations/) | Provider-specific docs |
| [Workflow Examples](./examples/workflows/) | Real-world workflow templates |

## 🛠️ Development

### Project Structure

```
NeighbourHood/
├── services/               # Microservices
│   ├── auth/              # Authentication & user management
│   ├── integration/       # Integration provider management
│   ├── workflow/          # Workflow orchestration engine
│   ├── consent/           # Consent & data governance
│   └── notification/      # Notifications & webhooks
├── proto/                 # Protobuf definitions
│   ├── auth.proto
│   ├── integration.proto
│   ├── workflow.proto
│   ├── consent.proto
│   ├── notification.proto
│   └── common.proto
├── configs/               # Service configurations
├── web/                   # Frontend (developer portal)
├── k8s/                   # Kubernetes manifests
├── docs/                  # Documentation
├── examples/              # Example workflows
└── Makefile               # Build automation
```

### Building Services

```bash
# Install dependencies
go mod download

# Generate protobuf code
make proto

# Build all services
make build-all

# Build individual service
make build-auth
make build-integration
make build-workflow

# Run tests
make test

# Lint code
make lint
```

### Adding a New Integration

1. **Define Provider**:
```go
// services/integration/pkg/providers/newprovider.go
type NewProviderProvider struct {
    config ProviderConfig
}

func (p *NewProviderProvider) GetAuthURL(state string, scopes []string) (string, error) {
    // OAuth URL generation
}

func (p *NewProviderProvider) ExchangeCode(code string) (*Token, error) {
    // Token exchange
}

func (p *NewProviderProvider) Execute(token *Token, action string, params map[string]interface{}) (interface{}, error) {
    switch action {
    case "send_message":
        return p.sendMessage(token, params)
    // ... other actions
    }
}
```

2. **Register Provider**:
```go
// services/integration/pkg/providers/registry.go
func init() {
    Register("newprovider", &NewProviderProvider{})
}
```

3. **Add Configuration**:
```yaml
# configs/integration.yaml
providers:
  newprovider:
    name: "New Provider"
    category: "communication"
    client_id: "${NEWPROVIDER_CLIENT_ID}"
    client_secret: "${NEWPROVIDER_CLIENT_SECRET}"
    scopes: ["read", "write"]
```

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## 📄 License

Proprietary software. Copyright © 2024 NeighbourHood Platform.

For licensing inquiries: licensing@neighbourhood.dev

## 🆘 Support

- **Documentation**: https://docs.neighbourhood.dev
- **API Status**: https://status.neighbourhood.dev
- **Community**: https://community.neighbourhood.dev
- **Email**: support@neighbourhood.dev
- **Enterprise**: enterprise@neighbourhood.dev

---

**Made with ❤️ by developers, for developers**

*Simplifying integration complexity so you can focus on building great products.*
