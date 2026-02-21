# Deployment Guide

This guide covers deploying NeighbourHood to production environments.

## 📋 Prerequisites

- Linux server (Ubuntu 20.04+ recommended)
- Docker & Docker Compose (for containerized deployment)
- PostgreSQL 15+ (if not using Docker)
- Domain name with SSL certificate
- OAuth credentials for integrations (Slack, Gmail, Jira)

---

## 🐳 Docker Deployment (Recommended)

### 1. Initial Setup

```bash
# Clone the repository
git clone <your-repo-url>
cd NeighbourHood

# Copy and configure environment
cp .env.example .env
nano .env  # Edit with your production values
```

### 2. Configure Environment Variables

Edit `.env` with production values:

```env
# Server
PORT=8080
ENV=production

# Database
DB_PASSWORD=your-secure-password-here

# Slack
SLACK_CLIENT_ID=your-slack-client-id
SLACK_CLIENT_SECRET=your-slack-client-secret
SLACK_REDIRECT_URL=https://yourdomain.com/callback/slack

# Gmail
GMAIL_CLIENT_ID=your-gmail-client-id
GMAIL_CLIENT_SECRET=your-gmail-client-secret
GMAIL_REDIRECT_URL=https://yourdomain.com/callback/gmail

# Jira
JIRA_CLIENT_ID=your-jira-client-id
JIRA_CLIENT_SECRET=your-jira-client-secret
JIRA_REDIRECT_URL=https://yourdomain.com/callback/jira

# JWT
JWT_SECRET=generate-a-strong-random-secret-here
```

### 3. Build and Start

```bash
# Build and start containers
docker-compose up -d

# Check logs
docker-compose logs -f

# Check health
curl http://localhost:8080/health
```

### 4. SSL/TLS Setup with Nginx

Create `/etc/nginx/sites-available/neighbourhood`:

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable and restart:

```bash
sudo ln -s /etc/nginx/sites-available/neighbourhood /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

---

## 🖥️ Bare Metal Deployment

### 1. Install Go

```bash
wget https://go.dev/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### 2. Install PostgreSQL

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database
sudo -u postgres psql
CREATE DATABASE neighbourhood;
CREATE USER neighbourhood WITH ENCRYPTED PASSWORD 'your-password';
GRANT ALL PRIVILEGES ON DATABASE neighbourhood TO neighbourhood;
\q
```

### 3. Build Application

```bash
cd NeighbourHood
go mod download
cd cmdapi
go build -o /usr/local/bin/neighbourhood main.go
```

### 4. Create Systemd Service

Create `/etc/systemd/system/neighbourhood.service`:

```ini
[Unit]
Description=NeighbourHood Integration Platform
After=network.target postgresql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/neighbourhood
Environment="PORT=8080"
Environment="ENV=production"
EnvironmentFile=/opt/neighbourhood/.env
ExecStart=/usr/local/bin/neighbourhood
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Start service:

```bash
sudo systemctl daemon-reload
sudo systemctl start neighbourhood
sudo systemctl enable neighbourhood
sudo systemctl status neighbourhood
```

---

## ☁️ Cloud Platform Deployments

### AWS Elastic Beanstalk

1. Install EB CLI:
```bash
pip install awsebcli
```

2. Initialize:
```bash
eb init -p docker neighbourhood-app
```

3. Create environment:
```bash
eb create neighbourhood-prod
```

4. Deploy:
```bash
eb deploy
```

### Google Cloud Run

1. Build container:
```bash
gcloud builds submit --tag gcr.io/PROJECT_ID/neighbourhood
```

2. Deploy:
```bash
gcloud run deploy neighbourhood \
  --image gcr.io/PROJECT_ID/neighbourhood \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

### Heroku

1. Create app:
```bash
heroku create neighbourhood-app
```

2. Add PostgreSQL:
```bash
heroku addons:create heroku-postgresql:hobby-dev
```

3. Deploy:
```bash
git push heroku main
```

### DigitalOcean App Platform

1. Create `app.yaml`:
```yaml
name: neighbourhood
services:
- name: web
  github:
    branch: main
    deploy_on_push: true
    repo: your-username/neighbourhood
  build_command: go build -o bin/neighbourhood ./cmdapi/main.go
  run_command: ./bin/neighbourhood
  envs:
  - key: ENV
    value: production
