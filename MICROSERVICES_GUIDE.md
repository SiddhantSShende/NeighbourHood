# NeighbourHood - Microservices Implementation Guide

## 📋 Project Overview

NeighbourHood is a production-grade third-party integration platform for SaaS/PaaS developers, enabling seamless integration across 45+ platforms including Communication, Email, Project Management, CRM, Development Tools, Storage, Payment, Data Analytics, and Social Media.

**Architecture Evolution:**
- **Phase 1:** Monolithic architecture with 45+ integrations ✅ Complete
- **Phase 2:** Microservices transformation with gRPC, Protobuf, Clean Architecture ✅ In Progress

## 🏗️ Microservices Architecture

### Services Overview

| Service | Port (gRPC) | Metrics Port | Database | Description |
|---------|-------------|--------------|----------|-------------|
| **Auth Service** | 50051 | 9091 | PostgreSQL | Authentication, JWT, OAuth (Google, GitHub, Microsoft) |
| **Integration Service** | 50052 | 9092 | PostgreSQL | 45+ provider integrations, OAuth flows, action execution |
| **Workflow Service** | 50053 | 9093 | PostgreSQL | Workflow orchestration, step execution, async processing |
| **Consent Service** | 50054 | 9094 | PostgreSQL | GDPR compliance, consent management, audit logs |
| **Notification Service** | 50055 | 9095 | MongoDB | Multi-channel notifications, webhooks, delivery tracking |
| **API Gateway** | 8080 | 9096 | Redis | HTTP to gRPC, rate limiting, authentication middleware |

### Technology Stack

**Core:**
- Go 1.21+
- gRPC for inter-service communication
- Protocol Buffers v3 for serialization
- Clean Architecture (Domain, UseCase, Repository, Delivery)

**Databases:**
- PostgreSQL (Auth, Integration, Workflow, Consent)
- MongoDB (Notifications)
- Redis (Sessions, Rate Limiting, Caching)

**Observability:**
- **Logging:** Zap (structured JSON logging)
- **Metrics:** Prometheus + Grafana
- **Tracing:** OpenTelemetry + Jaeger
- **Health Checks:** gRPC health protocol

**Service Discovery:**
- Consul for service registry and health checks

**Resilience:**
- Circuit Breaker pattern
- Retry with exponential backoff
- Rate limiting
- Timeout management

## 🚀 Quick Start

### Prerequisites

```bash
# Required tools
- Go 1.21+
- Protocol Buffers compiler (protoc)
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+
- MongoDB 7+ (for notifications)

# Install development tools
make install-tools
```

### Setup Steps

#### 1. Generate Protobuf Code

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code from .proto files
make proto
```

This generates Go code in `proto/gen/go/` from all `.proto` service definitions.

#### 2. Install Dependencies

```bash
# Download Go modules
go mod download
go mod tidy
```

#### 3. Configure Environment

Create a `.env` file with required environment variables:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=neighbourhood_auth

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-super-secret-jwt-key-min-32-chars

# OAuth Providers
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/oauth/google/callback

GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GITHUB_REDIRECT_URL=http://localhost:8080/api/v1/auth/oauth/github/callback

# Observability
JAEGER_ENDPOINT=http://localhost:14268/api/traces
TRACING_ENABLED=true
```

#### 4. Start Infrastructure Services

Using Docker Compose:

```bash
docker-compose up -d postgres redis mongodb consul jaeger prometheus grafana
```

Or manually:

```bash
# PostgreSQL
docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:15

# Redis
docker run -d --name redis -p 6379:6379 redis:7-alpine

# MongoDB
docker run -d --name mongodb -p 27017:27017 mongo:7

# Jaeger (tracing)
docker run -d --name jaeger -p 14268:14268 -p 16686:16686 jaegertracing/all-in-one:latest

# Prometheus (metrics)
docker run -d --name prometheus -p 9090:9090 prom/prometheus:latest

# Consul (service discovery)
docker run -d --name consul -p 8500:8500 hashicorp/consul:latest agent -dev -ui
```

#### 5. Build and Run Services

**Auth Service:**

```bash
# Build
make build-auth

# Run
make run-auth

# Or directly
./bin/auth-service
```

The auth service will:
- Start gRPC server on port 50051
- Expose Prometheus metrics on port 9091
- Connect to PostgreSQL and Redis
- Initialize database schema automatically
- Send traces to Jaeger

**Check Service Health:**

```bash
# Using grpcurl
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check

# Check metrics
curl http://localhost:9091/metrics
```

## 📁 Project Structure

