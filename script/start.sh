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
if [ ! -f "../Backend/.env" ]; then
    echo -e "${RED}✗ Configuration not found${NC}"
    echo "  Please run ./install.sh first"
    exit 1
fi

# Load configuration
source <(grep -v '^#' ../Backend/.env | sed 's/^/export /')

# Load environment paths
export PATH=$PATH:/usr/local/go/bin
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

# ============================================
# Prerequisites Check
# ============================================
echo -e "${BLUE}Checking system requirements...${NC}"

# Check Backend binary
if [ ! -f "../Backend/app" ]; then
    echo -e "${RED}✗ Backend binary not found${NC}"
    echo "  Please run ./build.sh first"
    exit 1
fi
echo -e "${GREEN}✓ Backend binary${NC}"

# Check Frontend (if exists)
if [ -d "Frontend" ] && [ -f "Frontend/package.json" ]; then
    if [ ! -d "Frontend/node_modules" ]; then
        echo -e "${RED}✗ Frontend dependencies not installed${NC}"
        echo "  Please run ./build.sh first"
        exit 1
    fi
    echo -e "${GREEN}✓ Frontend ready${NC}"
fi

# Check Database Connection
echo -n "Checking database connection... "
if mysql -u "$DB_USER" -p"$DB_PASS" -D "$DB_NAME" -e "SELECT 1;" 2>/dev/null; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}Failed${NC}"
    echo -e "${RED}✗ Cannot connect to database${NC}"
    exit 1
fi

# ============================================
# Service Startup
# ============================================
cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down services...${NC}"
    
    # Kill backend
    if [ ! -z "$BACKEND_PID" ] && kill -0 $BACKEND_PID 2>/dev/null; then
        kill $BACKEND_PID
        wait $BACKEND_PID 2>/dev/null
        echo -e "${GREEN}✓ Backend stopped${NC}"
    fi
    
    # Kill frontend
    if [ ! -z "$FRONTEND_PID" ] && kill -0 $FRONTEND_PID 2>/dev/null; then
        kill $FRONTEND_PID
        wait $FRONTEND_PID 2>/dev/null
        echo -e "${GREEN}✓ Frontend stopped${NC}"
    fi
    
    echo ""
    exit 0
}

trap cleanup SIGINT SIGTERM EXIT

# ============================================
# Start Backend
# ============================================
echo -e "\n${BLUE}Starting Backend API...${NC}"

cd ../Backend
./app > backend.log 2>&1 &
BACKEND_PID=$!
cd ..

# Wait for backend to initialize
sleep 3

# Verify backend is running
if ! ps -p $BACKEND_PID > /dev/null; then
    echo -e "${RED}✗ Backend failed to start${NC}"
    echo -e "${YELLOW}Check Backend/backend.log for details${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Backend API started (PID: $BACKEND_PID)${NC}"

# ============================================
# Start Frontend (if exists)
# ============================================
FRONTEND_PID=""
FRONTEND_PORT=""

if [ -d "Frontend" ] && [ -f "Frontend/package.json" ]; then
    echo -e "\n${BLUE}Starting Frontend...${NC}"
    
    # Function to check if port is available
    is_port_available() {
        ! lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1
    }
    
    # Find available port
    for offset in $(seq 1 10); do
        port=$((APP_PORT + offset))
        if is_port_available $port; then
            FRONTEND_PORT=$port
            break
        fi
    done
    
    # Try common ports if offset didn't work
    if [ -z "$FRONTEND_PORT" ]; then
        for port in 3000 3001 3002 5173 5174 5175; do
            if is_port_available $port; then
                FRONTEND_PORT=$port
                break
            fi
        done
    fi
    
    if [ -z "$FRONTEND_PORT" ]; then
        echo -e "${YELLOW}⚠ No available port found for frontend${NC}"
    else
        cd Frontend
        
        # Start development server
        if grep -q '"dev"' package.json; then
            npm run dev -- --host 0.0.0.0 --port $FRONTEND_PORT > frontend.log 2>&1 &
            FRONTEND_PID=$!
        elif grep -q '"start"' package.json; then
            PORT=$FRONTEND_PORT npm start > frontend.log 2>&1 &
            FRONTEND_PID=$!
        fi
        
        cd ..
        sleep 2
        
        if [ ! -z "$FRONTEND_PID" ] && ps -p $FRONTEND_PID > /dev/null; then
            echo -e "${GREEN}✓ Frontend started (PID: $FRONTEND_PID)${NC}"
        else
            echo -e "${YELLOW}⚠ Frontend failed to start${NC}"
            FRONTEND_PID=""
        fi
    fi
fi

# ============================================
# Display Access Information
# ============================================

