/**
 * NeighbourHood — Production Frontend Client
 *
 * This file is the backend-connected counterpart to dashboard-demo.js.
 * It is NOT used while DEMO_MODE is active (see dashboard-demo.js).
 * Load it on pages where the NeighbourHood backend is running.
 *
 * Storage key: 'nh_token'  (backend-issued JWT)
 */

'use strict';

// ── OAuth helpers ─────────────────────────────────────────────────────────────

window.loginWithGoogle = () => {
    window.location.href = '/auth/google/login';
};

window.loginWithGitHub = () => {
    window.location.href = '/auth/github/login';
};

// ── Handle OAuth callback: token passed back as ?token=<jwt> ─────────────────
(function extractOAuthToken() {
    const params = new URLSearchParams(window.location.search);
    const jwt = params.get('token');
    if (!jwt) return;
    localStorage.setItem('nh_token', jwt);
    // Remove token from URL to avoid accidental sharing / history leakage
    const clean = window.location.pathname + (window.location.hash || '');
    window.history.replaceState({}, document.title, clean);
}());

// ── Fetch integrations from backend ──────────────────────────────────────────

/**
 * Fetch and display integrations from /api/integrations.
 * Requires a valid JWT stored under 'nh_token'.
 */
async function fetchIntegrations() {
    const container = document.getElementById('integrations-grid');
    if (!container) return; // element absent on this page

    const token = localStorage.getItem('nh_token');
    const headers = token ? { Authorization: 'Bearer ' + token } : {};

    try {
        const res = await fetch('/api/integrations', { headers });
        if (!res.ok) throw new Error('HTTP ' + res.status);
        const data = await res.json();
        const list = Array.isArray(data.integrations) ? data.integrations : [];
        renderIntegrationList(container, list);
    } catch (err) {
        container.innerHTML = '<p class="text-secondary">Could not load integrations. Is the backend running?</p>';
        console.error('[app.js] fetchIntegrations:', err);
    }
}

function renderIntegrationList(container, integrations) {
    if (integrations.length === 0) {
        container.innerHTML = '<p class="text-secondary">No integrations registered.</p>';
        return;
    }
    container.innerHTML = integrations.map(i => `
        <div class="integration-card">
            <h4>${escHtml(i.name || i.type)}</h4>
            <span class="badge">${escHtml(i.category || '')}</span>
            <p class="text-secondary text-sm">${escHtml(i.description || '')}</p>
            <button class="btn-sm" onclick="getAuthUrl('${escHtml(i.type)}')">Connect</button>
        </div>
    `).join('');
}

// ── Get OAuth URL for a provider ──────────────────────────────────────────────

async function getAuthUrl(provider) {
    const token = localStorage.getItem('nh_token');
    try {
        const res = await fetch('/api/integration/authurl', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                ...(token ? { Authorization: 'Bearer ' + token } : {})
            },
            body: JSON.stringify({ provider, state: generateNonce() })
        });
        if (!res.ok) throw new Error('HTTP ' + res.status);
        const data = await res.json();
        if (data.url) window.location.href = data.url;
    } catch (err) {
        showError('Could not get auth URL: ' + err.message);
    }
}

// ── DOM-ready initialisation ──────────────────────────────────────────────────

document.addEventListener('DOMContentLoaded', () => {
    const loginForm     = document.getElementById('login-form');
    const loginView     = document.getElementById('login-view');
    const dashboardView = document.getElementById('dashboard-view');
    const authNav       = document.getElementById('auth-nav');

    const savedToken = localStorage.getItem('nh_token');
    if (savedToken) showDashboard();

    if (loginForm) {
        loginForm.addEventListener('submit', handleEmailLogin);
    }

    async function handleEmailLogin(e) {
        e.preventDefault();
        const email    = document.getElementById('email').value.trim();
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
                const err = await res.json().catch(() => ({}));
                showError(err.error || 'Login failed. Check your credentials.');
            }
        } catch (err) {
            console.error('[app.js] login:', err);
            showError('Could not reach the server. Is the backend running?');
        }
    }

    function showDashboard() {
        if (loginView)     loginView.style.display     = 'none';
        if (dashboardView) dashboardView.style.display = 'block';
        if (authNav) {
            authNav.innerHTML = '<button class="btn-secondary btn-sm" onclick="logout()">Sign Out</button>';
        }
        fetchIntegrations();
    }
});

// ── Session management ────────────────────────────────────────────────────────

window.logout = () => {
    localStorage.removeItem('nh_token');
    window.location.reload();
};

// ── Utilities ─────────────────────────────────────────────────────────────────

/** Escape HTML special characters to prevent XSS when inserting into innerHTML. */
function escHtml(str) {
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

/** Generate a cryptographically random nonce for OAuth state parameter. */
function generateNonce(len = 16) {
    const arr = new Uint8Array(len);
    crypto.getRandomValues(arr);
    return Array.from(arr, b => b.toString(16).padStart(2, '0')).join('');
}

function showError(msg) {
    const el = document.getElementById('login-error');
    if (el) {
        el.textContent = msg;
        el.style.display = 'block';
        return;
    }
    // Fallback: console only (avoid alert in production)
    console.error('[app.js]', msg);
}
