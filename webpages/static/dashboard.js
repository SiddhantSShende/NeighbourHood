// NeighbourHood Dashboard - Production-Grade Frontend Logic
// Time Complexity: O(1) for most operations, O(n) for list rendering
// Space Complexity: O(n) where n is number of integrations/keys

// Configuration
const API_BASE = window.location.origin;
let currentWorkspace = null;
let currentUser = null;

// XSS-safe HTML escaping helper - O(n) string length
function escapeHTML(str) {
    if (str == null) return '';
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#x27;');
}

// Integration Catalog with Requirements
const INTEGRATIONS_CATALOG = [
    {
        type: "slack",
        name: "Slack",
        category: "Communication",
        icon: "💬",
        description: "Send messages, manage channels, and notifications",
        requiredScopes: ["chat:write", "channels:read", "users:read"],
        setupFields: [
            { name: "workspace_url", label: "Slack Workspace URL", type: "text", required: true, placeholder: "your-workspace.slack.com" },
            { name: "default_channel", label: "Default Channel", type: "text", required: false, placeholder: "#general" }
        ]
    },
    {
        type: "gmail",
        name: "Gmail",
        category: "Email",
        icon: "✉️",
        description: "Send emails, read inbox, manage labels",
        requiredScopes: ["gmail.send", "gmail.readonly", "gmail.labels"],
        setupFields: [
            { name: "default_from", label: "Default From Email", type: "email", required: true }
        ]
    },
    {
        type: "github",
        name: "GitHub",
        category: "Developer Tools",
        icon: "🐙",
        description: "Manage repositories, issues, pull requests",
        requiredScopes: ["repo", "read:org", "workflow"],
        setupFields: [
            { name: "organization", label: "Organization/Username", type: "text", required: false }
        ]
    },
    {
        type: "jira",
        name: "Jira",
        category: "Project Management",
        icon: "📋",
        description: "Create issues, update tickets, track projects",
        requiredScopes: ["read:jira-work", "write:jira-work"],
        setupFields: [
            { name: "site_url", label: "Jira Site URL", type: "text", required: true, placeholder: "yourcompany.atlassian.net" },
            { name: "default_project", label: "Default Project Key", type: "text", required: false, placeholder: "PROJ" }
        ]
    },
    {
        type: "notion",
        name: "Notion",
        category: "Productivity",
        icon: "📝",
        description: "Create pages, update databases, manage workspaces",
        requiredScopes: ["read_content", "update_content", "insert_content"],
        setupFields: []
    },
    {
        type: "google_calendar",
        name: "Google Calendar",
        category: "Productivity",
        icon: "📅",
        description: "Create events, manage calendars, send invites",
        requiredScopes: ["calendar.events", "calendar.readonly"],
        setupFields: [
            { name: "default_calendar", label: "Default Calendar", type: "text", required: false, placeholder: "primary" }
        ]
    },
    {
        type: "stripe",
        name: "Stripe",
        category: "Payments",
        icon: "💳",
        description: "Process payments, manage customers, subscriptions",
        requiredScopes: ["read_write"],
        setupFields: [
            { name: "webhook_url", label: "Webhook URL (optional)", type: "url", required: false }
        ]
    },
    {
        type: "salesforce",
        name: "Salesforce",
        category: "CRM",
        icon: "☁️",
        description: "Manage leads, opportunities, and customer data",
        requiredScopes: ["api", "full"],
        setupFields: [
            { name: "instance_url", label: "Instance URL", type: "text", required: true, placeholder: "https://yourinstance.salesforce.com" }
        ]
    }
];

// SSO Login Functions
window.loginWithGoogle = () => {
    window.location.href = `${API_BASE}/auth/google/login`;
};

window.loginWithGitHub = () => {
    window.location.href = `${API_BASE}/auth/github/login`;
};

