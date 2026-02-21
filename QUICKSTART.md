# 🚀 Quick Start Guide

Get NeighbourHood up and running in 5 minutes!

## Prerequisites

- Go 1.20+ installed
- Git installed
- (Optional) Docker & Docker Compose

## Option 1: Run Without Docker (Fastest)

### 1. Clone and Setup

```bash
git clone <your-repo-url>
cd NeighbourHood
cp .env.example .env
```

### 2. Run the Server

```bash
cd cmdapi
go run main.go
```

### 3. Access the Platform

Open your browser and visit:
```
http://localhost:8080
```

You should see the developer portal!

## Option 2: Run With Docker

### 1. Clone and Setup

```bash
git clone <your-repo-url>
cd NeighbourHood
cp .env.example .env
```

### 2. Start Services

```bash
docker-compose up
```

### 3. Access the Platform

```
http://localhost:8080
```

## 🧪 Test the API

### Test Health Endpoint

```bash
curl http://localhost:8080/health
```

Expected: `OK`

### Get Available Integrations

```bash
curl http://localhost:8080/api/integrations
```

Expected:
```json
{
  "integrations": [
    {"type": "slack", "name": "Slack", "description": "Team collaboration and messaging"},
    {"type": "gmail", "name": "Gmail", "description": "Email service"},
    {"type": "jira", "name": "Jira", "description": "Project management and issue tracking"}
  ]
}
```

### Get Integration Auth URL

```bash
curl -X POST http://localhost:8080/api/integration/authurl \
  -H "Content-Type: application/json" \
  -d '{"provider": "slack", "state": "test-state"}'
```

Expected:
```json
{
  "url": "https://slack.com/oauth/v2/authorize?client_id=...&state=test-state..."
}
```

### Execute Integration Action (Mock)

```bash
curl -X POST http://localhost:8080/api/integration/execute \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "slack",
    "token": {"access_token": "test-token"},
    "action": "send_message",
    "payload": {"channel": "#general", "text": "Hello World!"}
  }'
```

Expected:
```json
{
  "result": {
    "status": "success",
    "message": "Sent 'Hello World!' to #general"
  }
}
```

### Execute a Workflow (Mock)

```bash
curl -X POST http://localhost:8080/api/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "workflow": {
      "id": "00000000-0000-0000-0000-000000000000",
      "name": "Test Workflow",
      "steps": [
        {
          "provider": "slack",
          "action": "send_message",
          "payload": {"channel": "#general", "text": "Step 1 complete"}
        }
      ]
    },
    "tokens": {
      "slack": {"access_token": "test-token"}
    }
  }'
```

## 📝 Next Steps

### 1. Configure OAuth Credentials

To use real integrations, you need OAuth credentials:

**Slack:**
1. Go to https://api.slack.com/apps
2. Create a new app
3. Get Client ID and Secret
4. Update `.env`:
```env
SLACK_CLIENT_ID=your-client-id
SLACK_CLIENT_SECRET=your-client-secret
SLACK_REDIRECT_URL=http://localhost:8080/callback/slack
```

**Gmail:**
1. Go to https://console.cloud.google.com
2. Create a new project
3. Enable Gmail API
4. Create OAuth 2.0 credentials
5. Update `.env`:
```env
GMAIL_CLIENT_ID=your-client-id
GMAIL_CLIENT_SECRET=your-client-secret
GMAIL_REDIRECT_URL=http://localhost:8080/callback/gmail
```

**Jira:**
1. Go to https://developer.atlassian.com/console/myapps/
2. Create an OAuth 2.0 integration
3. Update `.env`:
```env
JIRA_CLIENT_ID=your-client-id
JIRA_CLIENT_SECRET=your-client-secret
JIRA_REDIRECT_URL=http://localhost:8080/callback/jira
```

### 2. Set Up Database (For Production Features)

```bash
# Install PostgreSQL
# On Ubuntu/Debian:
sudo apt install postgresql

# On macOS:
brew install postgresql

# Create database
createdb neighbourhood

# Run migrations
psql -d neighbourhood -f internal/models/schema.sql
```

Update `.env`:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=neighbourhood
```

### 3. Explore the Developer Portal

Visit http://localhost:8080 to:
- View available integrations
- Get OAuth authorization URLs
- Test workflow execution
- Manage API keys

### 4. Read Full Documentation

- [README.md](README.md) - Complete overview
- [API_DOCUMENTATION.md](API_DOCUMENTATION.md) - API reference
- [DEPLOYMENT.md](DEPLOYMENT.md) - Production deployment
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines

## 🛠️ Development Commands

```bash
# Run with hot reload (install air first: go install github.com/cosmtrek/air@latest)
make dev

# Build binary
make build

# Run tests
make test

# Format code
make format

# Check for issues
make lint
```

## 🐛 Troubleshooting

### Port 8080 already in use

Change the port in `.env`:
```env
PORT=3000
```

### Database connection error

The platform runs in offline mode without a database. Core features work, but some features require a database:
- User authentication
- Stored integrations
- Workflow history

### Go module errors

```bash
go mod tidy
go mod download
```

## 📚 Example Use Cases

### 1. Send Slack Message

```javascript
// Using JavaScript/Node.js
const sendSlackMessage = async () => {
  const response = await fetch('http://localhost:8080/api/integration/execute', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      provider: 'slack',
      token: { access_token: 'your-token' },
      action: 'send_message',
      payload: { channel: '#general', text: 'Hello from API!' }
    })
  });
  return await response.json();
};
```

### 2. Create Jira Issue

```python
# Using Python
import requests

response = requests.post(
    'http://localhost:8080/api/integration/execute',
    json={
        'provider': 'jira',
        'token': {'access_token': 'your-token'},
        'action': 'create_issue',
        'payload': {
            'project': 'PROJ',
            'summary': 'New task from API',
            'issue_type': 'Task'
        }
    }
)
print(response.json())
```

### 3. Multi-Step Workflow

```bash
# Email notification + Slack alert
curl -X POST http://localhost:8080/api/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "workflow": {
      "name": "Notification Workflow",
      "steps": [
        {
          "provider": "gmail",
          "action": "send_email",
          "payload": {
            "to": "user@example.com",
            "subject": "Alert",
            "body": "Important notification"
          }
        },
        {
          "provider": "slack",
          "action": "send_message",
          "payload": {
            "channel": "#alerts",
            "text": "✅ Email sent!"
          }
        }
      ]
    },
    "tokens": {
      "gmail": {"access_token": "gmail-token"},
      "slack": {"access_token": "slack-token"}
    }
  }'
```

## 🎉 You're Ready!

You now have a fully functional integration platform. Start building amazing integrations!

## 💬 Need Help?

- Check the [FAQ](README.md#faq)
- Open an [issue](../../issues)
- Read the [full documentation](API_DOCUMENTATION.md)

Happy coding! 🚀