# Get network IP
get_network_ip() {
    local all_ips=$(ip -4 addr show | grep "inet " | grep -v "127.0.0.1" | \
                    awk '{print $2}' | cut -d/ -f1)
    
    local good_ips=$(echo "$all_ips" | grep -v "^10\.0\.2\." | grep -v "^172\.17\." | grep -v "^169\.254\.")
    
    if [ ! -z "$good_ips" ]; then
        local default_iface=$(ip route | grep "^default" | awk '{print $5}' | head -n1)
        if [ ! -z "$default_iface" ]; then
            local default_ip=$(ip addr show "$default_iface" | grep "inet " | grep -v "127.0.0.1" | \
                              awk '{print $2}' | cut -d/ -f1 | head -n1)
            if [ ! -z "$default_ip" ] && [[ ! $default_ip =~ ^10\.0\.2\. ]]; then
                echo "$default_ip"
                return 0
            fi
        fi
        
        for ip in $good_ips; do
            if [[ $ip =~ ^192\.168\. ]] || [[ $ip =~ ^10\.[0-9]+\.[0-9]+\. ]]; then
                echo "$ip"
                return 0
            fi
        done
        
        echo "$good_ips" | head -n1
        return 0
    fi
    
    if [ ! -z "$all_ips" ]; then
        echo "$all_ips" | head -n1
        return 1
    fi
    
    echo "127.0.0.1"
    return 2
}

SERVER_IP=$(get_network_ip)
IP_STATUS=$?

ALL_IPS=$(ip -4 addr show | grep "inet " | grep -v "127.0.0.1" | \
          awk '{print $2}' | cut -d/ -f1)
IP_COUNT=$(echo "$ALL_IPS" | wc -l)

echo ""
echo -e "${GREEN}┌──────────────────────────────────────────────┐${NC}"
echo -e "${GREEN}│        System Started Successfully           │${NC}"
echo -e "${GREEN}└──────────────────────────────────────────────┘${NC}"
echo ""

# Display access URLs
if [ $IP_STATUS -eq 1 ]; then
    echo -e "${BLUE}Access URL:${NC}"
    echo -e "  ${GREEN}➜${NC} API:  http://$SERVER_IP:$APP_PORT"
    
    if [ ! -z "$FRONTEND_PORT" ] && [ ! -z "$FRONTEND_PID" ]; then
        echo -e "  ${GREEN}➜${NC} Web:  http://$SERVER_IP:$FRONTEND_PORT"
    fi
    
    echo ""
    echo -e "${YELLOW}⚠ Network Notice:${NC}"
    echo "  Using NAT IP - may not be accessible from other devices"
    echo "  Fix: VirtualBox → Network → Bridged Adapter"
    
elif [ $IP_COUNT -gt 1 ]; then
    echo -e "${BLUE}Access URLs (multiple networks):${NC}"
    while IFS= read -r ip; do
        if [ "$ip" = "$SERVER_IP" ]; then
            echo -e "  ${GREEN}➜${NC} http://$ip:$APP_PORT (primary)"
            if [ ! -z "$FRONTEND_PORT" ] && [ ! -z "$FRONTEND_PID" ]; then
                echo -e "  ${GREEN}➜${NC} http://$ip:$FRONTEND_PORT"
            fi
        else
            echo -e "    http://$ip:$APP_PORT"
            if [ ! -z "$FRONTEND_PORT" ] && [ ! -z "$FRONTEND_PID" ]; then
                echo -e "    http://$ip:$FRONTEND_PORT"
            fi
        fi
    done <<< "$ALL_IPS"
else
    echo -e "${BLUE}Access URL:${NC}"
    echo -e "  ${GREEN}➜${NC} API:  http://$SERVER_IP:$APP_PORT"
    
    if [ ! -z "$FRONTEND_PORT" ] && [ ! -z "$FRONTEND_PID" ]; then
        echo -e "  ${GREEN}➜${NC} Web:  http://$SERVER_IP:$FRONTEND_PORT"
    fi
fi

echo ""
echo -e "${BLUE}Service Information:${NC}"
echo "  Backend PID:  $BACKEND_PID"
if [ ! -z "$FRONTEND_PID" ]; then
    echo "  Frontend PID: $FRONTEND_PID"
fi
echo "  Logs: Backend/backend.log"
if [ ! -z "$FRONTEND_PID" ]; then
    echo "        Frontend/frontend.log"
fi

echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
echo ""

# ============================================
# Monitor Services
# ============================================
while true; do
    sleep 10
    
    # Check if backend is still running
    if ! ps -p $BACKEND_PID > /dev/null; then
        echo -e "\n${RED}✗ Backend stopped unexpectedly${NC}"
        echo -e "${YELLOW}Check Backend/backend.log for errors${NC}"
        break
    fi
    
    # Check frontend if it was started
    if [ ! -z "$FRONTEND_PID" ] && ! ps -p $FRONTEND_PID > /dev/null; then
        echo -e "\n${YELLOW}⚠ Frontend stopped${NC}"
        FRONTEND_PID=""
    fi
done

cleanup