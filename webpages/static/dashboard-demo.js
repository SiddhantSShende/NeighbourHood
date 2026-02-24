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

// Integration Catalog - 100+ Platforms
const INTEGRATIONS_CATALOG = [
    // --- Communication ---
    { type: "slack", name: "Slack", category: "Communication", description: "Send messages, manage channels, and team notifications", requiredScopes: ["chat:write", "channels:read", "users:read"] },
    { type: "ms-teams", name: "Microsoft Teams", category: "Communication", description: "Collaborate with meetings, chats, and file sharing", requiredScopes: ["ChannelMessage.Send", "Chat.ReadWrite"] },
    { type: "discord", name: "Discord", category: "Communication", description: "Send messages and manage Discord servers and channels", requiredScopes: ["bot", "messages.read"] },
    { type: "telegram", name: "Telegram", category: "Communication", description: "Send messages and notifications via Telegram bots", requiredScopes: ["bot:message"] },
    { type: "whatsapp", name: "WhatsApp Business", category: "Communication", description: "Send transactional messages via WhatsApp Business API", requiredScopes: ["messages:send"] },
    { type: "twilio", name: "Twilio", category: "Communication", description: "Send SMS, voice calls, and WhatsApp messages globally", requiredScopes: ["sms:send", "voice:call"] },
    { type: "zoom", name: "Zoom", category: "Communication", description: "Create meetings, manage webinars, and recordings", requiredScopes: ["meeting:write", "recording:read"] },
    { type: "google-meet", name: "Google Meet", category: "Communication", description: "Schedule and manage Google Meet video conferences", requiredScopes: ["calendar.events", "meet.readonly"] },
    { type: "webex", name: "Webex", category: "Communication", description: "Cisco Webex meetings, messaging, and teamwork", requiredScopes: ["spark:messages_write", "spark:rooms_read"] },
    { type: "rocketchat", name: "Rocket.Chat", category: "Communication", description: "Open-source team communication and collaboration", requiredScopes: ["chat:write", "channels:read"] },
    { type: "mattermost", name: "Mattermost", category: "Communication", description: "Self-hosted team messaging and collaboration platform", requiredScopes: ["posts:write", "channels:read"] },
    { type: "ringcentral", name: "RingCentral", category: "Communication", description: "Cloud phone, video meetings, and messaging platform", requiredScopes: ["SMS", "Meetings"] },

    // --- Email ---
    { type: "gmail", name: "Gmail", category: "Email", description: "Send emails, read inbox, and manage labels", requiredScopes: ["gmail.send", "gmail.readonly", "gmail.labels"] },
    { type: "outlook", name: "Outlook / Microsoft 365", category: "Email", description: "Send and manage emails via Microsoft 365 mail", requiredScopes: ["Mail.ReadWrite", "Mail.Send"] },
    { type: "mailchimp", name: "Mailchimp", category: "Email", description: "Email marketing campaigns, automation, and analytics", requiredScopes: ["campaigns:read", "campaigns:write"] },
    { type: "sendgrid", name: "SendGrid", category: "Email", description: "Deliver transactional and marketing emails at scale", requiredScopes: ["mail.send"] },
    { type: "mailgun", name: "Mailgun", category: "Email", description: "Powerful email sending, tracking, and routing API", requiredScopes: ["messages:send", "events:read"] },
    { type: "amazon-ses", name: "Amazon SES", category: "Email", description: "High-volume transactional and bulk email service", requiredScopes: ["ses:SendEmail"] },
    { type: "postmark", name: "Postmark", category: "Email", description: "Fast transactional email delivery with open tracking", requiredScopes: ["email:send"] },
    { type: "klaviyo", name: "Klaviyo", category: "Email", description: "Email and SMS marketing for e-commerce automation", requiredScopes: ["campaigns:read", "profiles:read"] },
    { type: "convertkit", name: "ConvertKit", category: "Email", description: "Email marketing for creators, subscribers, and sequences", requiredScopes: ["subscribers:read", "broadcasts:write"] },

    // --- Developer Tools ---
    { type: "github", name: "GitHub", category: "Developer Tools", description: "Manage repositories, issues, and pull requests", requiredScopes: ["repo", "read:org", "workflow"] },
    { type: "gitlab", name: "GitLab", category: "Developer Tools", description: "Git repositories, CI/CD pipelines, and DevOps", requiredScopes: ["read_user", "api", "write_repository"] },
    { type: "bitbucket", name: "Bitbucket", category: "Developer Tools", description: "Git code hosting with pull requests and pipelines", requiredScopes: ["repository:write", "pullrequest:write"] },
    { type: "azure-devops", name: "Azure DevOps", category: "Developer Tools", description: "CI/CD pipelines, boards, and Azure repos", requiredScopes: ["vso.build", "vso.code"] },
    { type: "jenkins", name: "Jenkins", category: "Developer Tools", description: "Open-source automation server for CI/CD pipelines", requiredScopes: ["job:read", "job:build"] },
    { type: "circleci", name: "CircleCI", category: "Developer Tools", description: "Automate builds, test, and deployment pipelines", requiredScopes: ["read:project", "write:pipeline"] },
    { type: "travis-ci", name: "Travis CI", category: "Developer Tools", description: "Hosted continuous integration service for GitHub projects", requiredScopes: ["build:read", "build:create"] },
    { type: "vercel", name: "Vercel", category: "Developer Tools", description: "Deploy frontends and serverless functions instantly", requiredScopes: ["deployments:read", "deployments:write"] },
    { type: "netlify", name: "Netlify", category: "Developer Tools", description: "Build, deploy, and host modern web projects", requiredScopes: ["sites:read", "deploys:write"] },
    { type: "sonarqube", name: "SonarQube", category: "Developer Tools", description: "Static code analysis and code quality management", requiredScopes: ["scan:execute", "analysis:create"] },
    { type: "docker-hub", name: "Docker Hub", category: "Developer Tools", description: "Container image registry and distribution service", requiredScopes: ["repository:pull", "repository:push"] },
    { type: "npm", name: "npm Registry", category: "Developer Tools", description: "Publish and manage JavaScript packages", requiredScopes: ["packages:read", "packages:write"] },

    // --- Project Management ---
    { type: "jira", name: "Jira", category: "Project Management", description: "Create issues, manage projects, and track workflows", requiredScopes: ["read:jira-work", "write:jira-work"] },
    { type: "linear", name: "Linear", category: "Project Management", description: "Modern issue tracking and project management tool", requiredScopes: ["issues:read", "issues:write"] },
    { type: "asana", name: "Asana", category: "Project Management", description: "Manage tasks, projects, timelines, and goals", requiredScopes: ["tasks:read", "tasks:write", "projects:read"] },
    { type: "trello", name: "Trello", category: "Project Management", description: "Visual boards, lists, and cards for project management", requiredScopes: ["read", "write", "account"] },
    { type: "clickup", name: "ClickUp", category: "Project Management", description: "All-in-one project management and productivity platform", requiredScopes: ["tasks:write", "goals:read"] },
    { type: "monday", name: "Monday.com", category: "Project Management", description: "Work OS for managing teams, projects, and workflows", requiredScopes: ["boards:read", "items:write"] },
    { type: "basecamp", name: "Basecamp", category: "Project Management", description: "Project management and team communication hub", requiredScopes: ["projects:read", "todos:write"] },
    { type: "shortcut", name: "Shortcut", category: "Project Management", description: "Collaborative project management for software teams", requiredScopes: ["stories:read", "stories:write"] },
    { type: "height", name: "Height", category: "Project Management", description: "Autonomous project management with AI assistance", requiredScopes: ["tasks:read", "tasks:write"] },
    { type: "wrike", name: "Wrike", category: "Project Management", description: "Work management platform for cross-functional teams", requiredScopes: ["tasks:read", "folders:write"] },

    // --- Productivity ---
    { type: "notion", name: "Notion", category: "Productivity", description: "Manage pages, databases, and collaborative content", requiredScopes: ["read:content", "write:content"] },
    { type: "airtable", name: "Airtable", category: "Productivity", description: "Spreadsheet-database hybrid for workflows and apps", requiredScopes: ["data.records:read", "data.records:write"] },
    { type: "coda", name: "Coda", category: "Productivity", description: "Interactive documents and automated workflows", requiredScopes: ["docs:read", "rows:write"] },
    { type: "confluence", name: "Confluence", category: "Productivity", description: "Team wiki and knowledge management platform", requiredScopes: ["read:confluence-content.all", "write:confluence-content"] },
    { type: "google-docs", name: "Google Docs", category: "Productivity", description: "Create, edit, and collaborate on documents", requiredScopes: ["documents:read", "documents:write"] },
    { type: "google-sheets", name: "Google Sheets", category: "Productivity", description: "Read and write data in spreadsheets programmatically", requiredScopes: ["spreadsheets:read", "spreadsheets:write"] },
    { type: "ms-word", name: "Microsoft Word", category: "Productivity", description: "Create and edit Word documents via Microsoft 365", requiredScopes: ["Files.ReadWrite", "Documents.Write"] },
    { type: "ms-excel", name: "Microsoft Excel", category: "Productivity", description: "Read and write Excel spreadsheets via Microsoft 365", requiredScopes: ["Files.ReadWrite"] },
    { type: "todoist", name: "Todoist", category: "Productivity", description: "Task management and personal productivity app", requiredScopes: ["tasks:read", "tasks:write"] },
    { type: "obsidian", name: "Obsidian", category: "Productivity", description: "Personal knowledge management and note linking", requiredScopes: ["vault:read", "vault:write"] },

    // --- CRM ---
    { type: "salesforce", name: "Salesforce", category: "CRM", description: "Manage leads, contacts, accounts, and opportunities", requiredScopes: ["api", "full"] },
    { type: "hubspot", name: "HubSpot", category: "CRM", description: "CRM, marketing, sales, and customer service platform", requiredScopes: ["contacts", "crm.objects.deals.write"] },
    { type: "pipedrive", name: "Pipedrive", category: "CRM", description: "Sales-focused CRM and pipeline management", requiredScopes: ["deals:full", "contacts:full"] },
    { type: "zoho-crm", name: "Zoho CRM", category: "CRM", description: "Sales automation, analytics, and omnichannel CRM", requiredScopes: ["ZohoCRM.modules.ALL"] },
    { type: "freshsales", name: "Freshsales", category: "CRM", description: "AI-powered CRM with built-in phone and email", requiredScopes: ["contacts:read", "deals:write"] },
    { type: "close", name: "Close CRM", category: "CRM", description: "Sales CRM designed for inside sales teams", requiredScopes: ["leads:read", "calls:write"] },
    { type: "copper", name: "Copper", category: "CRM", description: "Google Workspace-native CRM for small businesses", requiredScopes: ["people:read", "opportunities:write"] },
    { type: "attio", name: "Attio", category: "CRM", description: "Data-driven CRM for modern, fast-growing companies", requiredScopes: ["records:read", "notes:write"] },

    // --- Customer Support ---
    { type: "zendesk", name: "Zendesk", category: "Customer Support", description: "Help desk, ticketing, and customer service platform", requiredScopes: ["tickets:read", "tickets:write", "users:read"] },
    { type: "freshdesk", name: "Freshdesk", category: "Customer Support", description: "Cloud-based helpdesk and customer support software", requiredScopes: ["tickets:full", "contacts:read"] },
    { type: "intercom", name: "Intercom", category: "Customer Support", description: "Conversational support and customer messaging platform", requiredScopes: ["conversations:read", "messages:write"] },
    { type: "help-scout", name: "Help Scout", category: "Customer Support", description: "Simple, human customer support and help desk", requiredScopes: ["conversations:read", "mailboxes:read"] },
    { type: "crisp", name: "Crisp", category: "Customer Support", description: "Live chat, chatbot, and knowledge base platform", requiredScopes: ["conversations:write", "profiles:read"] },
    { type: "drift", name: "Drift", category: "Customer Support", description: "Conversational AI platform for sales and support", requiredScopes: ["conversations:read", "contacts:write"] },
    { type: "liveagent", name: "LiveAgent", category: "Customer Support", description: "Multi-channel helpdesk with live chat and ticketing", requiredScopes: ["tickets:full", "chats:read"] },

    // --- Cloud Storage ---
    { type: "google-drive", name: "Google Drive", category: "Cloud Storage", description: "Store, share, and access files from anywhere", requiredScopes: ["drive.file", "drive.readonly"] },
    { type: "dropbox", name: "Dropbox", category: "Cloud Storage", description: "Cloud file storage, sync, and team collaboration", requiredScopes: ["files.content.read", "files.content.write"] },
    { type: "box", name: "Box", category: "Cloud Storage", description: "Secure cloud content management and file sharing", requiredScopes: ["root_readwrite", "manage_groups"] },
    { type: "onedrive", name: "OneDrive", category: "Cloud Storage", description: "Microsoft cloud storage with Office integration", requiredScopes: ["Files.ReadWrite", "Sites.Read.All"] },
    { type: "amazon-s3", name: "Amazon S3", category: "Cloud Storage", description: "Scalable object storage in AWS", requiredScopes: ["s3:GetObject", "s3:PutObject"] },
    { type: "backblaze", name: "Backblaze B2", category: "Cloud Storage", description: "Low-cost S3-compatible cloud object storage", requiredScopes: ["readFiles", "writeFiles"] },
    { type: "cloudflare-r2", name: "Cloudflare R2", category: "Cloud Storage", description: "Zero-egress S3-compatible object storage", requiredScopes: ["object:read", "object:write"] },

    // --- Analytics & Monitoring ---
    { type: "google-analytics", name: "Google Analytics", category: "Analytics", description: "Track website traffic, events, and user behavior", requiredScopes: ["analytics.readonly"] },
    { type: "mixpanel", name: "Mixpanel", category: "Analytics", description: "Product analytics for events, funnels, and retention", requiredScopes: ["data:read", "events:write"] },
    { type: "amplitude", name: "Amplitude", category: "Analytics", description: "Behavioral analytics and product intelligence platform", requiredScopes: ["events:ingest", "cohorts:read"] },
    { type: "segment", name: "Segment", category: "Analytics", description: "Customer data platform for event collection and routing", requiredScopes: ["track:write", "identify:write"] },
    { type: "heap", name: "Heap", category: "Analytics", description: "Automatically capture every user interaction", requiredScopes: ["events:read", "sessions:read"] },
    { type: "hotjar", name: "Hotjar", category: "Analytics", description: "Heatmaps, recordings, and user behavior analytics", requiredScopes: ["recordings:read", "surveys:read"] },
    { type: "plausible", name: "Plausible", category: "Analytics", description: "Privacy-friendly, lightweight website analytics", requiredScopes: ["stats:read", "sites:read"] },
    { type: "datadog", name: "Datadog", category: "Analytics", description: "Cloud monitoring, APM, and infrastructure observability", requiredScopes: ["metrics:read", "logs:write"] },
    { type: "sentry", name: "Sentry", category: "Analytics", description: "Application error tracking and performance monitoring", requiredScopes: ["event:read", "project:read"] },
    { type: "newrelic", name: "New Relic", category: "Analytics", description: "Full-stack observability and application performance", requiredScopes: ["insights:read", "apm:read"] },

    // --- Payment & Finance ---
    { type: "stripe", name: "Stripe", category: "Finance", description: "Accept online payments, manage subscriptions and invoices", requiredScopes: ["charges:read", "customers:write"] },
    { type: "paypal", name: "PayPal", category: "Finance", description: "Accept PayPal payments and manage transactions", requiredScopes: ["payments:read", "invoicing:write"] },
    { type: "square", name: "Square", category: "Finance", description: "Point-of-sale, payments, and business management", requiredScopes: ["PAYMENTS_READ", "INVOICES_WRITE"] },
    { type: "razorpay", name: "Razorpay", category: "Finance", description: "Indian payment gateway for cards, UPI, and wallets", requiredScopes: ["payments:read", "orders:write"] },
    { type: "quickbooks", name: "QuickBooks", category: "Finance", description: "Accounting, invoicing, payroll, and expense tracking", requiredScopes: ["accounting", "payment"] },
    { type: "xero", name: "Xero", category: "Finance", description: "Online accounting for small businesses and accountants", requiredScopes: ["accounting.transactions", "accounting.contacts"] },
    { type: "freshbooks", name: "FreshBooks", category: "Finance", description: "Invoicing, time tracking, and accounting for freelancers", requiredScopes: ["invoices:read", "expenses:write"] },
    { type: "braintree", name: "Braintree", category: "Finance", description: "Full-stack payment processing from PayPal", requiredScopes: ["transactions:read", "subscriptions:write"] },

    // --- E-Commerce ---
    { type: "shopify", name: "Shopify", category: "E-Commerce", description: "Manage products, orders, customers, and Shopify store", requiredScopes: ["read_orders", "write_products"] },
    { type: "woocommerce", name: "WooCommerce", category: "E-Commerce", description: "WordPress e-commerce orders, products, and customers", requiredScopes: ["products:read", "orders:write"] },
    { type: "bigcommerce", name: "BigCommerce", category: "E-Commerce", description: "Scalable e-commerce platform for growing brands", requiredScopes: ["store_v2_products", "store_v2_orders"] },
    { type: "magento", name: "Magento / Adobe Commerce", category: "E-Commerce", description: "Enterprise-grade e-commerce platform integration", requiredScopes: ["catalog:read", "orders:write"] },
    { type: "amazon-seller", name: "Amazon Seller", category: "E-Commerce", description: "Manage Amazon marketplace listings and orders", requiredScopes: ["Orders", "Inventory", "Reports"] },
    { type: "ebay", name: "eBay", category: "E-Commerce", description: "Manage eBay listings, orders, and seller accounts", requiredScopes: ["sell.fulfillment", "sell.inventory"] },
    { type: "etsy", name: "Etsy", category: "E-Commerce", description: "Manage Etsy shop listings and orders", requiredScopes: ["listings_r", "transactions_r"] },

    // --- Social Media ---
    { type: "twitter", name: "Twitter / X", category: "Social Media", description: "Post tweets, manage replies, and track mentions", requiredScopes: ["tweet.read", "tweet.write", "users.read"] },
    { type: "linkedin", name: "LinkedIn", category: "Social Media", description: "Post updates, manage company pages, and analytics", requiredScopes: ["w_member_social", "r_organization_social"] },
    { type: "facebook", name: "Facebook", category: "Social Media", description: "Manage Facebook pages, posts, and ad campaigns", requiredScopes: ["pages_manage_posts", "ads_management"] },
    { type: "instagram", name: "Instagram", category: "Social Media", description: "Schedule posts, manage media, and view insights", requiredScopes: ["instagram_basic", "instagram_content_publish"] },
    { type: "youtube", name: "YouTube", category: "Social Media", description: "Upload videos, manage channel, and view analytics", requiredScopes: ["youtube.upload", "youtube.readonly"] },
    { type: "pinterest", name: "Pinterest", category: "Social Media", description: "Create pins, manage boards, and run ads", requiredScopes: ["boards:read", "pins:write"] },
    { type: "tiktok", name: "TikTok", category: "Social Media", description: "Manage TikTok business content and analytics", requiredScopes: ["video.upload", "user.info.basic"] },
    { type: "reddit", name: "Reddit", category: "Social Media", description: "Submit posts, manage subreddits, and monitor mentions", requiredScopes: ["submit", "read"] },

    // --- AI & Machine Learning ---
    { type: "openai", name: "OpenAI", category: "AI & ML", description: "GPT-4, DALL-E, Whisper, and Embeddings API", requiredScopes: ["models:read", "completions:write"] },
    { type: "anthropic", name: "Anthropic Claude", category: "AI & ML", description: "Claude AI for text analysis, summaries, and generation", requiredScopes: ["messages:write"] },
    { type: "google-ai", name: "Google Gemini", category: "AI & ML", description: "Gemini models for text, vision, and reasoning tasks", requiredScopes: ["generativelanguage.generateContent"] },
    { type: "cohere", name: "Cohere", category: "AI & ML", description: "Text generation, embeddings, and semantic search", requiredScopes: ["generate", "embed"] },
    { type: "huggingface", name: "Hugging Face", category: "AI & ML", description: "Access thousands of open-source ML models via API", requiredScopes: ["inference:write", "repos:read"] },
    { type: "replicate", name: "Replicate", category: "AI & ML", description: "Run open-source ML models via cloud API", requiredScopes: ["predictions:write", "models:read"] },
    { type: "pinecone", name: "Pinecone", category: "AI & ML", description: "Vector database for semantic search and AI memory", requiredScopes: ["vectors:upsert", "queries:run"] },
    { type: "openrouter", name: "OpenRouter", category: "AI & ML", description: "Unified API gateway for 200+ AI models", requiredScopes: ["chat:completions"] },

    // --- Database & Data ---
    { type: "supabase", name: "Supabase", category: "Database", description: "Open-source Firebase alternative with Postgres", requiredScopes: ["database.read", "database.write"] },
    { type: "firebase", name: "Firebase", category: "Database", description: "Google Firebase Realtime Database and Firestore", requiredScopes: ["firestore.read", "firestore.write"] },
    { type: "mongodb-atlas", name: "MongoDB Atlas", category: "Database", description: "Cloud-native MongoDB data platform and APIs", requiredScopes: ["clusters:read", "data:write"] },
    { type: "planetscale", name: "PlanetScale", category: "Database", description: "Serverless MySQL platform with branching", requiredScopes: ["read_credentials", "write_data"] },
    { type: "neon", name: "Neon", category: "Database", description: "Serverless Postgres with branching and autoscaling", requiredScopes: ["projects:read", "branches:write"] },
    { type: "redis-cloud", name: "Redis Cloud", category: "Database", description: "Managed Redis for caching, sessions, and pub/sub", requiredScopes: ["db:read", "db:write"] },

    // --- HR & People ---
    { type: "workday", name: "Workday", category: "HR", description: "Enterprise HR, payroll, and talent management", requiredScopes: ["workers:read", "staffing:write"] },
    { type: "bamboohr", name: "BambooHR", category: "HR", description: "HR software for employee data, PTO, and onboarding", requiredScopes: ["employees:read", "timeoff:write"] },
    { type: "rippling", name: "Rippling", category: "HR", description: "HR, IT, and finance management in one platform", requiredScopes: ["employees:read", "payroll:read"] },
    { type: "gusto", name: "Gusto", category: "HR", description: "Payroll, benefits, and HR for small businesses", requiredScopes: ["employees:read", "payrolls:run"] },
    { type: "greenhouse", name: "Greenhouse", category: "HR", description: "Applicant tracking system and hiring platform", requiredScopes: ["candidates:read", "jobs:write"] },
    { type: "workable", name: "Workable", category: "HR", description: "Recruiting software and applicant tracking", requiredScopes: ["jobs:read", "candidates:write"] },

    // --- Marketing ---
    { type: "google-ads", name: "Google Ads", category: "Marketing", description: "Manage Google Ads campaigns, keywords, and analytics", requiredScopes: ["adwords:read", "campaigns:write"] },
    { type: "facebook-ads", name: "Facebook Ads", category: "Marketing", description: "Create and manage Facebook and Instagram ad campaigns", requiredScopes: ["ads_management", "ads_read"] },
    { type: "activecampaign", name: "ActiveCampaign", category: "Marketing", description: "Email marketing, automation, and CRM platform", requiredScopes: ["contacts:write", "automations:read"] },
    { type: "marketo", name: "Marketo / Adobe Marketo", category: "Marketing", description: "B2B marketing automation and lead management", requiredScopes: ["leads:read", "campaigns:write"] },
    { type: "brevo", name: "Brevo (Sendinblue)", category: "Marketing", description: "Email marketing, SMS, and marketing automation", requiredScopes: ["campaigns:write", "contacts:read"] },

    // --- Scheduling & Calendar ---
    { type: "google-calendar", name: "Google Calendar", category: "Scheduling", description: "Create events, manage calendars, and send invitations", requiredScopes: ["calendar.events", "calendar.readonly"] },
    { type: "outlook-calendar", name: "Outlook Calendar", category: "Scheduling", description: "Manage calendar events via Microsoft 365", requiredScopes: ["Calendars.ReadWrite"] },
    { type: "calendly", name: "Calendly", category: "Scheduling", description: "Schedule meetings and automate appointment booking", requiredScopes: ["event_types:read", "scheduled_events:write"] },
    { type: "cal-com", name: "Cal.com", category: "Scheduling", description: "Open-source scheduling and meeting booking platform", requiredScopes: ["bookings:read", "availability:write"] },
    { type: "acuity", name: "Acuity Scheduling", category: "Scheduling", description: "Online appointment scheduling and client management", requiredScopes: ["appointments:read", "availability:write"] },
];

