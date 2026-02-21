# NeighbourHood Microservices Architecture

## Architecture Overview

The NeighbourHood platform has been redesigned as a microservices-based system using gRPC for inter-service communication, Protocol Buffers for serialization, and following clean architecture principles.

## Microservices Structure

```
neighbourhood/
├── api-gateway/          # HTTP/REST to gRPC gateway
├── services/
│   ├── auth/            # Authentication & Authorization Service
│   ├── integration/     # Integration Provider Service
│   ├── workflow/        # Workflow Engine Service
│   ├── consent/         # Consent Management Service
│   └── notification/    # Notification Service
├── proto/               # Protocol Buffer Definitions
├── pkg/                 # Shared libraries
│   ├── config/         # Configuration management
│   ├── logger/         # Structured logging
│   ├── metrics/        # Prometheus metrics
│   ├── tracing/        # OpenTelemetry tracing
│   └── database/       # Database utilities
├── configs/            # External configuration files
├── deployments/        # Kubernetes/Docker configs
└── scripts/            # Build and deployment scripts
```

## Core Microservices

### 1. API Gateway Service
**Port**: 8080 (HTTP/REST)  
**Purpose**: Public-facing REST API, converts HTTP to gRPC  
**Responsibilities**:
- REST API endpoints
- Request validation
- Authentication middleware
- Rate limiting
- Request routing to services
- Response aggregation

### 2. Auth Service
**Port**: 50051 (gRPC)  
**Purpose**: Authentication and user management  
**Responsibilities**:
- User registration and login
- OAuth 2.0 flow (Google, GitHub, etc.)
- JWT token generation and validation
- Session management
- User profile management
- Permission management

**gRPC Methods**:
```protobuf
service AuthService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc InitiateOAuth(OAuthRequest) returns (OAuthResponse);
  rpc CompleteOAuth(OAuthCallbackRequest) returns (OAuthCallbackResponse);
  rpc GetUserProfile(GetUserRequest) returns (UserProfile);
}
```

### 3. Integration Service
**Port**: 50052 (gRPC)  
**Purpose**: Manage integration providers and execute actions  
**Responsibilities**:
- Provider registration and discovery
- OAuth URL generation
- Token exchange
- Action execution for all providers
- Provider health monitoring
- Rate limiting per provider

**gRPC Methods**:
```protobuf
service IntegrationService {
  rpc ListProviders(ListProvidersRequest) returns (ListProvidersResponse);
  rpc GetAuthURL(GetAuthURLRequest) returns (GetAuthURLResponse);
  rpc ExchangeCode(ExchangeCodeRequest) returns (ExchangeCodeResponse);
  rpc ExecuteAction(ExecuteActionRequest) returns (ExecuteActionResponse);
  rpc GetProviderStatus(ProviderStatusRequest) returns (ProviderStatusResponse);
  rpc RevokeIntegration(RevokeIntegrationRequest) returns (RevokeIntegrationResponse);
}
```

### 4. Workflow Engine Service
**Port**: 50053 (gRPC)  
**Purpose**: Orchestrate multi-step workflows  
**Responsibilities**:
- Workflow definition management
- Workflow execution
- Step orchestration
- Error handling and retry logic
- Workflow state management
- Event-driven execution

**gRPC Methods**:
```protobuf
service WorkflowService {
  rpc CreateWorkflow(CreateWorkflowRequest) returns (CreateWorkflowResponse);
  rpc GetWorkflow(GetWorkflowRequest) returns (Workflow);
  rpc ListWorkflows(ListWorkflowsRequest) returns (ListWorkflowsResponse);
  rpc ExecuteWorkflow(ExecuteWorkflowRequest) returns (ExecuteWorkflowResponse);
  rpc GetExecutionStatus(ExecutionStatusRequest) returns (ExecutionStatus);
  rpc CancelExecution(CancelExecutionRequest) returns (CancelExecutionResponse);
}
```