databases:
- name: db
  engine: PG
  version: "15"
```

2. Deploy via UI or CLI:
```bash
doctl apps create --spec app.yaml
```

---

## 🔧 Production Configuration

### Database Optimization

In production PostgreSQL:

```sql
-- Increase connection pool
ALTER SYSTEM SET max_connections = 200;

-- Enable query logging
ALTER SYSTEM SET log_statement = 'all';
ALTER SYSTEM SET log_duration = on;

-- Performance tuning
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;

-- Reload config
SELECT pg_reload_conf();
```

### Environment Variables Checklist

- [ ] `ENV` set to `production`
- [ ] Strong `DB_PASSWORD`
- [ ] Valid OAuth credentials for all providers
- [ ] Strong `JWT_SECRET` (min 32 characters)
- [ ] Proper `REDIRECT_URL` values (HTTPS)
- [ ] Database connection pooling configured
- [ ] Log level set appropriately

### Security Hardening

1. **Firewall Rules**:
```bash
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 80/tcp   # HTTP
sudo ufw allow 443/tcp  # HTTPS
sudo ufw enable
```

2. **Rate Limiting** (Nginx):
```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

location /api/ {
    limit_req zone=api burst=20 nodelay;
    proxy_pass http://localhost:8080;
}
```

3. **Database Connection Pooling**:
```go
// In database/db.go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

---

## 📊 Monitoring & Logging

### Application Logs

```bash
# Docker logs
docker-compose logs -f app

# Systemd logs
journalctl -u neighbourhood -f

# Export logs to file
docker-compose logs app > app.log
```

### Health Monitoring

Set up monitoring with:
- Prometheus for metrics
- Grafana for visualization
- Uptime Robot for availability monitoring

Example Prometheus config:

```yaml
scrape_configs:
  - job_name: 'neighbourhood'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Database Backups

Automated daily backups:

```bash
# Create backup script
cat > /usr/local/bin/backup-neighbourhood.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -U neighbourhood neighbourhood > /backups/neighbourhood_$DATE.sql
find /backups -name "neighbourhood_*.sql" -mtime +7 -delete
EOF

chmod +x /usr/local/bin/backup-neighbourhood.sh

# Add to crontab
crontab -e
0 2 * * * /usr/local/bin/backup-neighbourhood.sh
```

---

## 🚀 Deployment Checklist

- [ ] Environment variables configured
- [ ] Database migrations run
- [ ] SSL/TLS certificates installed
- [ ] Firewall configured
- [ ] Backups scheduled
- [ ] Monitoring set up
- [ ] Logs configured
- [ ] Performance testing completed
- [ ] Security audit performed
- [ ] Documentation updated
- [ ] Team trained on operations

---

## 🆘 Troubleshooting

### Application won't start

```bash
# Check logs
docker-compose logs app

# Verify environment
docker-compose exec app env

# Test database connection
docker-compose exec app nc -zv db 5432
```

### Database connection issues

```bash
# Check PostgreSQL is running
systemctl status postgresql

# Test connection
psql -h localhost -U neighbourhood -d neighbourhood

# Check logs
tail -f /var/log/postgresql/postgresql-15-main.log
```

### Performance issues

```bash
# Check resource usage
docker stats

# Profile application
go tool pprof http://localhost:8080/debug/pprof/profile

# Check database slow queries
SELECT query, calls, total_time, mean_time 
FROM pg_stat_statements 
ORDER BY total_time DESC 
LIMIT 10;
```

---

## 📞 Support

For deployment assistance:
- GitHub Issues: [issues](../../issues)
- Email: ops@neighbourhood.dev
- Documentation: [docs](../../wiki)

---

**Happy Deploying! 🎉**