// OAuth Callback Handling — runs before DOMContentLoaded to store token early
(function handleOAuthCallback() {
    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get('token');
    if (token) {
        localStorage.setItem('nh_token', token);
        // Remove token from URL without losing current path
        const cleanUrl = window.location.pathname + window.location.hash;
        window.history.replaceState({}, document.title, cleanUrl);
    }
}());

// Initialize Dashboard
document.addEventListener('DOMContentLoaded', async () => {
    const loginForm = document.getElementById('login-form');
    const loginView = document.getElementById('login-view');
    const dashboardView = document.getElementById('dashboard-view');
    const authNav = document.getElementById('auth-nav');

    // Check for existing token
    const savedToken = localStorage.getItem('nh_token');
    if (savedToken) {
        await loadDashboard();
    }

    // Login Form Handler
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;

        try {
            const res = await fetch(`${API_BASE}/api/auth/login`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password })
            });

            if (res.ok) {
                const data = await res.json();
                localStorage.setItem('nh_token', data.access_token);
                await loadDashboard();
            } else {
                const error = await res.json();
                alert(error.message || 'Login failed');
            }
        } catch (err) {
            console.error('Login error:', err);
            alert('Error logging in. Please try again.');
        }
    });

    async function loadDashboard() {
        loginView.style.display = 'none';
        dashboardView.style.display = 'block';
        
        authNav.innerHTML = `
            <button class="btn-secondary btn-sm" onclick="logout()">Sign Out</button>
        `;

        // Load user data and workspaces
        await loadUserData();
        await loadWorkspaces();
        await loadIntegrations();
    }
});

// Load User Data - O(1) API call
async function loadUserData() {
    try {
        const res = await apiCall('/api/auth/profile');
        currentUser = res.user;
    } catch (err) {
        console.error('Failed to load user data:', err);
    }
}

// Load Workspaces - O(k) where k is number of workspaces
async function loadWorkspaces() {
    try {
        const res = await apiCall('/api/workspaces');
        const workspaces = res.workspaces || [];
        
        const selector = document.getElementById('workspace-selector');
        
        if (workspaces.length === 0) {
            // Create default workspace
            await createDefaultWorkspace();
            return;
        }
        
        selector.innerHTML = workspaces.map(ws => 
            `<option value="${escapeHTML(ws.id)}">${escapeHTML(ws.name)}</option>`
        ).join('');
        
        currentWorkspace = workspaces[0];
        selector.value = currentWorkspace.id;
        
        // Load workspace data
        await loadWorkspaceData(currentWorkspace.id);
        
        selector.addEventListener('change', async (e) => {
            const selectedId = e.target.value;
            currentWorkspace = workspaces.find(ws => ws.id === selectedId);
            await loadWorkspaceData(selectedId);
        });
    } catch (err) {
        console.error('Failed to load workspaces:', err);
    }
}

// Create Default Workspace
async function createDefaultWorkspace() {
    try {
        const res = await apiCall('/api/workspaces', {
            method: 'POST',
            body: JSON.stringify({
                name: 'My Workspace',
                description: 'Default workspace'
            })
        });
        
        currentWorkspace = res.workspace;
        await loadWorkspaces();
    } catch (err) {
        console.error('Failed to create workspace:', err);
    }
}

// Load Workspace Data - O(1) for API keys, O(n) for integrations
async function loadWorkspaceData(workspaceId) {
    try {
        // Load API Keys
        const keysRes = await apiCall(`/api/workspaces/${workspaceId}/api-keys`);
        displayAPIKeys(keysRes.api_keys || []);
        
        // Load Active Integrations
        const integrationsRes = await apiCall(`/api/workspaces/${workspaceId}/integrations`);
        displayActiveIntegrations(integrationsRes.integrations || []);
    } catch (err) {
        console.error('Failed to load workspace data:', err);
    }
}

