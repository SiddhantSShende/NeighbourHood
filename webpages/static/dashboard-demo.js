// NeighbourHood Dashboard - Client-Side Demo Version
// Works entirely with localStorage (no backend required)
// Perfect for GitHub Pages static hosting

// Demo Configuration
const DEMO_MODE = true;
const DEMO_USER = {
    id: 'demo-user-123',
    email: 'demo@neighbourhood.dev',
    name: 'Demo User',
    created_at: new Date().toISOString()
};

// Integration Catalog
const INTEGRATIONS_CATALOG = [
    {
        type: "slack",
        name: "Slack",
        category: "Communication",
        description: "Send messages, manage channels, and notifications",
        requiredScopes: ["chat:write", "channels:read", "users:read"]
    },
    {
        type: "gmail",
        name: "Gmail",
        category: "Email",
        description: "Send emails, read inbox, manage labels",
        requiredScopes: ["gmail.send", "gmail.readonly", "gmail.labels"]
    },
    {
        type: "github",
        name: "GitHub",
        category: "Developer Tools",
        description: "Manage repositories, issues, pull requests",
        requiredScopes: ["repo", "read:org", "workflow"]
    },
    {
        type: "jira",
        name: "Jira",
        category: "Project Management",
        description: "Create issues, manage projects, track workflows",
        requiredScopes: ["read:jira-work", "write:jira-work"]
    },
    {
        type: "notion",
        name: "Notion",
        category: "Productivity",
        description: "Manage pages, databases, and content",
        requiredScopes: ["read:content", "write:content"]
    },
    {
        type: "salesforce",
        name: "Salesforce",
        category: "CRM",
        description: "Manage leads, contacts, and opportunities",
        requiredScopes: ["api", "full"]
    }
];

// State Management
let currentUser = null;
let currentWorkspace = null;

// Demo Authentication Functions
window.loginWithGoogle = () => {
    showDemoNotice('Google OAuth');
    setTimeout(() => {
        performDemoLogin('Google');
    }, 1500);
};

window.loginWithGitHub = () => {
    showDemoNotice('GitHub OAuth');
    setTimeout(() => {
        performDemoLogin('GitHub');
    }, 1500);
};

function showDemoNotice(provider) {
    const notice = document.createElement('div');
    notice.style.cssText = `
        position: fixed;
        top: 2rem;
        right: 2rem;
        background: linear-gradient(135deg, #7C3AED, #A855F7);
        color: white;
        padding: 1rem 1.5rem;
        border-radius: 0.5rem;
        box-shadow: 0 4px 20px rgba(124, 58, 237, 0.4);
        z-index: 10000;
        animation: slideIn 0.3s ease-out;
    `;
    notice.innerHTML = `
        <div style="font-weight: 600; margin-bottom: 0.25rem;">Demo Mode</div>
        <div style="font-size: 0.875rem; opacity: 0.9;">Simulating ${provider} authentication...</div>
    `;
    
    document.body.appendChild(notice);
    setTimeout(() => notice.remove(), 1500);
}

function performDemoLogin(provider) {
    localStorage.setItem('nh_demo_token', 'demo-token-' + Date.now());
    localStorage.setItem('nh_demo_user', JSON.stringify(DEMO_USER));
    localStorage.setItem('nh_demo_provider', provider);
    window.location.reload();
}

// Email/Password Login
window.loginWithEmail = async (event) => {
    event.preventDefault();
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    
    if (email && password.length >= 6) {
        DEMO_USER.email = email;
        performDemoLogin('Email');
    } else {
        showError('Please enter valid credentials (password min 6 chars)');
    }
};

// Initialize Dashboard
document.addEventListener('DOMContentLoaded', async () => {
    const loginForm = document.getElementById('login-form');
    const loginView = document.getElementById('login-view');
    const dashboardView = document.getElementById('dashboard-view');
    const authNav = document.getElementById('auth-nav');

    // Show demo banner
    showDemoBanner();

    // Check for existing session
    const savedToken = localStorage.getItem('nh_demo_token');
    const savedUser = localStorage.getItem('nh_demo_user');
    
    if (savedToken && savedUser) {
        currentUser = JSON.parse(savedUser);
        await loadDashboard();
    }

    // Login Form Handler
    if (loginForm) {
        loginForm.addEventListener('submit', loginWithEmail);
    }

    async function loadDashboard() {
        loginView.style.display = 'none';
        dashboardView.style.display = 'block';
        
        authNav.innerHTML = `
            <span style="margin-right: 1rem; color: var(--text-secondary);">${currentUser.email}</span>
            <button class="btn-secondary btn-sm" onclick="logout()">Sign Out</button>
        `;

        await loadWorkspaceData();
        await loadIntegrations();
        await loadAPIKeys();
    }
});