```
NeighbourHood/
├── proto/                          # Protocol Buffer definitions
│   ├── common.proto               # Shared types (Error, Pagination, Metadata)
│   ├── auth.proto                 # Auth service (11 RPC methods)
│   ├── integration.proto          # Integration service (12 RPC methods)
│   ├── workflow.proto             # Workflow service (12 RPC methods)
│   ├── consent.proto              # Consent service (9 RPC methods)
│   ├── notification.proto         # Notification service (11 RPC methods)
│   └── gen/go/                    # Generated Go code
│
├── configs/                        # External YAML configurations
│   ├── auth.yaml                  # Auth service config
│   ├── integration.yaml           # Integration service config
│   ├── workflow.yaml              # Workflow service config
│   ├── consent.yaml               # Consent service config
│   ├── notification.yaml          # Notification service config
│   └── api-gateway.yaml           # API Gateway config
│
├── services/                       # Microservices implementation
│   └── auth/                      # Auth microservice ✅ COMPLETE
│       ├── cmd/server/            # Entry point
│       │   └── main.go            # Server initialization
│       ├── internal/              # Private application code
│       │   ├── config/            # Configuration loading
│       │   ├── domain/            # Domain entities & repository interfaces
│       │   ├── usecase/           # Business logic
│       │   ├── repository/        # Data access implementations
│       │   │   ├── postgres/      # PostgreSQL (User, OAuth)
│       │   │   └── redis/         # Redis (Session, Rate Limiting)
│       │   └── delivery/grpc/     # gRPC handlers
│       └── pkg/                   # Public utilities
│           ├── logger/            # Structured logging (Zap)
│           ├── metrics/           # Prometheus metrics
│           └── tracing/           # OpenTelemetry tracing
│
├── internal/                       # Legacy monolith code
├── bin/                           # Compiled binaries
├── Makefile                       # Build automation
├── go.mod                         # Go module dependencies
└── docker-compose.yml             # Infrastructure services
```

## 🎯 Auth Service - Implementation Details

### Clean Architecture Layers

**1. Domain Layer** (`internal/domain/`)
- Pure business entities (User, OAuthAccount, Session, LoginAttempt)
- Repository interfaces (no implementations)
- No external dependencies

**2. UseCase Layer** (`internal/usecase/`)
- Business logic for authentication workflows
- JWT token generation and validation
- OAuth flow orchestration (Google, GitHub, Microsoft)
- Password hashing with bcrypt
- Login attempt tracking and account locking
- Methods: Register, Login, ValidateToken, RefreshToken, InitiateOAuth, CompleteOAuth, GetUserProfile, UpdateUserProfile, Logout

**3. Repository Layer** (`internal/repository/`)
- **PostgreSQL:** User and OAuth account persistence
- **Redis:** Session management, login attempt tracking, rate limiting
- Implements domain repository interfaces
- Database schema auto-initialization

**4. Delivery Layer** (`internal/delivery/grpc/`)
- gRPC server implementation
- Protobuf message conversion
- Error handling and status codes
- Request context propagation

### Features Implemented

✅ **User Registration & Login**
- Email/password authentication
- Password strength validation
- Bcrypt hashing (cost: 12)
- Email uniqueness check

✅ **JWT Token Management**
- Access tokens (15min expiry)
- Refresh tokens (7 day expiry)
- Token validation with signature verification
- Automatic token refresh
- Multiple session support

✅ **OAuth Integration**
- Google OAuth 2.0
- GitHub OAuth
- Microsoft Azure AD
- OAuth state management
- Account linking (multiple providers per user)
- Token refresh for OAuth

✅ **Security Features**
- Rate limiting (5 failed attempts → 30min lockout)
- Secure session storage in Redis
- JWT secret from environment variables
- Password complexity requirements
- Protection against timing attacks

✅ **Observability**
- Structured JSON logging with Zap
- Prometheus metrics:
  - `grpc_requests_total` - Total gRPC requests
  - `grpc_request_duration_seconds` - Request latency
  - `auth_registrations_total` - User registrations
  - `auth_logins_total` - Login attempts (success/failure)
  - `auth_oauth_logins_total` - OAuth logins by provider
  - `auth_active_sessions` - Active session count
  - `auth_token_validations_total` - Token validation results
- OpenTelemetry tracing with Jaeger
- Request ID propagation
- gRPC interceptors for logging, metrics, tracing

✅ **Resilience**
- Graceful shutdown with timeout
- Database connection pooling
- Redis connection pooling
- Health check endpoint
- Automatic session cleanup (TTL-based)

### API Methods (gRPC)