// Display API Keys - O(n) rendering
function displayAPIKeys(keys) {
    const container = document.getElementById('api-keys-list');
    
    if (keys.length === 0) {
        container.innerHTML = `
            <div style="text-align: center; padding: 2rem; color: var(--text-secondary);">
                <p>No API keys yet. Create your first one to get started!</p>
            </div>
        `;
        return;
    }
    
    container.innerHTML = keys.map(key => `
        <div class="api-key-box">
            <div class="api-key-label">${escapeHTML(key.name)}</div>
            <div class="api-key-value">
                <code>${escapeHTML(key.key_prefix)}••••••••••••••••</code>
                <button class="btn-outline btn-sm" data-key-id="${escapeHTML(key.id)}">Revoke</button>
            </div>
            <div class="flex justify-between items-center" style="margin-top: 0.75rem;">
                <span class="text-sm text-secondary">
                    Last used: ${key.last_used_at ? new Date(key.last_used_at).toLocaleDateString() : 'Never'}
                </span>
                <span class="status-badge ${key.active ? 'status-connected' : 'status-disconnected'}">
                    ${key.active ? 'Active' : 'Revoked'}
                </span>
            </div>
        </div>
    `).join('');

    // Bind revoke buttons safely — avoid inline onclick with server data
    container.querySelectorAll('[data-key-id]').forEach(btn => {
        btn.addEventListener('click', () => revokeAPIKey(btn.dataset.keyId));
    });
}

// Load and Display Available Integrations - O(n)
function loadIntegrations() {
    const grid = document.getElementById('integrations-grid');
    const searchInput = document.getElementById('integration-search');
    
    function renderIntegrations(filter = '') {
        const filtered = INTEGRATIONS_CATALOG.filter(int => 
            int.name.toLowerCase().includes(filter.toLowerCase()) ||
            int.category.toLowerCase().includes(filter.toLowerCase())
        );
        
        grid.innerHTML = filtered.map(integration => `
            <div class="integration-item">
                <div class="integration-info">
                    <div class="integration-icon">${integration.icon}</div>
                    <div class="integration-details">
                        <h3>${integration.name}</h3>
                        <p>${integration.description}</p>
                        <span class="text-sm text-purple" style="font-weight: 600;">${integration.category}</span>
                    </div>
                </div>
                <div class="integration-actions">
                    <button class="btn-sm" onclick='showIntegrationModal(${JSON.stringify(integration)})'>
                        Connect
                    </button>
                </div>
            </div>
        `).join('');
    }
    
    renderIntegrations();
    
    searchInput.addEventListener('input', (e) => {
        renderIntegrations(e.target.value);
    });
}

// Show Integration Signup Modal
window.showIntegrationModal = (integration) => {
    const modal = document.getElementById('integration-modal');
    const title = document.getElementById('modal-title');
    const body = document.getElementById('modal-body');
    
    title.textContent = `Connect to ${integration.name}`;
    
    body.innerHTML = `
        <div style="margin-bottom: 1.5rem;">
            <div style="display: flex; align-items: center; gap: 1rem; margin-bottom: 1rem;">
                <div class="integration-icon" style="width: 56px; height: 56px; font-size: 1.75rem;">${integration.icon}</div>
                <div>
                    <h4 style="font-size: 1.125rem; font-weight: 600; margin-bottom: 0.25rem;">${integration.name}</h4>
                    <p class="text-secondary text-sm">${integration.description}</p>
                </div>
            </div>
        </div>
        
        <div class="card" style="background: var(--purple-50); border-color: var(--purple-500); margin-bottom: 1.5rem;">
            <h4 style="font-size: 0.938rem; font-weight: 600; margin-bottom: 0.75rem;">Required Permissions</h4>
            <div style="display: flex; flex-wrap: gap; gap: 0.5rem;">
                ${integration.requiredScopes.map(scope => `
                    <span class="status-badge status-connected">${scope}</span>
                `).join('')}
            </div>
        </div>
        
        <form id="integration-setup-form" onsubmit="connectIntegration(event, '${integration.type}')">
            ${integration.setupFields.map(field => `
                <label for="field-${field.name}">${field.label}${field.required ? ' *' : ''}</label>
                <input 
                    type="${field.type}" 
                    id="field-${field.name}" 
                    name="${field.name}"
                    placeholder="${field.placeholder || ''}"
                    ${field.required ? 'required' : ''}
                >
            `).join('')}
            
            <div class="card" style="background: var(--warning-bg); margin-top: 1rem;">
                <p class="text-sm" style="color: var(--warning-text);">
                    ⚠️ You'll be redirected to ${integration.name} to authorize access. Make sure you're logged in to your ${integration.name} account.
                </p>
            </div>
            
            <div class="flex gap-2" style="margin-top: 1.5rem;">
                <button type="submit" style="flex: 1;">
                    Authorize & Connect
                </button>
                <button type="button" class="btn-secondary" onclick="closeIntegrationModal()">
                    Cancel
                </button>
            </div>
        </form>
    `;
    
    modal.style.display = 'flex';
};

