# NeighbourHood Platform - Developer Guide

## 🚀 Overview

NeighbourHood is a production-grade, open-source B2B2C integration platform that enables developers to connect their applications to 100+ SaaS platforms with a single, unified API. Built with enterprise-grade security, role-based access control, and optimized for performance.

## ✨ Key Features

### 🎨 Modern Light Mode UI
- **Professional Design**: Clean white and purple theme optimized for developer productivity
- **Responsive Layout**: Works seamlessly on desktop, tablet, and mobile devices
- **Intuitive Navigation**: Easy-to-use dashboard with clear information hierarchy
- **Dark/Light Toggle**: Coming soon - customizable theme preferences

### 🔐 Role-Based Access Control (RBAC)

#### Roles & Permissions

**Four Built-in Roles:**

1. **Admin** - Full workspace control
   - All integration operations
   - API key management (create, revoke)
   - User management
   - Workspace settings

2. **Developer** - Primary development role
   - Create and manage integrations
   - Create API keys
   - Execute integration actions
   - Read workspace data

3. **User** - Limited integration access
   - Read integrations
   - Execute approved actions
   - View workspace data

4. **Viewer** - Read-only access
   - View integrations
   - View workspace members
   - No modification permissions

#### Permission System

Granular permission model with O(1) lookup performance:

```go
const (
    // Integration permissions
    PermissionIntegrationRead   Permission = "integration:read"
    PermissionIntegrationWrite  Permission = "integration:write"
    PermissionIntegrationDelete Permission = "integration:delete"
    PermissionIntegrationExecute Permission = "integration:execute"
    
    // API Key permissions
    PermissionAPIKeyRead   Permission = "apikey:read"
    PermissionAPIKeyCreate Permission = "apikey:create"
    PermissionAPIKeyRevoke Permission = "apikey:revoke"
    
    // User management permissions
    PermissionUserRead   Permission = "user:read"
    PermissionUserWrite  Permission = "user:write"
    PermissionUserDelete Permission = "user:delete"
    
    // Workspace permissions
    PermissionWorkspaceRead   Permission = "workspace:read"
    PermissionWorkspaceWrite  Permission = "workspace:write"
    PermissionWorkspaceDelete Permission = "workspace:delete"
)
```

### 🔑 API Key Management

#### Secure Key Generation
- **Cryptographically Secure**: 256-bit random keys using `crypto/rand`
- **SHA-256 Hashing**: Keys hashed before storage, never stored in plaintext
- **Unique Prefixes**: Easy identification with `nh_live_pk_` prefix
- **One-Time Display**: Full key shown only once during creation

#### Key Format
```
nh_live_pk_abc123def456_fullsecretkey789xyz
│    │     │             │
│    │     │             └─ Secret portion (base64-encoded)
│    │     └─ Key prefix (first 12 chars of secret)
│    └─ Key type (live/test)
└─ Platform identifier
```

#### Key Features
- Rate limiting (configurable per key)
- Expiration dates (optional)
- Scope restrictions
- Usage analytics
- Instant revocation

### 🔗 Integration Signup Flow

#### Smooth Developer Experience

**Step 1: Browse Integrations**
- Search through 100+ available integrations
- Filter by category (Communication, Email, CRM, etc.)
- View required permissions upfront

**Step 2: Configure Integration**
- Integration-specific setup fields
- Clear permission requirements
- OAuth scope explanation

**Step 3: Authorize**
- Secure OAuth 2.0 flow
- User grant consent
- Automatic token management

**Step 4: Use Integration**
- Instant availability via API
- SDK support
- Webhook notifications

### 📊 Database Schema

Optimized for O(1) lookups with proper indexing:

```sql
-- User Roles (RBAC)
CREATE TABLE user_roles (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    workspace_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    permissions JSONB,
    -- Composite index for O(1) permission checks
    CONSTRAINT unique_user_workspace UNIQUE(user_id, workspace_id)
);
CREATE UNIQUE INDEX idx_user_roles_composite ON user_roles(user_id, workspace_id);

-- API Keys
CREATE TABLE api_keys (
    id VARCHAR(255) PRIMARY KEY,
    key_hash VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 hash
    workspace_id VARCHAR(255) NOT NULL,
    rate_limit INTEGER DEFAULT 1000,
    active BOOLEAN DEFAULT true
);
-- Hash index for O(1) validation
CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(key_hash) WHERE active = true;

-- Integration Subscriptions
CREATE TABLE integration_subscriptions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    integration_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    required_scopes JSONB,
    granted_scopes JSONB,
    config JSONB
);
CREATE UNIQUE INDEX idx_integration_sub_composite 
    ON integration_subscriptions(user_id, workspace_id, integration_type);
```

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Node.js 18+ (for SDK development)

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/neighbourhood/platform.git
cd platform
```

2. **Install dependencies**
```bash
make deps
```

3. **Set up database**
```bash
# Run migrations
psql -U postgres -d neighbourhood < internal/models/schema.sql
psql -U postgres -d neighbourhood < internal/models/rbac_schema.sql
```

4. **Configure environment**
```bash
cp configs/auth.yaml.example configs/auth.yaml
cp configs/integration.yaml.example configs/integration.yaml
# Edit with your OAuth credentials
```

5. **Build services**
```bash
make build-all
```

6. **Run services**
```bash
# Auth service
./bin/auth-service

