# NeighbourHood Frontend - Demo Mode

This is the client-side demo version of the NeighbourHood dashboard, designed to work on **GitHub Pages** (static hosting).

## Features

### Client-Side Demo (No Backend Required)
- ✅ All data stored in browser localStorage
- ✅ Mock OAuth authentication
- ✅ Full integration management simulation
- ✅ API key generation and management
- ✅ Workspace management
- ✅ Complete UI/UX experience

## Pages

### [index.html](index.html)
Landing page showcasing the platform features and integrations.

### [dashboard.html](dashboard.html)
Interactive developer dashboard with:
- Google/GitHub/Email authentication simulation
- Integration connection management
- API key generation and revocation
- Workspace selector
- Real-time data persistence

### [developer-guide.html](developer-guide.html)
Comprehensive technical documentation including:
- Quick start guide
- Authentication flows
- API endpoints
- Workflow orchestration
- System architecture theory
- Security model
- Performance optimization
- Distributed systems concepts

### [tests.html](tests.html)
Automated test suite for validating all functionality:
- Demo mode tests
- Authentication tests
- Integration management tests
- API key tests
- LocalStorage tests

## How It Works

### Demo Authentication
When you click "Continue with Google" or "Continue with GitHub":
1. Shows a demo notice overlay
2. Simulates OAuth flow (1.5 second delay)
3. Creates local session token
4. Redirects to dashboard

### Data Persistence
All data is stored in browser localStorage with keys prefixed `nh_demo_`:
- `nh_demo_token` - Session token
- `nh_demo_user` - User profile
- `nh_demo_workspace` - Current workspace
- `nh_demo_integrations` - Connected integrations
- `nh_demo_api_keys` - Generated API keys

### No Backend Calls
The demo version (`dashboard-demo.js`) **does not make any HTTP requests** to backend APIs. Everything runs entirely client-side.

## Testing

### Run Automated Tests
1. Open [tests.html](tests.html) in your browser
2. Tests run automatically on page load
3. Click "Run All Tests" to re-run

### Manual Testing
1. Open [dashboard.html](dashboard.html)
2. Click "Continue with Google" or "Continue with GitHub"
3. Test features:
   - Create API keys
   - Connect integrations (Slack, Gmail, GitHub, etc.)
   - View integration catalog
   - Test localStorage persistence (refresh page)
   - Test logout/login flow

## Full Backend Version

For production use with real OAuth and database persistence, run the backend server locally:

```bash
# Clone the repository
git clone https://github.com/SiddhantSShende/NeighbourHood.git
cd NeighbourHood

# Run with Docker
docker-compose up -d

# Or run natively with Go
go run cmdapi/main.go
```

Then access the full version at `http://localhost:8080`

## Architecture

### Frontend Stack
- Pure HTML/CSS/JavaScript (no frameworks)
- CSS Grid and Flexbox for layouts
- LocalStorage API for data persistence
- Web APIs (Clipboard, Toast notifications)

### Code Organization
```
webpages/
├── index.html              # Landing page
├── dashboard.html          # Demo dashboard
├── developer-guide.html    # Technical docs
├── tests.html             # Test suite
├── .nojekyll              # Disable Jekyll processing
└── static/
    ├── landing.css        # Landing styles
    ├── dashboard.css      # Dashboard styles
    ├── dashboard-demo.js  # Demo logic (client-side only)
    ├── dashboard.js       # Production logic (requires backend)
    ├── dashboard.js       # Dashboard interactions
    ├── app.js            # Shared utilities
    └── styles.css        # Global styles
```

### Key Design Decisions

**Why Demo Mode?**
GitHub Pages only serves static files. To provide a working demo without a backend, we:
1. Mock all API calls with localStorage
2. Simulate OAuth flow delays for realism
3. Generate demo API keys client-side
4. Persist all data in browser storage

**Why localStorage?**
- Works offline
- Fast (no network latency)
- Simple API
- ~5-10MB quota (sufficient for demo)
- Browser-native (no dependencies)

**Limitations**
- Data not shared across devices
- Data lost if cache cleared
- No server-side validation
- No real OAuth tokens
- No actual API calls to third-party services

## Browser Compatibility

- ✅ Chrome 90+
- ✅ Firefox 88+
- ✅ Safari 14+
- ✅ Edge 90+
- ⚠️ IE 11 (not supported)

## Security Notes

- Demo tokens are **not** real OAuth tokens
- localStorage data is **not encrypted** (demo only)
- Generated API keys are **not** valid for production
- For production, use the backend server with real authentication

## Performance

- Initial Load: < 1 second
- Authentication Flow: ~1.5 seconds (simulated)
- Integration Connect: Instant
- API Key Generation: Instant
- Total Bundle Size: ~120 KB

## License

MIT License - See [LICENSE](../LICENSE) for details

## Contributing

Contributions welcome! See [CONTRIBUTING.md](../CONTRIBUTING.md)

## Support

- GitHub Issues: https://github.com/SiddhantSShende/NeighbourHood/issues
- Documentation: [developer-guide.html](developer-guide.html)
- API Reference: [../API_DOCUMENTATION.md](../API_DOCUMENTATION.md)

---

**Live Demo**: https://siddhantssshende.github.io/NeighbourHood/