// Close Integration Modal
window.closeIntegrationModal = () => {
    document.getElementById('integration-modal').style.display = 'none';
};

// Connect Integration - O(1) API call
window.connectIntegration = async (event, integrationType) => {
    event.preventDefault();
    
    const form = event.target;
    const formData = new FormData(form);
    const config = {};
    
    for (let [key, value] of formData.entries()) {
        config[key] = value;
    }
    
    try {
        const res = await apiCall('/api/integrations/connect', {
            method: 'POST',
            body: JSON.stringify({
                workspace_id: currentWorkspace.id,
                integration_type: integrationType,
                config: config
            })
        });
        
        // Redirect to OAuth
        if (res.auth_url) {
            window.location.href = res.auth_url;
        }
    } catch (err) {
        alert('Failed to connect integration: ' + err.message);
    }
};

// Display Active Integrations - O(n)
function displayActiveIntegrations(integrations) {
    const container = document.getElementById('active-integrations-list');
    
    if (integrations.length === 0) {
        container.innerHTML = `
            <p class="text-secondary" style="text-align: center; padding: 2rem;">
                No active integrations yet. Connect your first integration above!
            </p>
        `;
        return;
    }
    
    container.innerHTML = integrations.map(int => {
        const catalog = INTEGRATIONS_CATALOG.find(c => c.type === int.integration_type);
        return `
            <div class="integration-item">
                <div class="integration-info">
                    <div class="integration-icon">${catalog?.icon || '🔗'}</div>
                    <div class="integration-details">
                        <h3>${escapeHTML(catalog?.name || int.integration_type)}</h3>
                        <p class="text-sm">Connected ${new Date(int.connected_at).toLocaleDateString()}</p>
                    </div>
                </div>
                <div class="integration-actions">
                    <span class="status-badge status-connected">Connected</span>
                    <button class="btn-outline btn-sm" data-integration-id="${escapeHTML(int.id)}">
                        Disconnect
                    </button>
                </div>
            </div>
        `;
    }).join('');

    // Bind disconnect buttons safely
    container.querySelectorAll('[data-integration-id]').forEach(btn => {
        btn.addEventListener('click', () => disconnectIntegration(btn.dataset.integrationId));
    });
}

// API Key Management
window.showCreateAPIKeyModal = () => {
    document.getElementById('api-key-modal').style.display = 'flex';
    
    document.getElementById('key-expiry-enabled').addEventListener('change', (e) => {
        document.getElementById('key-expiry-date').style.display = e.target.checked ? 'block' : 'none';
    });
};

window.closeAPIKeyModal = () => {
    document.getElementById('api-key-modal').style.display = 'none';
    document.getElementById('create-api-key-form').reset();
};