| Method | Description | Request | Response |
|--------|-------------|---------|----------|
| `Register` | Create new user account | Email, Password, FirstName, LastName | User |
| `Login` | Authenticate user | Email, Password, UserAgent, IP | AccessToken, RefreshToken, User |
| `ValidateToken` | Verify JWT token | Token | Valid (bool), UserID |
| `RefreshToken` | Get new access token | RefreshToken | AccessToken, RefreshToken |
| `InitiateOAuth` | Start OAuth flow | Provider, State | AuthURL |
| `CompleteOAuth` | Finish OAuth flow | Provider, Code | AccessToken, RefreshToken, User |
| `GetUserProfile` | Get user details | UserID | User |
| `UpdateUserProfile` | Update user info | UserID, FirstName, LastName, AvatarURL | User |
| `Logout` | Invalidate session | UserID, AccessToken | Success |
| `HealthCheck` | Service health | - | Status, Service |

### Configuration

See `configs/auth.yaml`:

```yaml
server:
  port: 50051                    # gRPC port
  grpc:
    max_recv_msg_size: 4194304   # 4MB
    max_send_msg_size: 4194304
    keepalive:
      time: 30s
      timeout: 10s
  shutdown_timeout: 30s

jwt:
  access_token_expiry: 15m
  refresh_token_expiry: 168h     # 7 days
  issuer: neighbourhood-platform
  audience: neighbourhood-api

oauth:
  providers:
    google:
      enabled: true
      scopes: [profile, email, openid]
    github:
      enabled: true
      scopes: [user:email]
    microsoft:
      enabled: true
      scopes: [User.Read]

security:
  bcrypt_cost: 12
  max_login_attempts: 5
  lockout_duration: 30m
  password_min_length: 8
  require_special_char: true
  require_number: true
  require_uppercase: true

metrics:
  enabled: true
  port: 9091
  path: /metrics

tracing:
  enabled: true
  service_name: auth-service
  sample_rate: 0.3              # 30% sampling
```

## 🔜 Next Steps

### Phase 2.1: Remaining Microservices ⏳

**Integration Service** (Same clean architecture pattern):
```
services/integration/
├── cmd/server/main.go
├── internal/
│   ├── config/
│   ├── domain/              # Provider, Token, UserIntegration entities
│   ├── usecase/             # ExecuteAction, ConnectProvider, RefreshToken
│   ├── repository/postgres/ # Provider configs, user integrations
│   └── delivery/grpc/       # gRPC handlers
└── pkg/
    ├── providers/           # 45+ provider implementations
    │   ├── slack/
    │   ├── gmail/
    │   ├── jira/
    │   └── ...
    └── ratelimit/           # Per-provider rate limiting
```

**Workflow Service:**
```
services/workflow/
├── internal/
│   ├── domain/              # Workflow, WorkflowStep, Execution
│   ├── usecase/             # ExecuteWorkflow, StepOrchestration
│   ├── repository/postgres/ # Workflow storage
│   └── engine/              # Step execution, parallel processing
└── pkg/
    ├── executor/            # Runtime execution engine
    └── scheduler/           # Cron-like scheduling
```

**Consent Service:**
```
services/consent/
├── internal/
│   ├── domain/              # Consent, AuditLog, Scope
│   ├── usecase/             # GrantConsent, ValidateConsent, GDPR export
│   └── repository/postgres/ # Consent storage with full audit trail
```

**Notification Service:**
```
services/notification/
├── internal/
│   ├── domain/              # Notification, Webhook, DeliveryResult
│   ├── usecase/             # SendNotification, webhook delivery
│   └── repository/mongodb/  # Notification persistence
└── pkg/
    ├── channels/            # Email, SMS, Push, In-App
    │   ├── email/           # SMTP, SendGrid, AWS SES
    │   ├── sms/             # Twilio, AWS SNS
    │   └── push/            # FCM, APNS
    └── delivery/            # Retry logic, circuit breaker
```

### Phase 2.2: API Gateway 🌐

HTTP to gRPC translation layer:
```
services/gateway/
├── cmd/server/main.go
├── internal/
│   ├── router/              # HTTP route definitions
│   ├── middleware/          # Auth, CORS, Rate Limiting, Logging
│   ├── handler/             # HTTP handlers → gRPC calls
│   └── client/              # gRPC client connections
└── pkg/
    ├── auth/                # JWT validation interceptor
    └── ratelimit/           # Token bucket algorithm
```

Features:
- RESTful API endpoints mapped to gRPC services
- OpenAPI/Swagger documentation
- CORS configuration
- Request validation
- API key authentication
- Multi-service request aggregation
- Response caching (Redis)