// Get unique categories from catalog
function getCatalogCategories() {
    return ['All', ...new Set(INTEGRATIONS_CATALOG.map(i => i.category))].sort((a, b) => a === 'All' ? -1 : a.localeCompare(b));
}

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
            <span style="margin-right: 1rem; color: var(--text-secondary);">${escHtml(currentUser.email)}</span>
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
        <a href="https://github.com/SiddhantSShende/NeighbourHood" target="_blank" rel="noopener noreferrer"
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
// Track current filter state
let currentCategory = 'All';
let currentSearch = '';

async function loadIntegrations() {
    const integrations = JSON.parse(localStorage.getItem('nh_demo_integrations') || '[]');

    // Render category tabs if not yet rendered
    renderCategoryTabs();

    const container = document.getElementById('integrations-grid');
    if (!container) return;

    // Apply filters
    let catalog = INTEGRATIONS_CATALOG;
    if (currentCategory !== 'All') {
        catalog = catalog.filter(i => i.category === currentCategory);
    }
    if (currentSearch.trim()) {
        const q = currentSearch.toLowerCase();
        catalog = catalog.filter(i =>
            i.name.toLowerCase().includes(q) ||
            i.category.toLowerCase().includes(q) ||
            i.description.toLowerCase().includes(q)
        );
    }

    // Update count label
    const countEl = document.getElementById('integration-count');
    if (countEl) countEl.textContent = `${catalog.length} of ${INTEGRATIONS_CATALOG.length} platforms`;

    if (catalog.length === 0) {
        container.innerHTML = `<div style="grid-column:1/-1;padding:2rem;text-align:center;color:#6B7280;">No integrations match your search.</div>`;
        return;
    }

    container.innerHTML = catalog.map(integration => {
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
                        <span class="text-success">✓ Connected</span>
                    ` : `
                        <button class="btn-sm" onclick="showIntegrationModal('${integration.type}')">
                            Connect
                        </button>
                    `}
                </div>
            </div>
        `;
    }).join('');
}

