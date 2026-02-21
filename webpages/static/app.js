// Basic frontend logic

// SSO login functions
window.loginWithGoogle = () => {
    window.location.href = '/auth/google/login';
};

window.loginWithGitHub = () => {
    window.location.href = '/auth/github/login';
};

// Check for token in URL (from OAuth callback)
const urlParams = new URLSearchParams(window.location.search);
const token = urlParams.get('token');
const provider = urlParams.get('provider');

if (token) {
    localStorage.setItem('nh_token', token);
    // Clean URL
    window.history.replaceState({}, document.title, '/');
}

document.addEventListener('DOMContentLoaded', () => {
    const loginForm = document.getElementById('login-form');
    const loginView = document.getElementById('login-view');
    const dashboardView = document.getElementById('dashboard-view');
    const authNav = document.getElementById('auth-nav');

    // Check for existing token (mock)
    const token = localStorage.getItem('nh_token');
    if (token) {
        showDashboard();
    }

    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;

        try {
            const res = await fetch('/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password })
            });

            if (res.ok) {
                const data = await res.json();
                localStorage.setItem('nh_token', data.token);
                showDashboard();
            } else {
                alert('Login failed');
            }
        } catch (err) {
            console.error(err);
            alert('Error logging in');
        }
    });

    function showDashboard() {
        loginView.style.display = 'none';
        dashboardView.style.display = 'block';
        authNav.innerHTML = '<button onclick="logout()" style="background: transparent; border: 1px solid var(--border-color);">Sign Out</button>';
    }

    window.logout = () => {
        localStorage.removeItem('nh_token');
        location.reload();
    };

    window.connectProvider = (provider) => {
        // Mock connection flow
        alert(`Redirecting to ${provider} OAuth...`);
        // In real app: window.location.href = `/api/integrations/${provider}/auth`;

        // Mock success update
        if (provider === 'slack') {
            const badge = document.getElementById('slack-status');
            badge.className = 'status-badge status-connected';
            badge.innerText = 'Connected';
        }
    };
});

// Fetch and display integrations
async function fetchIntegrations() {
    // In production, fetch from backend or config
    const integrations = [
        { type: "slack" },
        { type: "gmail" },
        { type: "jira" }
    ];
    const list = document.getElementById("integrations-list");
    list.innerHTML = integrations.map(i => `<div>${i.type} <button onclick="getAuthUrl('${i.type}')">Get Auth URL</button></div>`).join("");
}

async function getAuthUrl(provider) {
    const res = await fetch("/api/integration/authurl", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ provider, state: "dev" })
    });
    const data = await res.json();
    alert("Auth URL: " + data.url);
}

async function executeWorkflow() {
    const wfText = document.getElementById("workflow-json").value;
    let wf;
    try { wf = JSON.parse(wfText); } catch (e) { alert("Invalid JSON"); return; }
    // For demo, tokens are empty. In production, collect from user/session.
    const res = await fetch("/api/workflow/execute", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ workflow: wf, tokens: {} })
    });
    const data = await res.json();
    document.getElementById("workflow-result").textContent = JSON.stringify(data, null, 2);
}

// Auto-load integrations on page load
window.onload = fetchIntegrations;
