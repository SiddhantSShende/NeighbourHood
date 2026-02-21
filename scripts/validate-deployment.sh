#!/bin/bash
# ========================================
# NeighbourHood - Pre-Deployment Validation
# Professional DevOps Grade
# ========================================

set -e  # Exit on any error
set -u  # Exit on undefined variables

echo "Starting Pre-Deployment Validation..."

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Validation Functions
validate_html() {
    echo "Validating HTML files..."
    
    # Check if webpages folder exists
    if [ ! -d "webpages" ]; then
        echo -e "${RED}ERROR: webpages/ folder not found${NC}"
        exit 1
    fi
    
    # Validate required HTML files exist
    required_files=("index.html" "dashboard.html" "developer-guide.html")
    for file in "${required_files[@]}"; do
        if [ ! -f "webpages/$file" ]; then
            echo -e "${RED}ERROR: webpages/$file not found${NC}"
            exit 1
        fi
        echo -e "${GREEN}webpages/$file exists${NC}"
    done
}

validate_assets() {
    echo "Validating static assets..."
    
    if [ ! -d "webpages/static" ]; then
        echo -e "${RED}ERROR: webpages/static/ folder not found${NC}"
        exit 1
    fi
    
    required_assets=("landing.css" "dashboard.css" "dashboard.js")
    for asset in "${required_assets[@]}"; do
        if [ ! -f "webpages/static/$asset" ]; then
            echo -e "${RED}ERROR: webpages/static/$asset not found${NC}"
            exit 1
        fi
        echo -e "${GREEN}webpages/static/$asset exists${NC}"
    done
}

validate_structure() {
    echo "Validating project structure..."
    
    # Check .nojekyll exists
    if [ ! -f "webpages/.nojekyll" ]; then
        echo -e "${YELLOW}WARNING: Creating .nojekyll file${NC}"
        touch webpages/.nojekyll
    fi
    echo -e "${GREEN}.nojekyll file present${NC}"
    
    # Check for any broken internal links
    echo "Checking for broken references..."
    
    # Simple validation - check CSS/JS references exist
    if grep -q './static/landing.css' webpages/index.html; then
        echo -e "${GREEN}CSS references correct${NC}"
    else
        echo -e "${RED}ERROR: CSS reference broken in index.html${NC}"
        exit 1
    fi
}

validate_no_secrets() {
    echo "Checking for exposed secrets..."
    
    # Check for common secret patterns
    if grep -r -i "password\s*=\s*['\"]" webpages/ 2>/dev/null; then
        echo -e "${RED}ERROR: Potential password found in code${NC}"
        exit 1
    fi
    
    if grep -r -i "api[_-]key\s*:\s*['\"][^'\"]+" webpages/ 2>/dev/null | grep -v "NH_API_KEY" | grep -v "process.env"; then
        echo -e "${RED}ERROR: Potential API key found in code${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}No exposed secrets detected${NC}"
}

calculate_size() {
    echo "Calculating deployment size..."
    
    total_size=$(du -sh webpages/ | cut -f1)
    echo -e "${GREEN}Total deployment size: $total_size${NC}"
    
    # Warn if too large
    size_bytes=$(du -sb webpages/ | cut -f1)
    if [ "$size_bytes" -gt 104857600 ]; then  # 100MB
        echo -e "${YELLOW}WARNING: Deployment size exceeds 100MB${NC}"
    fi
}

# Run all validations
echo "========================================"
validate_html
echo "========================================"
validate_assets
echo "========================================"
validate_structure
echo "========================================"
validate_no_secrets
echo "========================================"
calculate_size
echo "========================================"

echo -e "${GREEN}All pre-deployment validations passed!${NC}"
echo "Ready for deployment"
exit 0