# Integration service
./bin/integration-service
```

### Quick Start Example

```typescript
import { NeighbourHood } from '@neighbourhood/sdk';

// Initialize with your API key
const nh = new NeighbourHood({
  apiKey: process.env.NH_API_KEY
});

// Connect to Slack
const slack = await nh.connect('slack');

// Send message with automatic permission check
await slack.sendMessage({
  channel: '#general',
  text: 'Hello from NeighbourHood! 🎉'
});

// Create Jira issue
const jira = await nh.connect('jira');
await jira.createIssue({
  project: 'PROJ',
  summary: 'Integration test',
  type: 'Task'
});
```

## 🏗️ Architecture

### Microservices

**Auth Service** (Port 50051)
- User authentication (email/password, OAuth)
- RBAC permission checks
- API key validation
- Workspace management

**Integration Service** (Port 50052)
- Integration catalog
- OAuth flow management
- Action execution
- Webhook handling

### Performance Optimizations

#### Time Complexity
- **Permission Check**: O(1) - Composite index lookup + small array scan
- **API Key Validation**: O(1) - Hash index lookup
- **User Lookup**: O(1) - Primary key index
- **Integration List**: O(n) where n is typically < 100

#### Space Complexity
- **API Key**: O(1) - Fixed 32 bytes + metadata
- **User Session**: O(1) - Fixed size JWT token
- **Permission Cache**: O(k) - k permissions per role (< 20)

#### Database Indexes
```sql
-- Critical for O(1) performance
CREATE UNIQUE INDEX idx_user_roles_composite ON user_roles(user_id, workspace_id);
CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_integration_sub_user ON integration_subscriptions(user_id);
```

## 🎨 UI Components

### Landing Page
- **Hero Section**: Clear value proposition with code example
- **Features Grid**: 6-card layout showcasing capabilities
- **Integration Logos**: Visual representation of available platforms
- **How It Works**: 3-step process explanation
- **CTA Section**: Prominent "Get Started" and "Contribute" buttons
- **Footer**: Comprehensive navigation and resources

### Dashboard
- **Workspace Selector**: Switch between multiple workspaces
- **API Key Manager**: Create, view, and revoke keys
- **Integration Browser**: Search and filter 100+ integrations
- **Active Integrations**: Manage connected services
- **Usage Analytics**: Coming soon

## 🔒 Security Best Practices

### API Key Security
1. **Never commit keys** to version control
2. **Use environment variables** for key storage
3. **Rotate keys regularly** (recommended: every 90 days)
4. **Set expiration dates** for temporary access
5. **Monitor usage** for suspicious activity
6. **Use scopes** to limit key capabilities

### OAuth Security
1. **State parameter** for CSRF protection
2. **PKCE** for public clients
3. **Secure token storage** (encrypted at rest)
4. **Automatic token refresh** with rotation
5. **Audit logging** for all OAuth events

### RBAC Best Practices
1. **Principle of Least Privilege**: Grant minimum required permissions
2. **Regular Access Reviews**: Audit user roles quarterly
3. **Workspace Isolation**: Separate production/development workspaces
4. **Audit Trails**: Log all permission changes
5. **Custom Permissions**: Fine-tune access per integration

## 📈 Production Grade Features

### Reliability
- **99.9% Uptime SLA**
- **Automatic retry** with exponential backoff
- **Circuit breakers** for third-party APIs
- **Health checks** on all services
- **Graceful degradation**

### Scalability
- **Horizontal scaling** via load balancers
- **Database connection pooling**
- **Redis caching** for hot paths
- **Rate limiting** per workspace
- **Async job processing**

### Monitoring
- **Prometheus metrics** on all endpoints
- **Distributed tracing** with OpenTelemetry
- **Structured logging** via Zap
- **Error tracking** with Sentry integration
- **Performance profiling**

### Compliance
- **GDPR compliant** data handling
- **SOC 2 Type II** (in progress)
- **Data encryption** at rest and in transit
- **Audit logs** for all actions
- **Data retention policies**

## 🤝 Contributing

We welcome contributions! This is an open-source project built by developers, for developers.

### Ways to Contribute

1. **Add Integrations**: Implement new platform connectors
2. **Fix Bugs**: Check our issue tracker
3. **Improve Docs**: Help make our documentation clearer
4. **Write Tests**: Increase code coverage
5. **Share Feedback**: Tell us what you need

### Development Workflow

```bash
# Fork and clone
git clone https://github.com/YOUR_USERNAME/neighbourhood.git

# Create feature branch
git checkout -b feature/amazing-integration

# Make changes and test
make test
make build-all

# Commit with conventional commits
git commit -m "feat: add Asana integration"

# Push and create PR
git push origin feature/amazing-integration
```

## 📝 License

MIT License - See [LICENSE](LICENSE) for details

## 🙏 Acknowledgments

Built with ❤️ by the developer community using:
- Go for high-performance backends
- gRPC for efficient microservice communication
- PostgreSQL for reliable data storage
- Redis for blazing-fast caching
- Vanilla JS for lightweight frontend

---

**Made with 💜 by the NeighbourHood Team**

[Website](https://neighbourhood.dev) • [Documentation](https://docs.neighbourhood.dev) • [GitHub](https://github.com/neighbourhood/platform) • [Discord](https://discord.gg/neighbourhood)
