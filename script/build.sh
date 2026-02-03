#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}"
echo "┌──────────────────────────────────────────────┐"
echo "│        Building Application                  │"
echo "└──────────────────────────────────────────────┘"
echo -e "${NC}"

# Load environment for Go/Node paths
export PATH=$PATH:/usr/local/go/bin
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

# ============================================
# Prerequisites Check
# ============================================
echo -e "${BLUE}Checking build environment...${NC}"

# Check if configuration exists
if [ ! -f "../Backend/.env" ]; then
    echo -e "${RED}✗ Configuration not found${NC}"
    echo "  Please run ./install.sh first"
    exit 1
fi

# Check Go
if ! command -v go &>/dev/null; then
    echo -e "${RED}✗ Go not found${NC}"
    echo "  Please run ./install.sh first"
    exit 1
fi

GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
echo -e "${GREEN}✓ Go $GO_VERSION${NC}"

# Check Node.js if frontend exists
if [ -d "Frontend" ] && [ -f "Frontend/package.json" ]; then
    if ! command -v node &>/dev/null; then
        echo -e "${RED}✗ Node.js not found${NC}"
        echo "  Please run ./install.sh first"
        exit 1
    fi
    
    NODE_VERSION=$(node -v)
    echo -e "${GREEN}✓ Node.js $NODE_VERSION${NC}"
    
    if ! command -v npm &>/dev/null; then
        echo -e "${RED}✗ npm not found${NC}"
        exit 1
    fi
    
    NPM_VERSION=$(npm -v)
    echo -e "${GREEN}✓ npm $NPM_VERSION${NC}"
fi

# ============================================
# Backend Build
# ============================================
echo -e "\n${BLUE}Building Backend...${NC}"

cd ../Backend

# Backup current binary if exists
if [ -f "app" ]; then
    echo "Backing up current binary..."
    cp app app.backup
fi

# Fix go.mod version format if needed
if grep -q 'go [0-9]\+\.[0-9]\+\.[0-9]\+' go.mod 2>/dev/null; then
    echo -e "${YELLOW}Fixing go.mod version format...${NC}"
    sed -i 's/go \([0-9]\+\.[0-9]\+\)\.[0-9]\+/go \1/' go.mod
    echo -e "${GREEN}✓ go.mod fixed${NC}"
fi

# Download dependencies
echo -n "Downloading Go dependencies... "
if go mod download 2>&1 > /tmp/go_download.log; then
    echo -e "${GREEN}Done${NC}"
else
    echo -e "${RED}Failed${NC}"
    cat /tmp/go_download.log
    [ -f "app.backup" ] && mv app.backup app
    cd ..
    exit 1
fi

# Tidy modules
go mod tidy > /dev/null 2>&1

# Build application
echo -n "Compiling Backend... "
if go build -o app ./cmd/main.go 2>&1 > /tmp/go_build.log; then
    echo -e "${GREEN}Done${NC}"
    rm -f app.backup 2>/dev/null
    echo -e "${GREEN}✓ Backend built successfully${NC}"
else
    echo -e "${RED}Failed${NC}"
    echo -e "${YELLOW}Build errors:${NC}"
    cat /tmp/go_build.log
    [ -f "app.backup" ] && mv app.backup app
    cd ..
    exit 1
fi

cd ..

# ============================================
# Frontend Build (if exists)
# ============================================
if [ -d "Frontend" ] && [ -f "Frontend/package.json" ]; then
    echo -e "\n${BLUE}Building Frontend...${NC}"
    
    cd Frontend
    
    # Install/Update dependencies
    echo -n "Installing npm packages... "
    if npm install --silent 2>&1 > /tmp/npm_install.log; then
        echo -e "${GREEN}Done${NC}"
    else
        echo -e "${RED}Failed${NC}"
        cat /tmp/npm_install.log
        cd ..
        exit 1
    fi
    
    # Check if there's a build script
    if grep -q '"build"' package.json; then
        echo -n "Building production assets... "
        if npm run build --silent 2>&1 > /tmp/npm_build.log; then
            echo -e "${GREEN}Done${NC}"
            echo -e "${GREEN}✓ Frontend built successfully${NC}"
        else
            echo -e "${YELLOW}Warning: Build script failed${NC}"
            cat /tmp/npm_build.log
        fi
    else
        echo -e "${GREEN}✓ Frontend dependencies ready${NC}"
    fi
    
    cd ..
fi

# ============================================
# Build Summary
# ============================================
echo ""
echo -e "${GREEN}┌──────────────────────────────────────────────┐${NC}"
echo -e "${GREEN}│        Build Completed Successfully          │${NC}"
echo -e "${GREEN}└──────────────────────────────────────────────┘${NC}"
echo ""

# Show what was built
echo -e "${BLUE}Built Components:${NC}"
if [ -f "../Backend/app" ]; then
    BACKEND_SIZE=$(du -h ../Backend/app | cut -f1)
    echo -e "  ${GREEN}✓${NC} Backend API     (${BACKEND_SIZE})"
fi

if [ -d "Frontend/node_modules" ]; then
    echo -e "  ${GREEN}✓${NC} Frontend Assets"
fi

if [ -d "Frontend/dist" ]; then
    FRONTEND_SIZE=$(du -sh Frontend/dist | cut -f1)
    echo -e "  ${GREEN}✓${NC} Production Build (${FRONTEND_SIZE})"
fi

echo ""
echo -e "${YELLOW}Next step:${NC} Run ${GREEN}./start.sh${NC} to start the application"
echo "" 