### 5. Consent Service
**Port**: 50054 (gRPC)  
**Purpose**: GDPR-compliant consent management  
**Responsibilities**:
- Consent tracking
- Consent validation
- Consent revocation
- Audit trail
- Compliance reporting

**gRPC Methods**:
```protobuf
service ConsentService {
  rpc GrantConsent(GrantConsentRequest) returns (GrantConsentResponse);
  rpc RevokeConsent(RevokeConsentRequest) returns (RevokeConsentResponse);
  rpc ValidateConsent(ValidateConsentRequest) returns (ValidateConsentResponse);
  rpc ListConsents(ListConsentsRequest) returns (ListConsentsResponse);
  rpc GetConsentHistory(ConsentHistoryRequest) returns (ConsentHistoryResponse);
}
```

### 6. Notification Service
**Port**: 50055 (gRPC)  
**Purpose**: Handle webhooks and notifications  
**Responsibilities**:
- Webhook registration
- Event processing
- Notification delivery
- Email/SMS sending
- Push notifications

**gRPC Methods**:
```protobuf
service NotificationService {
  rpc RegisterWebhook(WebhookRequest) returns (WebhookResponse);
  rpc SendNotification(NotificationRequest) returns (NotificationResponse);
  rpc GetNotificationHistory(NotificationHistoryRequest) returns (NotificationHistoryResponse);
}
```

## Communication Patterns

### Synchronous (gRPC)
- API Gateway → Services
- Service-to-Service direct calls
- Request-Response pattern

### Asynchronous (Message Queue)
- Event-driven workflows
- Background job processing
- Inter-service events

**Message Broker**: RabbitMQ / Apache Kafka

## Data Management

### Database Strategy
Each service has its own database (Database per Service pattern):

- **Auth Service**: PostgreSQL (users, sessions)
- **Integration Service**: PostgreSQL (providers, tokens)
- **Workflow Service**: PostgreSQL (workflows, executions)
- **Consent Service**: PostgreSQL (consents, audit logs)
- **Notification Service**: MongoDB (events, notifications)

### Caching Layer
- **Redis** for:
  - Session storage
  - OAuth state management
  - Rate limiting
  - API response caching

## Service Discovery

**Consul** for service registration and discovery:
- Health checks
- Service registry
- Configuration management
- KV store

## Configuration Management

### External Configuration (configs/)
```yaml
# configs/api-gateway.yaml
server:
  port: 8080
  timeout: 30s
  
cors:
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE"]

services:
  auth:
    address: "auth-service:50051"
  integration:
    address: "integration-service:50052"
  workflow:
    address: "workflow-service:50053"
```

### Configuration Sources
1. YAML/JSON files (default configs)
2. Environment variables (overrides)
3. Consul KV (dynamic config)
4. Command-line flags

**Library**: Viper for configuration management

## Observability

### Logging
- **Structured logging** with Zap/Zerolog
- **Log aggregation** with ELK Stack (Elasticsearch, Logstash, Kibana)
- **Correlation IDs** for request tracing across services

### Metrics
- **Prometheus** for metrics collection
- **Grafana** for visualization
- Standard metrics:
  - Request count
  - Response time
  - Error rate
  - Resource utilization

### Distributed Tracing
- **OpenTelemetry** for instrumentation
- **Jaeger** for trace visualization
- Trace propagation across services

### Health Checks
- **Liveness** probes (is service running?)
- **Readiness** probes (can service handle requests?)
- **Startup** probes (has service started?)

## Resilience Patterns

### Circuit Breaker
- Prevent cascading failures
- Automatic retry with exponential backoff
- Library: go-resiliency/circuitbreaker

### Rate Limiting
- Per-user rate limiting
- Per-provider rate limiting
- Token bucket algorithm

### Timeouts
- Request timeouts
- Connection timeouts
- gRPC deadlines

### Retry Logic
- Automatic retry for transient failures
- Exponential backoff
- Maximum retry attempts

## Security

### Authentication & Authorization
- **JWT** tokens for user authentication
- **mTLS** for service-to-service communication
- **API Keys** for external clients
- **RBAC** for permission management