### Phase 2.3: Deployment & Infrastructure 🐳

**Docker Compose** (Development):
```yaml
services:
  # Databases
  postgres:
  redis:
  mongodb:
  
  # Observability
  consul:
  jaeger:
  prometheus:
  grafana:
  
  # Microservices
  auth-service:
  integration-service:
  workflow-service:
  consent-service:
  notification-service:
  api-gateway:
```

**Kubernetes** (Production):
- Helm charts for each service
- ConfigMaps for configuration
- Secrets for sensitive data
- HorizontalPodAutoscaler
- Service meshes (Istio/Linkerd)
- Ingress controller (NGINX)

**CI/CD Pipeline**:
```
GitHub Actions:
1. Lint & Format → golangci-lint
2. Unit Tests → go test with coverage
3. Build → Docker multi-stage builds
4. Integration Tests → Testcontainers
5. Security Scan → Trivy, gosec
6. Deploy → Kubernetes manifests
```

### Phase 2.4: Advanced Features 🚀

**Service Mesh:**
- mTLS between services
- Traffic management (canary, blue-green)
- Circuit breaking at infrastructure level
- Distributed tracing correlation

**Caching Strategy:**
- Redis for:
  - User sessions
  - OAuth states
  - Rate limiting counters
  - Workflow execution state
  - Consent validation cache

**Message Queue:**
- RabbitMQ/Kafka for:
  - Async workflow execution
  - Notification delivery
  - Integration event streaming
  - Audit log processing

**Monitoring & Alerts:**
- Grafana dashboards per service
- Alertmanager for critical alerts
- PagerDuty integration
- SLO/SLI tracking

## 🧪 Testing Strategy

### Unit Tests
```bash
# Test specific service
go test ./services/auth/internal/usecase/...

# With coverage
go test -cover ./services/auth/...
```

### Integration Tests
```bash
# Using testcontainers
go test -tags=integration ./services/auth/test/integration/...
```

### End-to-End Tests
```bash
# Full workflow testing
go test -tags=e2e ./test/e2e/...
```

### Load Testing
```bash
# gRPC load test with ghz
ghz --insecure \
  --proto proto/auth.proto \
  --call auth.AuthService.Login \
  -d '{"email":"test@example.com","password":"password123"}' \
  -n 10000 -c 100 \
  localhost:50051
```

## 📊 Monitoring

### Prometheus Queries

```promql
# Request rate
rate(grpc_requests_total[5m])

# Error rate
rate(grpc_requests_total{status!="OK"}[5m])

# P95 latency
histogram_quantile(0.95, rate(grpc_request_duration_seconds_bucket[5m]))

# Active sessions
auth_active_sessions

# Login success rate
rate(auth_logins_total{status="success"}[5m]) / rate(auth_logins_total[5m])
```

### Grafana Dashboards

Import community dashboards:
- gRPC Server Metrics (ID: 12239)
- Node Exporter Full (ID: 1860)
- PostgreSQL Database (ID: 9628)

### Jaeger Tracing

Access UI: `http://localhost:16686`

Trace example:
```
API Gateway → Auth Service → PostgreSQL
            → Integration Service → External API
            → Workflow Service → Redis
```

## 🔒 Security Best Practices

1. **Secrets Management:**
   - Use HashiCorp Vault or AWS Secrets Manager
   - Never commit `.env` files
   - Rotate credentials regularly

2. **Network Security:**
   - mTLS between microservices
   - TLS for external connections
   - Service mesh for zero-trust networking

3. **Authentication:**
   - JWT with short expiry
   - Refresh token rotation
   - OAuth with PKCE flow

4. **Authorization:**
   - Role-Based Access Control (RBAC)
   - Scope-based permissions
   - Consent validation on every action

5. **Data Protection:**
   - Encrypt data at rest (PostgreSQL encryption)
   - Encrypt data in transit (TLS 1.3)
   - PII masking in logs
   - GDPR compliance (consent service)

## 📚 Additional Documentation

- [MICROSERVICES_ARCHITECTURE.md](MICROSERVICES_ARCHITECTURE.md) - Detailed architecture design
- [API_DOCUMENTATION.md](API_DOCUMENTATION.md) - Legacy monolith API docs
- [DEPLOYMENT.md](DEPLOYMENT.md) - Deployment guides
- [CONTRIBUTING.md](CONTRIBUTING.md) - Development guidelines

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## 📄 License

Private project - All rights reserved.

---

**Status:** Phase 2.1 - Auth Service Complete ✅ | Next: Integration Service Implementation
