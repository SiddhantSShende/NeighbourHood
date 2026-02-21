-- RBAC System Schema with Optimized Indexes for O(1) Lookups

-- Workspaces table
CREATE TABLE IF NOT EXISTS workspaces (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    owner_id VARCHAR(255) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    plan VARCHAR(50) DEFAULT 'free', -- free, pro, enterprise
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    settings JSONB DEFAULT '{}'::jsonb,
    
    -- Indexes for O(1) lookups
    CONSTRAINT fk_workspace_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_workspaces_owner_id ON workspaces(owner_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_active ON workspaces(active) WHERE active = true;

-- User Roles table (RBAC)
CREATE TABLE IF NOT EXISTS user_roles (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    workspace_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL, -- admin, developer, user, viewer
    permissions JSONB DEFAULT '[]'::jsonb, -- Custom permissions array
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    
    -- Composite unique constraint
    CONSTRAINT unique_user_workspace UNIQUE(user_id, workspace_id),
    
    -- Foreign keys
    CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

-- Critical indexes for O(1) permission checks
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_composite ON user_roles(user_id, workspace_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_workspace ON user_roles(workspace_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role);

-- API Keys table with hash-based O(1) lookup
CREATE TABLE IF NOT EXISTS api_keys (
    id VARCHAR(255) PRIMARY KEY,
    workspace_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(50) NOT NULL UNIQUE, -- nh_live_pk_abc123
    key_hash VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 hash for validation
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    active BOOLEAN DEFAULT true,
    scopes JSONB DEFAULT '[]'::jsonb,
    rate_limit INTEGER DEFAULT 1000, -- requests per hour
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP,
    
    -- Foreign keys
    CONSTRAINT fk_api_keys_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    CONSTRAINT fk_api_keys_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Hash index for O(1) API key validation (most critical operation)
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash) WHERE active = true;
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_workspace ON api_keys(workspace_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(active) WHERE active = true;

-- Integration Subscriptions (what integrations a user has signed up for)
CREATE TABLE IF NOT EXISTS integration_subscriptions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    workspace_id VARCHAR(255) NOT NULL,
    integration_type VARCHAR(100) NOT NULL, -- slack, gmail, etc.
    status VARCHAR(50) DEFAULT 'pending', -- pending, active, disconnected
    config JSONB DEFAULT '{}'::jsonb, -- Integration-specific configuration
    required_scopes JSONB DEFAULT '[]'::jsonb,
    granted_scopes JSONB DEFAULT '[]'::jsonb,
    oauth_token_id VARCHAR(255), -- Reference to OAuth token
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    connected_at TIMESTAMP,
    last_used_at TIMESTAMP,
    
    -- Composite unique constraint
    CONSTRAINT unique_user_integration UNIQUE(user_id, workspace_id, integration_type),
    
    -- Foreign keys
    CONSTRAINT fk_integration_sub_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_integration_sub_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

-- Optimized indexes for integration lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_integration_sub_composite ON integration_subscriptions(user_id, workspace_id, integration_type);
CREATE INDEX IF NOT EXISTS idx_integration_sub_user ON integration_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_integration_sub_workspace ON integration_subscriptions(workspace_id);
CREATE INDEX IF NOT EXISTS idx_integration_sub_type ON integration_subscriptions(integration_type);
CREATE INDEX IF NOT EXISTS idx_integration_sub_status ON integration_subscriptions(status) WHERE status = 'active';

-- API Usage Analytics (for rate limiting and monitoring)
CREATE TABLE IF NOT EXISTS api_usage (
    id BIGSERIAL PRIMARY KEY,
    api_key_id VARCHAR(255) NOT NULL,
    workspace_id VARCHAR(255) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    
    -- Foreign keys
    CONSTRAINT fk_api_usage_key FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE,
    CONSTRAINT fk_api_usage_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

-- Time-series indexes for analytics
CREATE INDEX IF NOT EXISTS idx_api_usage_timestamp ON api_usage(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_api_usage_api_key ON api_usage(api_key_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_api_usage_workspace ON api_usage(workspace_id, timestamp DESC);

-- Rate Limiting Cache (Redis-backed table for high-performance rate limiting)
CREATE TABLE IF NOT EXISTS rate_limits (
    key VARCHAR(255) PRIMARY KEY, -- api_key_id:hour
    count INTEGER DEFAULT 0,
    window_start TIMESTAMP NOT NULL,
    window_end TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rate_limits_window ON rate_limits(window_end) WHERE window_end > CURRENT_TIMESTAMP;

-- Audit Log for security and compliance
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255),
    workspace_id VARCHAR(255),
    action VARCHAR(100) NOT NULL, -- create_api_key, revoke_key, grant_permission, etc.
    resource_type VARCHAR(100), -- api_key, workspace, user_role
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_workspace ON audit_logs(workspace_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);

-- Performance optimization: Create materialized view for workspace stats
CREATE MATERIALIZED VIEW IF NOT EXISTS workspace_stats AS
SELECT 
    w.id AS workspace_id,
    w.name,
    COUNT(DISTINCT ur.user_id) AS member_count,
    COUNT(DISTINCT ak.id) AS api_key_count,
    COUNT(DISTINCT isub.id) AS integration_count,
    MAX(au.timestamp) AS last_activity
FROM workspaces w
LEFT JOIN user_roles ur ON w.id = ur.workspace_id
LEFT JOIN api_keys ak ON w.id = ak.workspace_id AND ak.active = true
LEFT JOIN integration_subscriptions isub ON w.id = isub.workspace_id AND isub.status = 'active'
LEFT JOIN api_usage au ON w.id = au.workspace_id
GROUP BY w.id, w.name;

CREATE UNIQUE INDEX IF NOT EXISTS idx_workspace_stats_id ON workspace_stats(workspace_id);

-- Refresh materialized view periodically (can be done via cron job)
-- REFRESH MATERIALIZED VIEW CONCURRENTLY workspace_stats;

-- Comments for documentation
COMMENT ON TABLE workspaces IS 'Developer workspaces for organizing integrations and team members';
COMMENT ON TABLE user_roles IS 'Role-Based Access Control (RBAC) - Maps users to workspaces with specific roles and permissions';
COMMENT ON TABLE api_keys IS 'Unique API keys for developers with SHA-256 hashing for security';
COMMENT ON TABLE integration_subscriptions IS 'Tracks which integrations users have signed up for with their configuration';
COMMENT ON TABLE api_usage IS 'API call analytics for monitoring and rate limiting';
COMMENT ON TABLE audit_logs IS 'Security audit trail for compliance and debugging';