### Secret Management
- **Vault** by HashiCorp
- Encrypted environment variables
- Secret rotation

### Network Security
- Service mesh (Istio/Linkerd)
- Network policies
- TLS everywhere

## Deployment

### Container Orchestration
**Kubernetes** for production:
```
deployments/
├── kubernetes/
│   ├── api-gateway/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── ingress.yaml
│   ├── auth-service/
│   ├── integration-service/
│   ├── workflow-service/
│   ├── consent-service/
│   └── notification-service/
├── docker-compose.yml      # Local development
└── docker-compose.prod.yml # Production-like environment
```

### CI/CD Pipeline
```
GitHub Actions / GitLab CI:
1. Build → Test → Lint
2. Build Docker images
3. Push to container registry
4. Deploy to Kubernetes
5. Run smoke tests
```

## Development Workflow

### Local Development
```bash
# Start all services with docker-compose
docker-compose up -d

# Generate protobuf code
make proto

# Run specific service
cd services/auth
go run cmd/server/main.go

# Run tests
make test

# Build all services
make build
```

### Code Generation
```bash
# Install protoc tools
make install-tools

# Generate gRPC code from proto files
protoc --go_out=. --go-grpc_out=. proto/*.proto
```

## Clean Architecture Principles

### Layered Architecture (per service)
```
services/auth/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── domain/                  # Business entities
│   │   ├── user.go
│   │   └── session.go
│   ├── usecase/                 # Business logic
│   │   ├── register.go
│   │   └── login.go
│   ├── repository/              # Data access
│   │   ├── user_repository.go
│   │   └── postgres/
│   ├── delivery/                # Handlers
│   │   └── grpc/
│   │       └── auth_handler.go
│   └── infrastructure/          # External services
│       ├── oauth/
│       └── jwt/
├── pkg/                         # Public packages
└── proto/                       # Generated protobuf code
```

### Dependency Injection
- Wire or dig for dependency injection
- Constructor-based injection
- Interface-based design

### Repository Pattern
```go
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### Use Case Pattern
```go
type RegisterUseCase struct {
    userRepo UserRepository
    logger   logger.Logger
}

func (uc *RegisterUseCase) Execute(ctx context.Context, req RegisterRequest) (*User, error) {
    // Business logic here
}
```

## Testing Strategy

### Unit Tests
- Test business logic in isolation
- Mock dependencies
- 80%+ code coverage

### Integration Tests
- Test service with real database
- Test gRPC endpoints
- Use testcontainers

### End-to-End Tests
- Test complete workflows
- Test API Gateway → Services
- Use staging environment

### Performance Tests
- Load testing with k6
- Stress testing
- Benchmark critical paths

## Monitoring & Alerting

### Metrics to Monitor
- Service availability (uptime)
- Request rate and latency
- Error rate (4xx, 5xx)
- Database connection pool
- Message queue depth
- CPU/Memory usage

### Alerting Rules
- Service down for > 1 minute
- Error rate > 5%
- Response time > 2s (p95)
- Database connection pool > 80%

## Scalability

### Horizontal Scaling
- Stateless services
- Load balancing (Kubernetes)
- Auto-scaling based on metrics

### Vertical Scaling
- Resource limits and requests
- Right-sizing based on profiling

### Database Scaling
- Read replicas
- Connection pooling
- Query optimization
- Caching layer

## Migration Strategy

### Phase 1: Parallel Run
- Keep monolith running
- Deploy microservices alongside
- Route subset of traffic to microservices

### Phase 2: Gradual Migration
- Migrate one service at a time
- Start with least critical service
- Use feature flags for rollback

### Phase 3: Full Migration
- Route all traffic to microservices
- Decommission monolith
- Data migration complete

## Cost Optimization

- Right-size resources
- Use spot instances for non-critical workloads
- Implement auto-scaling
- Monitor and eliminate waste
- Use caching to reduce API calls

---

**Last Updated**: February 21, 2026  
**Architecture Version**: 2.0.0  
**Status**: Implementation In Progress
