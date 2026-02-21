#!/bin/bash
# ========================================
# NeighbourHood - Post-Deployment Verification
# Professional DevOps Grade
# ========================================

set -e

SITE_URL="${1:-https://siddhantssshende.github.io/NeighbourHood/}"
MAX_RETRIES=5
RETRY_DELAY=10

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "Post-Deployment Verification for: $SITE_URL"

verify_url() {
    local url=$1
    local name=$2
    local retry=0
    
    echo "Checking: $name ($url)"
    
    while [ $retry -lt $MAX_RETRIES ]; do
        http_code=$(curl -s -o /dev/null -w "%{http_code}" -L "$url" --max-time 15 || echo "000")
        
        if [ "$http_code" = "200" ]; then
            echo -e "${GREEN}✓ $name is accessible (HTTP $http_code)${NC}"
            return 0
        else
            retry=$((retry + 1))
            if [ $retry -lt $MAX_RETRIES ]; then
                echo -e "${YELLOW}⚠ Retry $retry/$MAX_RETRIES for $name (HTTP $http_code)${NC}"
                sleep $RETRY_DELAY
            fi
        fi
    done
    
    echo -e "${RED}❌ $name failed after $MAX_RETRIES retries (HTTP $http_code)${NC}"
    return 1
}

check_content() {
    local url=$1
    local expected_text=$2
    local name=$3
    
    echo "Verifying content: $name"
    
    content=$(curl -s -L "$url" --max-time 15 || echo "")
    
    if echo "$content" | grep -q "$expected_text"; then
        echo -e "${GREEN}✓ Content verification passed for $name${NC}"
        return 0
    else
        echo -e "${RED}❌ Expected content not found in $name${NC}"
        return 1
    fi
}

# Main verification
echo "========================================"
echo "Phase 1: URL Accessibility"
echo "========================================"

verify_url "$SITE_URL" "Landing Page" || exit 1
verify_url "${SITE_URL}dashboard.html" "Dashboard" || exit 1
verify_url "${SITE_URL}developer-guide.html" "Developer Guide" || exit 1

echo ""
echo "========================================"
echo "Phase 2: Content Verification"
echo "========================================"

check_content "$SITE_URL" "NeighbourHood" "Landing Page Title" || exit 1
check_content "$SITE_URL" "Integration Platform" "Landing Page Content" || exit 1
check_content "${SITE_URL}developer-guide.html" "Developer Guide" "Developer Guide Content" || exit 1

echo ""
echo "========================================"
echo "Phase 3: Asset Verification"
echo "========================================"

verify_url "${SITE_URL}static/landing.css" "Landing CSS" || exit 1
verify_url "${SITE_URL}static/dashboard.css" "Dashboard CSS" || exit 1
verify_url "${SITE_URL}static/dashboard.js" "Dashboard JS" || exit 1

echo ""
echo "========================================"
echo -e "${GREEN}All post-deployment verifications passed!${NC}"
echo "Deployment successful and verified!"
echo "Site: $SITE_URL"
echo "========================================"

exit 0
