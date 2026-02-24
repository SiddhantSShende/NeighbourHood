/**
 * NeighbourHood — Theme Manager
 * Handles light / dark mode toggle with localStorage persistence.
 * Loaded in <head> (render-blocking) to prevent flash of wrong theme.
 */

// ── 1. Apply saved theme immediately (before first paint) ──────────────────
(function () {
    var saved = localStorage.getItem('nh-theme');
    if (saved === 'dark') {
        document.documentElement.setAttribute('data-theme', 'dark');
    }
    // Default is light — no attribute needed
}());

// ── 2. Wire up the toggle button after DOM is ready ────────────────────────
function _nhInitTheme() {
    var btn = document.getElementById('theme-toggle');
    if (!btn) return;

    function isDark() {
        return document.documentElement.getAttribute('data-theme') === 'dark';
    }

    function setTheme(dark) {
        if (dark) {
            document.documentElement.setAttribute('data-theme', 'dark');
            localStorage.setItem('nh-theme', 'dark');
        } else {
            document.documentElement.removeAttribute('data-theme');
            localStorage.setItem('nh-theme', 'light');
        }
        updateAriaLabel();
    }

    function updateAriaLabel() {
        var label = isDark() ? 'Switch to light mode' : 'Switch to dark mode';
        btn.setAttribute('aria-label', label);
        btn.setAttribute('title', label);
    }

    btn.addEventListener('click', function () {
        setTheme(!isDark());
    });

    // Keep aria-label in sync with current state on page load too
    updateAriaLabel();
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', _nhInitTheme);
} else {
    _nhInitTheme();
}
