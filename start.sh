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
echo "│        Starting Backup System               │"
echo "└──────────────────────────────────────────────┘"
echo -e "${NC}"

# ============================================
# Configuration Check
# ============================================
if [ ! -f "Backend/.env" ]; then
    echo -e "${RED}Configuration not found${NC}"
    echo "Please run ./install.sh first"
    exit 1
fi

# Load configuration
source <(grep -v '^#' Backend/.env | sed 's/^/export /')

# Load nvm if available
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

# Load Go path
export PATH=$PATH:/usr/local/go/bin

# ============================================
# Node.js Version Check for Vite 5
# ============================================
check_node_version() {
    if command -v node &>/dev/null; then
        local version=$(node -v | sed 's/v//' | cut -d. -f1)
        
        if [ "$version" -ge 18 ]; then
            return 0
        else
            return 1
        fi
    fi
    return 1
}

if [ -d "Frontend" ] && [ -f "Frontend/package.json" ]; then
    # Check if project uses Vite 5
    if grep -q '"vite".*"5\.' Frontend/package.json 2>/dev/null || \
       grep -q '"vite".*"\^5\.' Frontend/package.json 2>/dev/null; then
        
        if ! check_node_version; then
            NODE_VERSION=$(node -v 2>/dev/null || echo "not found")
            echo -e "${YELLOW}⚠ Node.js $NODE_VERSION detected${NC}"
            echo -e "${YELLOW}  Vite 5 requires Node.js 18 or higher${NC}"
            echo ""
            read -p "Update Node.js to latest LTS version? (y/n): " -n 1 -r
            echo ""
            
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${BLUE}Installing Node.js LTS via nvm...${NC}"
                
                # Check if nvm is installed
                if [ ! -d "$HOME/.nvm" ]; then
                    echo "Installing nvm..."
                    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash
                    
                    # Load nvm
                    export NVM_DIR="$HOME/.nvm"
                    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
                fi
                
                # Install Node.js LTS
                if nvm install --lts; then
                    nvm use --lts
                    nvm alias default lts/*
                    
                    NEW_NODE_VERSION=$(node -v)
                    echo -e "${GREEN}✓ Node.js updated to $NEW_NODE_VERSION${NC}"
                else
                    echo -e "${RED}✗ Failed to update Node.js${NC}"
                    echo -e "${YELLOW}⚠ Continuing with current version${NC}"
                fi
            else
                echo -e "${YELLOW}⚠ Continuing with current version (may cause issues)${NC}"
            fi
        fi
    fi
fi

# ============================================
# Auto-Detect Changes & Build
# ============================================
echo -e "${BLUE}Checking system status...${NC}"

# Function to check if rebuild is needed
check_rebuild() {
    # Check if binary exists
    if [ ! -f "Backend/app" ]; then
        echo -e "${YELLOW}Binary not found, need to build${NC}"
        return 0
    fi
    
    # Get binary modification time
    local binary_mtime=$(stat -c %Y "Backend/app" 2>/dev/null || echo 0)
    
    # Check Go files for changes
    local go_files_changed=0
    for go_file in $(find Backend -name "*.go" -type f 2>/dev/null); do
        local file_mtime=$(stat -c %Y "$go_file" 2>/dev/null)
        if [ $file_mtime -gt $binary_mtime ]; then
            go_files_changed=1
            break
        fi
    done
    
    if [ $go_files_changed -eq 1 ]; then
        return 0
    fi
    
    # Check go.mod/go.sum
    if [ -f "Backend/go.mod" ] && [ "Backend/go.mod" -nt "Backend/app" ]; then
        return 0
    fi
    
    if [ -f "Backend/go.sum" ] && [ "Backend/go.sum" -nt "Backend/app" ]; then
        return 0
    fi
    
    # Check .env file
    if [ "Backend/.env" -nt "Backend/app" ]; then
        return 0
    fi
    
    return 1
}

# Check if rebuild is needed
if check_rebuild; then
    echo -e "\n${BLUE}Building application...${NC}"
    
    # Backup current binary if exists
    if [ -f "Backend/app" ]; then
        cp Backend/app Backend/app.backup 2>/dev/null || true
    fi
    
    # Build the application
    cd Backend
    
    # Fix go.mod version format if needed
    if grep -q 'go [0-9]\+\.[0-9]\+\.[0-9]\+' go.mod 2>/dev/null; then
        echo -e "${YELLOW}Fixing go.mod version format...${NC}"
        # Extract major.minor version only (e.g., 1.25.1 -> 1.25)
        sed -i 's/go \([0-9]\+\.[0-9]\+\)\.[0-9]\+/go \1/' go.mod
        echo -e "${GREEN}✓ go.mod fixed${NC}"
    fi
    
    # Ensure dependencies are downloaded
    echo -n "Downloading dependencies... "
    (go mod download 2>&1 > /tmp/go_download.log) &
    pid=$!
    while kill -0 $pid 2>/dev/null; do
        echo -n "▓"
        sleep 0.3
    done
    wait $pid
    if [ $? -ne 0 ]; then
        echo ""
        echo -e "${RED}✗ Failed to download dependencies${NC}"
        cat /tmp/go_download.log
        if [ -f "app.backup" ]; then
            mv app.backup app
        fi
        cd ..
        exit 1
    fi
    echo " Done"
    
    go mod tidy > /dev/null 2>&1
    
    # Build with proper path
    echo -n "Compiling... "
    (go build -o app ./cmd/main.go 2>&1 > /tmp/go_build.log) &
    pid=$!
    while kill -0 $pid 2>/dev/null; do
        echo -n "▓"
        sleep 0.2
    done
    wait $pid
    if [ $? -ne 0 ]; then
        echo ""
        echo -e "${RED}✗ Build failed${NC}"
        echo -e "${YELLOW}Error details:${NC}"
        cat /tmp/go_build.log
        # Restore backup if exists
        if [ -f "app.backup" ]; then
            mv app.backup app
        fi
        cd ..
        exit 1
    fi
    echo " Done"
    
    rm -f app.backup 2>/dev/null
    echo -e "${GREEN}✓ Application built${NC}"
    cd ..
else
    echo -e "${GREEN}✓ Application is up to date${NC}"
fi

# ============================================
# Frontend Dependency Check
# ============================================
if [ -d "Frontend" ] && [ -f "Frontend/package.json" ]; then
    # Check if node_modules exists or package.json is newer
    if [ ! -d "Frontend/node_modules" ] || [ "Frontend/package.json" -nt "Frontend/node_modules" ]; then
        echo -e "\n${BLUE}Installing Web Interface...${NC}"
        cd Frontend
        
        # Run npm install with progress
        echo -n "Installing packages... "
        (npm install --silent 2>&1 > /tmp/npm_install.log) &
        pid=$!
        while kill -0 $pid 2>/dev/null; do
            echo -n "▓"
            sleep 0.5
        done
        wait $pid
        if [ $? -ne 0 ]; then
            echo ""
            echo -e "${RED}✗ Web Interface installation failed${NC}"
            cat /tmp/npm_install.log
            cd ..
            exit 1
        fi
        echo " Done"

        echo -e "${GREEN}✓ Web Interface ready${NC}"
        cd ..
    fi
fi

# ============================================
# Database Connection Test
# ============================================
if ! mysql -u "$DB_USER" -p"$DB_PASS" -D "$DB_NAME" -e "SELECT 1;" 2>/dev/null; then
    echo -e "${RED}✗ Database connection failed${NC}"
    exit 1
fi

# ============================================
# Service Startup
# ============================================
cleanup() {
    echo ""
    
    # Kill backend if running
    if [ ! -z "$BACKEND_PID" ] && kill -0 $BACKEND_PID 2>/dev/null; then
        kill $BACKEND_PID
        wait $BACKEND_PID 2>/dev/null
    fi
    
    # Kill frontend if running
    if [ ! -z "$FRONTEND_PID" ] && kill -0 $FRONTEND_PID 2>/dev/null; then
        kill $FRONTEND_PID
        wait $FRONTEND_PID 2>/dev/null
    fi
    
    exit 0
}

trap cleanup SIGINT SIGTERM EXIT

# Start Backend
cd Backend
./app > backend.log 2>&1 &
BACKEND_PID=$!
cd ..

# Wait for backend to start
sleep 3

# Verify backend is running
if ! ps -p $BACKEND_PID > /dev/null; then
    echo -e "${RED}✗ API Service failed to start${NC}"
    exit 1
fi

# Start Frontend if exists
FRONTEND_PID=""
FRONTEND_PORT=""

if [ -d "Frontend" ] && [ -f "Frontend/package.json" ]; then
    # Function to check if port is available
    is_port_available() {
        ! lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1
    }
    
    # Find available port for frontend
    # Try APP_PORT + 1 first, then scan range
    FRONTEND_PORT=""
    
    # Try APP_PORT + 1 to + 10
    for offset in $(seq 1 10); do
        port=$((APP_PORT + offset))
        if is_port_available $port; then
            FRONTEND_PORT=$port
            break
        fi
    done
    
    # If still not found, scan common frontend ports
    if [ -z "$FRONTEND_PORT" ]; then
        for port in 3000 3001 3002 5173 5174 5175 4000 4001; do
            if is_port_available $port; then
                FRONTEND_PORT=$port
                break
            fi
        done
    fi
    
    # If still not found, scan range 5000-5100
    if [ -z "$FRONTEND_PORT" ]; then
        for port in $(seq 5000 5100); do
            if is_port_available $port; then
                FRONTEND_PORT=$port
                break
            fi
        done
    fi
    
    # Start frontend if port found
    if [ ! -z "$FRONTEND_PORT" ]; then
        cd Frontend
        
        # Check for development script
        if grep -q '"dev"' package.json; then
            npm run dev -- --host 0.0.0.0 --port $FRONTEND_PORT > frontend.log 2>&1 &
            FRONTEND_PID=$!
        elif grep -q '"start"' package.json; then
            PORT=$FRONTEND_PORT npm start > frontend.log 2>&1 &
            FRONTEND_PID=$!
        fi
        
        cd ..
        sleep 2
    else
        echo -e "${YELLOW}⚠ Could not find available port for frontend${NC}"
    fi
fi

# ============================================
# Display System Information
# ============================================

# Get the correct network IP (same logic as install.sh)
get_network_ip() {
    # Get all IPs excluding loopback
    local all_ips=$(ip -4 addr show | grep "inet " | grep -v "127.0.0.1" | \
                    awk '{print $2}' | cut -d/ -f1)
    
    # Get IPs excluding common NAT ranges
    local good_ips=$(echo "$all_ips" | grep -v "^10\.0\.2\." | grep -v "^172\.17\." | grep -v "^169\.254\.")
    
    # Check if we have good IPs
    if [ ! -z "$good_ips" ]; then
        local ip_count=$(echo "$good_ips" | wc -l)
        
        # If only one good IP, use it
        if [ $ip_count -eq 1 ]; then
            echo "$good_ips"
            return 0
        fi
        
        # Multiple good IPs - use priority logic
        # Priority 1: IP from default route interface
        local default_iface=$(ip route | grep "^default" | awk '{print $5}' | head -n1)
        if [ ! -z "$default_iface" ]; then
            local default_ip=$(ip addr show "$default_iface" | grep "inet " | grep -v "127.0.0.1" | \
                              awk '{print $2}' | cut -d/ -f1 | head -n1)
            if [ ! -z "$default_ip" ] && [[ ! $default_ip =~ ^10\.0\.2\. ]]; then
                echo "$default_ip"
                return 0
            fi
        fi
        
        # Priority 2: Prefer 192.168.x.x or 10.x.x.x
        for ip in $good_ips; do
            if [[ $ip =~ ^192\.168\. ]] || [[ $ip =~ ^10\.[0-9]+\.[0-9]+\. ]]; then
                echo "$ip"
                return 0
            fi
        done
        
        # Use first good IP
        echo "$good_ips" | head -n1
        return 0
    fi
    
    # No good IPs found, use NAT IP as fallback
    if [ ! -z "$all_ips" ]; then
        echo "$all_ips" | head -n1
        return 1  # Return 1 to indicate NAT IP is being used
    fi
    
    # Absolute fallback
    echo "127.0.0.1"
    return 2
}

SERVER_IP=$(get_network_ip)
IP_STATUS=$?

# Show all available IPs
ALL_IPS=$(ip -4 addr show | grep "inet " | grep -v "127.0.0.1" | \
          awk '{print $2}' | cut -d/ -f1)
IP_COUNT=$(echo "$ALL_IPS" | wc -l)

echo ""
echo -e "${GREEN}✓ System started successfully${NC}"
echo ""

# Check if using NAT IP (status = 1)
if [ $IP_STATUS -eq 1 ]; then
    echo -e "${BLUE}Access URL:${NC}"
    echo "  http://$SERVER_IP:$APP_PORT"
    
    if [ ! -z "$FRONTEND_PORT" ] && [ ! -z "$FRONTEND_PID" ]; then
        echo "  Frontend: http://$SERVER_IP:$FRONTEND_PORT"
    fi
    
    echo ""
    echo -e "${YELLOW}⚠ Network Notice:${NC}"
    echo "  Using NAT IP - may not be accessible from other devices"
    echo "  Fix: VirtualBox → Settings → Network → Bridged Adapter"
    
elif [ $IP_COUNT -gt 1 ]; then
    echo -e "${BLUE}Access URL (multiple networks):${NC}"
    while IFS= read -r ip; do
        if [ "$ip" = "$SERVER_IP" ]; then
            echo -e "  ${GREEN}➜ API Service: http://$ip:$APP_PORT${NC} (primary)"
            if [ ! -z "$FRONTEND_PORT" ] && [ ! -z "$FRONTEND_PID" ]; then
                echo -e "  ${GREEN}➜ Web Interface: http://$ip:$FRONTEND_PORT${NC} (primary)"
            fi
        else
            echo -e "    http://$ip:$APP_PORT"
            if [ ! -z "$FRONTEND_PORT" ] && [ ! -z "$FRONTEND_PID" ]; then
                echo -e "    ${GREEN}➜ Web Interface: http://$ip:$FRONTEND_PORT${NC}"
            fi
        fi
    done <<< "$ALL_IPS"
else
    echo -e "${BLUE}Access URL:${NC}"
    echo "  http://$SERVER_IP:$APP_PORT"
    if [ ! -z "$FRONTEND_PORT" ] && [ ! -z "$FRONTEND_PID" ]; then
        echo "  Frontend: http://$SERVER_IP:$FRONTEND_PORT"
    fi
fi

echo ""
echo -e "${YELLOW}Press Ctrl+C to stop${NC}"
echo ""

# ============================================
# Keep running
# ============================================
while true; do
    sleep 10
    
    # Check API Service health
    if ! ps -p $BACKEND_PID > /dev/null; then
        echo -e "\n${RED}✗ API Service stopped${NC}"
        break
    fi
done

cleanup