window.createAPIKey = async (event) => {
    event.preventDefault();
    
    const name = document.getElementById('key-name').value;
    const rateLimit = parseInt(document.getElementById('key-rate-limit').value);
    const expiryEnabled = document.getElementById('key-expiry-enabled').checked;
    const expiryDate = expiryEnabled ? document.getElementById('key-expiry-date').value : null;
    
    try {
        const res = await apiCall('/api/api-keys', {
            method: 'POST',
            body: JSON.stringify({
                workspace_id: currentWorkspace.id,
                name: name,
                rate_limit: rateLimit,
                expires_at: expiryDate
            })
        });
        
        // Show the generated key inline — only shown once, never via alert
        const form = document.getElementById('create-api-key-form');
        form.innerHTML = `
            <div style="text-align:center;padding:1rem 0;">
                <p style="font-weight:700;font-size:1.0625rem;margin-bottom:0.75rem;color:var(--text-primary);">API Key Created!</p>
                <p style="font-size:0.875rem;color:var(--text-secondary);margin-bottom:1rem;">
                    Copy and save this key now — it won't be shown again.
                </p>
                <div style="display:flex;gap:0.5rem;align-items:center;background:var(--bg-secondary);border:1px solid var(--border-color);border-radius:0.5rem;padding:0.75rem 1rem;margin-bottom:1.25rem;">
                    <code id="new-api-key-display" style="flex:1;word-break:break-all;font-size:0.8125rem;">${escapeHTML(res.api_key)}</code>
                    <button type="button" id="copy-new-key-btn" class="btn-outline btn-sm" style="flex-shrink:0;">Copy</button>
                </div>
                <button type="button" class="btn-primary" onclick="closeAPIKeyModal();loadWorkspaceData(currentWorkspace.id);">Done</button>
            </div>
        `;
        document.getElementById('copy-new-key-btn').addEventListener('click', async () => {
            try {
                await navigator.clipboard.writeText(res.api_key);
                document.getElementById('copy-new-key-btn').textContent = 'Copied!';
            } catch {
                document.getElementById('copy-new-key-btn').textContent = 'Failed';
            }
        });
    } catch (err) {
        alert('Failed to create API key: ' + err.message);
    }
};

window.revokeAPIKey = async (keyId) => {
    if (!confirm('Are you sure you want to revoke this API key? This action cannot be undone.')) {
        return;
    }
    
    try {
        await apiCall(`/api/api-keys/${keyId}/revoke`, { method: 'POST' });
        await loadWorkspaceData(currentWorkspace.id);
    } catch (err) {
        alert('Failed to revoke API key: ' + err.message);
    }
};

window.copyAPIKey = async () => {
    const keyElement = document.getElementById('api-key-display');
    const btn = document.querySelector('[onclick="copyAPIKey()"]');
    try {
        await navigator.clipboard.writeText(keyElement.textContent.trim());
        if (btn) { const orig = btn.textContent; btn.textContent = 'Copied!'; setTimeout(() => { btn.textContent = orig; }, 2000); }
    } catch {
        // Clipboard API unavailable — select the text as fallback
        const range = document.createRange();
        range.selectNodeContents(keyElement);
        window.getSelection()?.removeAllRanges();
        window.getSelection()?.addRange(range);
    }
};

// Disconnect Integration
window.disconnectIntegration = async (integrationId) => {
    if (!confirm('Are you sure you want to disconnect this integration?')) {
        return;
    }
    
    try {
        await apiCall(`/api/integrations/${integrationId}`, { method: 'DELETE' });
        await loadWorkspaceData(currentWorkspace.id);
    } catch (err) {
        alert('Failed to disconnect integration: ' + err.message);
    }
};

// Logout
window.logout = () => {
    localStorage.removeItem('nh_token');
    window.location.reload();
};

// Utility: API Call with Auth - O(1) for auth header setup
async function apiCall(endpoint, options = {}) {
    const token = localStorage.getItem('nh_token');
    
    const config = {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
            ...options.headers
        }
    };
    
    const res = await fetch(`${API_BASE}${endpoint}`, config);
    
    if (res.status === 401) {
        localStorage.removeItem('nh_token');
        window.location.reload();
        throw new Error('Unauthorized');
    }
    
    if (!res.ok) {
        const error = await res.json();
        throw new Error(error.message || 'API call failed');
    }
    
    return await res.json();
}
