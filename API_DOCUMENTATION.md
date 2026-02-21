# NeighbourHood API Documentation

## Base URL
```
http://localhost:8080
```

## Authentication

Most endpoints require authentication via Bearer token:

```http
Authorization: Bearer YOUR_JWT_TOKEN
```

---

## Endpoints

### 1. Health Check

Check if the server is running.

**Request:**
```http
GET /health
```

**Response:**
```
200 OK
```

---

### 2. List Available Integrations

Get all available integration providers.

**Request:**
```http
GET /api/integrations
```

**Response:**
```json
{
  "integrations": [
    {
      "type": "slack",
      "name": "Slack",
      "description": "Team collaboration and messaging"
    },
    {
      "type": "gmail",
      "name": "Gmail",
      "description": "Email service"
    },
    {
      "type": "jira",
      "name": "Jira",
      "description": "Project management and issue tracking"
    }
  ]
}
```

---

### 3. Get Integration Auth URL

Get the OAuth authorization URL for a provider.

**Request:**
```http
POST /api/integration/authurl
Content-Type: application/json

{
  "provider": "slack",
  "state": "random-state-string"
}
```

**Response:**
```json
{
  "url": "https://slack.com/oauth/v2/authorize?client_id=...&state=...&redirect_uri=..."
}
```

---

### 4. Execute Integration Action

Execute a single action on a provider.

**Request:**
```http
POST /api/integration/execute
Content-Type: application/json

{
  "provider": "slack",
  "token": {
    "access_token": "xoxb-your-token",
    "token_type": "Bearer"
  },
  "action": "send_message",
  "payload": {
    "channel": "#general",
    "text": "Hello from NeighbourHood!"
  }
}
```

**Response:**
```json
{
  "result": {
    "status": "success",
    "message": "Sent 'Hello from NeighbourHood!' to #general"
  }
}
```

**Error Response:**
```json
{
  "error": "consent not granted: consent not granted for this provider"
}
```

---

### 5. Execute Workflow

Execute a multi-step workflow across multiple integrations.

**Request:**
```http
POST /api/workflow/execute
Content-Type: application/json

{
  "workflow": {
    "id": "00000000-0000-0000-0000-000000000000",
    "name": "Email Notification Workflow",
    "steps": [
      {
        "provider": "gmail",
        "action": "send_email",
        "payload": {
          "to": "user@example.com",
          "subject": "Workflow Notification",
          "body": "This is an automated message from your workflow."
        }
      },
      {
        "provider": "slack",
        "action": "send_message",
        "payload": {
          "channel": "#notifications",
          "text": "✅ Email sent to user@example.com"
        }
      }
    ]
  },
  "tokens": {
    "gmail": {
      "access_token": "ya29.your-gmail-token"
    },
    "slack": {
      "access_token": "xoxb-your-slack-token"
    }
  }
}
```

**Response:**
```json
{
  "results": [
    {
      "status": "success",
      "message": "Email sent to user@example.com with subject 'Workflow Notification'"
    },
    {
      "status": "success",
      "message": "Sent '✅ Email sent to user@example.com' to #notifications"
    }
  ]
}
```

**Error Response:**
```json
{
  "error": "workflow execution failed: step 0 failed: missing 'to' field"
}
```

---

## Provider-Specific Actions

### Slack

#### send_message
Send a message to a Slack channel.

**Payload:**
```json
{
  "channel": "#general",
  "text": "Your message here"
}
```

**Result:**
```json
{
  "status": "success",
  "message": "Sent 'Your message here' to #general"
}
```

---

### Gmail

#### send_email
Send an email via Gmail.

**Payload:**
```json
{
  "to": "recipient@example.com",
  "subject": "Email Subject",
  "body": "Email body content"
}
```

**Result:**
```json
{
  "status": "success",
  "message": "Email sent to recipient@example.com with subject 'Email Subject'"
}
```

---

### Jira

#### create_issue
Create a new issue in Jira.

**Payload:**
```json
{
  "project": "PROJ",
  "summary": "Issue summary",
  "issue_type": "Task"
}
```

**Result:**
```json
{
  "status": "success",
  "message": "Created Task in project PROJ: Issue summary",
  "issue_key": "DEMO-123"
}
```

---

## Error Codes

| Status Code | Description |
|------------|-------------|
| 200 | Success |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing or invalid token |
| 403 | Forbidden - Consent not granted |
| 404 | Not Found - Provider or resource not found |
| 500 | Internal Server Error |

---

## Workflow Structure

A workflow consists of multiple steps executed sequentially:

```json
{
  "id": "uuid",
  "name": "Workflow Name",
  "steps": [
    {
      "provider": "provider_name",
      "action": "action_name",
      "payload": {
        // Action-specific payload
      }
    }
  ]
}
```

**Important Notes:**
- Steps are executed in order
- If a step fails, the workflow stops
- Each step requires a valid token for its provider
- All providers in the workflow must have user consent

---

## SDK Examples

### JavaScript/Node.js

```javascript
const BASE_URL = 'http://localhost:8080';

async function executeIntegration(provider, action, payload, token) {
  const response = await fetch(`${BASE_URL}/api/integration/execute`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      provider,
      token,
      action,
      payload
    })
  });
  
  return await response.json();
}

// Example usage
const result = await executeIntegration(
  'slack',
  'send_message',
  { channel: '#general', text: 'Hello!' },
  { access_token: 'xoxb-token' }
);
```

### Python

```python
import requests

BASE_URL = 'http://localhost:8080'

def execute_integration(provider, action, payload, token):
    response = requests.post(
        f'{BASE_URL}/api/integration/execute',
        json={
            'provider': provider,
            'token': token,
            'action': action,
            'payload': payload
        }
    )
    return response.json()

# Example usage
result = execute_integration(
    'slack',
    'send_message',
    {'channel': '#general', 'text': 'Hello!'},
    {'access_token': 'xoxb-token'}
)
```

### cURL

```bash
curl -X POST http://localhost:8080/api/integration/execute \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "slack",
    "token": {
      "access_token": "xoxb-token"
    },
    "action": "send_message",
    "payload": {
      "channel": "#general",
      "text": "Hello from cURL!"
    }
  }'
```

---

## Best Practices

1. **Always validate responses**: Check for error fields before using results
2. **Handle token expiration**: Implement token refresh logic
3. **Respect rate limits**: Implement exponential backoff for retries
4. **Secure token storage**: Never expose tokens in client-side code
5. **Use HTTPS in production**: Always encrypt API communication
6. **Implement consent checks**: Ensure users have granted consent before executing actions
7. **Log workflow executions**: Keep audit trail for debugging and compliance

---

## Support

For issues, questions, or feature requests, please contact [support@neighbourhood.dev](mailto:support@neighbourhood.dev)