// Show Demo Banner
function showDemoBanner() {
    const banner = document.createElement('div');
    banner.style.cssText = `
        position: fixed;
        bottom: 0;
        left: 0;
        right: 0;
        background: linear-gradient(135deg, #7C3AED, #A855F7);
        color: white;
        padding: 0.75rem;
        text-align: center;
        z-index: 9999;
        font-size: 0.875rem;
        box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
    `;
    banner.innerHTML = `
        <strong>Demo Mode:</strong> All data stored locally in your browser. 
        <a href="https://github.com/SiddhantSShende/NeighbourHood" target="_blank" 
           style="color: white; text-decoration: underline; margin-left: 0.5rem;">
           Run backend locally for full functionality
        </a>
        <button onclick="this.parentElement.remove()" 
                style="background: rgba(255,255,255,0.2); border: none; color: white; 
                       padding: 0.25rem 0.75rem; border-radius: 0.25rem; margin-left: 1rem; cursor: pointer;">
            Dismiss
        </button>
    `;
    document.body.appendChild(banner);
}

// Load Workspace Data
async function loadWorkspaceData() {
    let workspace = JSON.parse(localStorage.getItem('nh_demo_workspace') || 'null');
    
    if (!workspace) {
        workspace = {
            id: 'workspace-' + Date.now(),
            name: 'My Workspace',
            created_at: new Date().toISOString()
        };
        localStorage.setItem('nh_demo_workspace', JSON.stringify(workspace));
    }
    
    currentWorkspace = workspace;
    
    const selector = document.getElementById('workspace-selector');
    if (selector) {
        selector.innerHTML = `<option value="${workspace.id}">${workspace.name}</option>`;
    }
}

// Load Integrations
async function loadIntegrations() {
    const integrations = JSON.parse(localStorage.getItem('nh_demo_integrations') || '[]');
    
    const container = document.getElementById('integrations-grid');
    if (!container) return;
    
    container.innerHTML = INTEGRATIONS_CATALOG.map(integration => {
        const connected = integrations.find(i => i.type === integration.type);
        
        return `
            <div class="integration-card ${connected ? 'connected' : ''}" data-type="${integration.type}">
                <div class="integration-header">
                    <h4>${integration.name}</h4>
                    <span class="badge">${integration.category}</span>
                </div>
                <p class="text-secondary text-sm">${integration.description}</p>
                <div class="integration-footer">
                    ${connected ? `
                        <button class="btn-sm btn-secondary" onclick="disconnectIntegration('${integration.type}')">
                            Disconnect
                        </button>
                        <span class="text-success">Connected</span>
                    ` : `
                        <button class="btn-sm" onclick="connectIntegration('${integration.type}')">
                            Connect
                        </button>
                    `}
                </div>
            </div>
        `;
    }).join('');
}

// Connect Integration
window.connectIntegration = (type) => {
    const integration = INTEGRATIONS_CATALOG.find(i => i.type === type);
    if (!integration) return;
    
    const integrations = JSON.parse(localStorage.getItem('nh_demo_integrations') || '[]');
    
    if (!integrations.find(i => i.type === type)) {
        integrations.push({
            id: 'integration-' + Date.now(),
            type: type,
            name: integration.name,
            connected_at: new Date().toISOString(),
            status: 'active'
        });
        
        localStorage.setItem('nh_demo_integrations', JSON.stringify(integrations));
        showSuccess(`${integration.name} connected successfully!`);
        loadIntegrations();
        loadAPIKeys();
    }
};

// Disconnect Integration
window.disconnectIntegration = (type) => {
    if (!confirm('Disconnect this integration?')) return;
    
    const integrations = JSON.parse(localStorage.getItem('nh_demo_integrations') || '[]');
    const filtered = integrations.filter(i => i.type !== type);
    
    localStorage.setItem('nh_demo_integrations', JSON.stringify(filtered));
    showSuccess('Integration disconnected');
    loadIntegrations();
};