function renderCategoryTabs() {
    const tabsContainer = document.getElementById('category-tabs');
    if (!tabsContainer || tabsContainer.dataset.rendered) return;
    tabsContainer.dataset.rendered = 'true';

    const categories = getCatalogCategories();
    tabsContainer.innerHTML = categories.map(cat => `
        <button class="cat-tab ${cat === currentCategory ? 'active' : ''}"
                onclick="filterByCategory('${cat}')">${cat}</button>
    `).join('');
}

window.filterByCategory = (category) => {
    currentCategory = category;
    document.querySelectorAll('.cat-tab').forEach(btn => {
        btn.classList.toggle('active', btn.textContent === category);
    });
    loadIntegrations();
};

window.searchIntegrations = (query) => {
    currentSearch = query;
    loadIntegrations();
};

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
                    <h4 style="margin-bottom: 0.25rem;">${escHtml(key.name)}</h4>
                    <code style="font-size: 0.875rem; color: var(--text-secondary);">${key.key}</code>
                    <div style="font-size: 0.75rem; color: var(--text-tertiary); margin-top: 0.5rem;">
                        Created: ${new Date(key.created_at).toLocaleDateString()}
                    </div>
                </div>
                <div style="display:flex;gap:0.5rem;align-items:center;flex-shrink:0;">
                    <button class="btn-sm btn-secondary" onclick="copyToClipboard('${key.key}')">Copy</button>
                    <button class="btn-sm btn-danger" onclick="revokeAPIKey('${key.id}')">Revoke</button>
                </div>
            </div>
        </div>
    `).join('');
}

// Create API Key
// Rate Limit Helpers
window.setRateLimit = (value) => {
    document.getElementById('key-rate-limit').value = value;
    const slider = document.getElementById('key-rate-slider');
    if (slider) slider.value = Math.min(value, 100000);
    updateRateLimitDisplay(value);
    document.querySelectorAll('.rate-preset').forEach(btn => {
        const btnVal = parseInt(btn.getAttribute('onclick').match(/\d+/)[0]);
        btn.classList.toggle('selected', btnVal === value);
    });
};

window.syncRateLimit = (value, source) => {
    const num = parseInt(value);
    document.getElementById('key-rate-limit').value = num;
    if (source === 'slider') updateRateLimitDisplay(num);
    document.querySelectorAll('.rate-preset').forEach(btn => {
        const btnVal = parseInt(btn.getAttribute('onclick').match(/\d+/)[0]);
        btn.classList.toggle('selected', btnVal === num);
    });
};

function updateRateLimitDisplay(value) {
    const el = document.getElementById('rate-limit-display');
    if (!el) return;
    el.textContent = value >= 100000 ? 'Unlimited' : Number(value).toLocaleString() + ' req/hr';
}

window.showCreateAPIKeyModal = () => {
    const modal = document.getElementById('api-key-modal');
    if (modal) modal.style.display = 'flex';
};

window.closeAPIKeyModal = () => {
    const modal = document.getElementById('api-key-modal');
    if (modal) modal.style.display = 'none';
};

/** Close the integration connect/scope modal. */
window.closeIntegrationModal = () => {
    const modal = document.getElementById('integration-modal');
    if (modal) modal.style.display = 'none';
};

/**
 * Show the integration modal with required OAuth scopes before connecting.
 * Uses event-listener-based confirm button to avoid onclick-attribute injection.
 */
window.showIntegrationModal = (type) => {
    const integration = INTEGRATIONS_CATALOG.find(i => i.type === type);
    if (!integration) return;

    const modal   = document.getElementById('integration-modal');
    const titleEl = document.getElementById('modal-title');
    const body    = document.getElementById('modal-body');

    // Fallback: connect directly when modal markup is absent (e.g. test context)
    if (!modal || !titleEl || !body) {
        connectIntegration(type);
        return;
    }

    const integrations = JSON.parse(localStorage.getItem('nh_demo_integrations') || '[]');
    const alreadyConnected = integrations.some(i => i.type === type);

    titleEl.textContent = 'Connect to ' + integration.name;
    body.innerHTML = `
        <p style="margin-bottom:1rem;color:var(--text-secondary);">${escHtml(integration.description)}</p>
        <div style="margin-bottom:1.5rem;">
            <p style="font-weight:600;font-size:0.875rem;margin-bottom:0.5rem;">Required permissions:</p>
            <ul style="margin-left:1.25rem;line-height:1.8;font-size:0.875rem;">
                ${integration.requiredScopes.map(s => '<li><code>' + escHtml(s) + '</code></li>').join('')}
            </ul>
        </div>
        ${alreadyConnected
            ? '<p style="color:var(--text-success,#10B981);font-weight:600;">✓ Already connected</p>'
            : `<div style="display:flex;gap:0.75rem;">
                   <button class="btn-sm" style="flex:1;" id="modal-confirm-btn">✓ Authorize &amp; Connect</button>
                   <button class="btn-sm btn-secondary" onclick="closeIntegrationModal()">Cancel</button>
               </div>`
        }
    `;

    if (!alreadyConnected) {
        // Use event listener instead of inline onclick to avoid attribute injection
        document.getElementById('modal-confirm-btn').addEventListener('click', () => {
            connectIntegration(type);
            closeIntegrationModal();
        }, { once: true });
    }

    modal.style.display = 'flex';
};

// Close modals when clicking the backdrop outside the modal-content box
document.addEventListener('click', (e) => {
    const integrationModal = document.getElementById('integration-modal');
    const apiKeyModal = document.getElementById('api-key-modal');
    if (e.target === integrationModal) closeIntegrationModal();
    if (e.target === apiKeyModal) closeAPIKeyModal();
});

window.createAPIKey = (event) => {
    event.preventDefault();
    
    const name = document.getElementById('key-name').value.trim();
    if (!name) {
        showError('Please provide a key name');
        return;
    }

    const rateLimit = parseInt(document.getElementById('key-rate-limit').value, 10) || 1000;
    const keys = JSON.parse(localStorage.getItem('nh_demo_api_keys') || '[]');
    const newKey = {
        id: 'key-' + Date.now(),
        name: name,
        key: 'nh_demo_' + generateRandomKey(),
        created_at: new Date().toISOString(),
        rate_limit: rateLimit
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

/** Escape HTML special characters to prevent XSS when inserting user data into innerHTML. */
function escHtml(str) {
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
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