// Load API Keys
async function loadAPIKeys() {
    const keys = JSON.parse(localStorage.getItem('nh_demo_api_keys') || '[]');
    
    const container = document.getElementById('api-keys-list');
    if (!container) return;
    
    if (keys.length === 0) {
        container.innerHTML = `
            <div style="text-align: center; padding: 3rem; color: var(--text-secondary);">
                <p>No API keys yet. Create one to get started!</p>
            </div>
        `;
        return;
    }
    
    container.innerHTML = keys.map(key => `
        <div class="api-key-item" style="padding: 1rem; border: 1px solid var(--border-color); border-radius: 0.5rem; margin-bottom: 1rem;">
            <div class="flex justify-between items-center">
                <div>
                    <h4 style="margin-bottom: 0.25rem;">${key.name}</h4>
                    <code style="font-size: 0.875rem; color: var(--text-secondary);">${key.key}</code>
                    <div style="font-size: 0.75rem; color: var(--text-tertiary); margin-top: 0.5rem;">
                        Created: ${new Date(key.created_at).toLocaleDateString()}
                    </div>
                </div>
                <div>
                    <button class="btn-sm btn-secondary" onclick="copyToClipboard('${key.key}')">
                        Copy
                    </button>
                    <button class="btn-sm" style="background: var(--error); color: white; margin-left: 0.5rem;" 
                            onclick="revokeAPIKey('${key.id}')">
                        Revoke
                    </button>
                </div>
            </div>
        </div>
    `).join('');
}

// Create API Key
window.showCreateAPIKeyModal = () => {
    const modal = document.getElementById('api-key-modal');
    if (modal) modal.style.display = 'flex';
};

window.closeAPIKeyModal = () => {
    const modal = document.getElementById('api-key-modal');
    if (modal) modal.style.display = 'none';
};

window.createAPIKey = (event) => {
    event.preventDefault();
    
    const name = document.getElementById('key-name').value;
    if (!name) {
        showError('Please provide a key name');
        return;
    }
    
    const keys = JSON.parse(localStorage.getItem('nh_demo_api_keys') || '[]');
    const newKey = {
        id: 'key-' + Date.now(),
        name: name,
        key: 'nh_demo_' + generateRandomKey(),
        created_at: new Date().toISOString()
    };
    
    keys.push(newKey);
    localStorage.setItem('nh_demo_api_keys', JSON.stringify(keys));
    
    closeAPIKeyModal();
    showSuccess(`API Key created: ${newKey.key}`);
    loadAPIKeys();
};

// Revoke API Key
window.revokeAPIKey = (keyId) => {
    if (!confirm('Revoke this API key? This cannot be undone.')) return;
    
    const keys = JSON.parse(localStorage.getItem('nh_demo_api_keys') || '[]');
    const filtered = keys.filter(k => k.id !== keyId);
    
    localStorage.setItem('nh_demo_api_keys', JSON.stringify(filtered));
    showSuccess('API key revoked');
    loadAPIKeys();
};

// Logout
window.logout = () => {
    localStorage.removeItem('nh_demo_token');
    localStorage.removeItem('nh_demo_user');
    localStorage.removeItem('nh_demo_provider');
    window.location.reload();
};

// Utilities
function generateRandomKey() {
    return Array.from({length: 32}, () => 
        Math.random().toString(36)[2] || '0'
    ).join('');
}

window.copyToClipboard = (text) => {
    navigator.clipboard.writeText(text).then(() => {
        showSuccess('Copied to clipboard!');
    }).catch(() => {
        showError('Failed to copy');
    });
};

function showSuccess(message) {
    showToast(message, 'success');
}

function showError(message) {
    showToast(message, 'error');
}

function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    const colors = {
        success: '#10B981',
        error: '#EF4444',
        info: '#7C3AED'
    };
    
    toast.style.cssText = `
        position: fixed;
        top: 2rem;
        right: 2rem;
        background: ${colors[type]};
        color: white;
        padding: 1rem 1.5rem;
        border-radius: 0.5rem;
        box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
        z-index: 10000;
        animation: slideIn 0.3s ease-out;
    `;
    toast.textContent = message;
    
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
}

// Add slide-in animation
if (!document.getElementById('toast-animations')) {
    const style = document.createElement('style');
    style.id = 'toast-animations';
    style.textContent = `
        @keyframes slideIn {
            from {
                transform: translateX(100%);
                opacity: 0;
            }
            to {
                transform: translateX(0);
                opacity: 1;
            }
        }
    `;
    document.head.appendChild(style);
}

console.log('%c🚀 NeighbourHood Demo Mode', 'color: #7C3AED; font-size: 16px; font-weight: bold;');
console.log('%cAll data is stored locally in your browser.', 'color: #666; font-size: 12px;');
console.log('%cRun the backend server for full functionality:', 'color: #666; font-size: 12px;');
console.log('%chttps://github.com/SiddhantSShende/NeighbourHood', 'color: #7C3AED; font-size: 12px;');
