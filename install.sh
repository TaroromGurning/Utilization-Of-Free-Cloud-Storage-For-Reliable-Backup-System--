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
echo "│        System Installation Wizard           │"
echo "└──────────────────────────────────────────────┘"
echo -e "${NC}"

# ============================================
# Dependency Check & Install
# ============================================
echo -e "${BLUE}System Dependencies Check${NC}"

# Progress bar function
show_progress() {
    local duration=$1
    local message=$2
    local steps=20
    local step_duration=$(echo "scale=2; $duration / $steps" | bc)
    
    echo -n "$message "
    for i in $(seq 1 $steps); do
        echo -n "▓"
        sleep $step_duration
    done
    echo " Done"
}

# ============================================
# Go Installation (Official, >= 1.20)
# ============================================
check_go_version() {
    if command -v go &>/dev/null; then
        local version=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+' | head -1)
        local major=$(echo $version | cut -d. -f1)
        local minor=$(echo $version | cut -d. -f2)
        
        # Check if version >= 1.20
        if [ "$major" -ge 1 ] && [ "$minor" -ge 20 ]; then
            echo -e "${GREEN}✓ Go $version (compatible)${NC}"
            return 0
        else
            echo -e "${YELLOW}⚠ Go $version detected (requires >= 1.20)${NC}"
            return 1
        fi
    fi
    return 1
}

install_go_official() {
    echo -e "${BLUE}Installing Go (official)...${NC}"
    
    # Detect architecture
    local arch=$(uname -m)
    case $arch in
        x86_64) arch="amd64" ;;
        aarch64) arch="arm64" ;;
        armv6l) arch="armv6l" ;;
        *) 
            echo -e "${RED}✗ Unsupported architecture: $arch${NC}"
            return 1
            ;;
    esac
    
    # Get latest stable version (1.23.x)
    local go_version="1.23.5"
    local go_file="go${go_version}.linux-${arch}.tar.gz"
    local go_url="https://go.dev/dl/${go_file}"
    
    echo "Downloading Go ${go_version}..."
    if ! wget -q --show-progress "$go_url" -O "/tmp/${go_file}"; then
        echo -e "${RED}✗ Failed to download Go${NC}"
        return 1
    fi
    
    echo "Installing..."
    # Remove old installation
    sudo rm -rf /usr/local/go
    
    # Extract new version
    sudo tar -C /usr/local -xzf "/tmp/${go_file}"
    rm "/tmp/${go_file}"
    
    # Add to PATH if not already there
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    
    # Add to current session
    export PATH=$PATH:/usr/local/go/bin
    
    # Verify installation
    if command -v go &>/dev/null; then
        local installed_version=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+\.[0-9]+')
        echo -e "${GREEN}✓ Go ${installed_version} installed${NC}"
        return 0
    else
        echo -e "${RED}✗ Go installation failed${NC}"
        return 1
    fi
}

# Check and install Go
if ! check_go_version; then
    if ! install_go_official; then
        echo -e "${YELLOW}Falling back to apt installation...${NC}"
        sudo apt update -qq && sudo apt install -y golang-go -qq
    fi
fi

# ============================================
# Node.js Installation (Official via nvm)
# ============================================
check_node_version() {
    if command -v node &>/dev/null; then
        local version=$(node -v | sed 's/v//' | cut -d. -f1)
        
        # Vite 5 requires Node.js >= 18
        if [ "$version" -ge 18 ]; then
            echo -e "${GREEN}✓ Node.js v$version (compatible with Vite 5)${NC}"
            return 0
        else
            echo -e "${YELLOW}⚠ Node.js v$version detected (Vite 5 requires >= 18)${NC}"
            return 1
        fi
    fi
    return 1
}

install_node_official() {
    echo -e "${BLUE}Installing Node.js via nvm (official)...${NC}"
    
    # Check if nvm is already installed
    if [ ! -d "$HOME/.nvm" ]; then
        echo "Installing nvm..."
        if ! curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash; then
            echo -e "${RED}✗ Failed to install nvm${NC}"
            return 1
        fi
    fi
    
    # Load nvm
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
    
    # Install Node.js LTS (v20.x)
    echo "Installing Node.js LTS..."
    if nvm install --lts; then
        nvm use --lts
        nvm alias default lts/*
        
        local installed_version=$(node -v)
        echo -e "${GREEN}✓ Node.js ${installed_version} installed${NC}"
        return 0
    else
        echo -e "${RED}✗ Node.js installation failed${NC}"
        return 1
    fi
}

update_node_version() {
    echo ""
    echo -e "${YELLOW}Vite 5 requires Node.js 18 or higher${NC}"
    read -p "Update Node.js to LTS version? (y/n): " -n 1 -r
    echo ""
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if ! install_node_official; then
            echo -e "${YELLOW}Falling back to NodeSource installation...${NC}"
            echo -e "${YELLOW}This will require sudo password...${NC}"
            echo ""
            
            if curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -; then
                echo ""
                if sudo apt-get install -y nodejs; then
                    local new_version=$(node -v)
                    echo -e "${GREEN}✓ Node.js updated to $new_version${NC}"
                    return 0
                fi
            fi
            
            echo -e "${RED}✗ Failed to update Node.js${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠ Continuing with current Node.js version (may cause issues)${NC}"
        return 1
    fi
}

# Check and install Node.js
if ! check_node_version; then
    # Check if Frontend uses Vite 5
    if [ -d "Frontend" ] && [ -f "Frontend/package.json" ]; then
        if grep -q '"vite".*"5\.' Frontend/package.json 2>/dev/null || \
           grep -q '"vite".*"\^5\.' Frontend/package.json 2>/dev/null; then
            update_node_version
        else
            # No Vite 5, install without asking
            if ! command -v node &>/dev/null; then
                echo -e "${YELLOW}Installing Node.js...${NC}"
                install_node_official || sudo apt install -y nodejs npm -qq
            fi
        fi
    else
        # No Frontend directory, install Node.js anyway
        if ! command -v node &>/dev/null; then
            echo -e "${YELLOW}Installing Node.js...${NC}"
            install_node_official || sudo apt install -y nodejs npm -qq
        fi
    fi
fi

# ============================================
# NPM Check
# ============================================
echo -n "Checking npm... "
if command -v npm &>/dev/null && npm --version &>/dev/null; then
    NPM_VERSION=$(npm --version 2>/dev/null)
    echo -e "${GREEN}✓ npm v${NPM_VERSION}${NC}"
else
    echo -e "${YELLOW}not found${NC}"
    echo -e "${YELLOW}Installing npm...${NC}"
    
    # Load nvm if available
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
    
    if command -v npm &>/dev/null; then
        NPM_VERSION=$(npm --version 2>/dev/null)
        echo -e "${GREEN}✓ npm v${NPM_VERSION}${NC}"
    else
        sudo apt update -qq && sudo apt install -y npm -qq
        if npm --version &>/dev/null; then
            NPM_VERSION=$(npm --version 2>/dev/null)
            echo -e "${GREEN}✓ npm v${NPM_VERSION} installed${NC}"
        else
            echo -e "${RED}✗ npm installation failed${NC}"
            exit 1
        fi
    fi
fi

# ============================================
# Other Dependencies
# ============================================
install_dep() {
    if ! command -v $1 &>/dev/null; then
        echo -e "${YELLOW}Installing $1...${NC}"
        (sudo apt update -qq && sudo apt install -y $2 -qq) &
        show_progress 2 "Downloading"
        wait
        echo -e "${GREEN}✓ $1 installed${NC}"
    else
        echo -e "${GREEN}✓ $1 available${NC}"
    fi
}

install_dep "mysql" "default-mysql-client"

# Rclone
if ! command -v rclone &>/dev/null; then
    echo -e "${YELLOW}Installing rclone...${NC}"
    (curl -fsSL https://rclone.org/install.sh | sudo bash > /dev/null 2>&1) &
    show_progress 4 "Downloading"
    wait
    echo -e "${GREEN}✓ rclone installed${NC}"
else
    echo -e "${GREEN}✓ rclone available${NC}"
fi

# ============================================
# MySQL Setup
# ============================================
echo -e "\n${BLUE}Database Server Setup${NC}"

check_mysql() {
    # Check if MySQL is running
    if systemctl is-active --quiet mysql 2>/dev/null || \
       systemctl is-active --quiet mysqld 2>/dev/null; then
        echo -e "${GREEN}✓ MySQL server running${NC}"
        return 0
    fi
    
    # Check if installed but not running
    if dpkg -l | grep -q mysql-server || dpkg -l | grep -q mariadb-server; then
        echo -e "${YELLOW}Starting MySQL service...${NC}"
        sudo systemctl start mysql 2>/dev/null || sudo systemctl start mysqld
        sleep 3
        return 0
    fi
    
    # MySQL not installed
    return 1
}

mysql_newly_installed=0
if check_mysql; then
    echo -e "${GREEN}✓ MySQL ready${NC}"
else
    # Install MySQL with user-defined credentials
    echo -e "${YELLOW}MySQL not found. Installing MySQL server...${NC}"
    echo ""
    echo -e "${BLUE}Configure MySQL Administrator Account${NC}"
    echo -e "${YELLOW}Set username and password for MySQL${NC}"
    
    # Get username from user
    read -p "MySQL username [root]: " MYSQL_NEW_USER
    MYSQL_NEW_USER=${MYSQL_NEW_USER:-root}
    
    # Get password from user
    while true; do
        read -sp "MySQL password for '$MYSQL_NEW_USER': " MYSQL_NEW_PASS
        echo ""
        
        if [ -z "$MYSQL_NEW_PASS" ]; then
            echo -e "${RED}Password cannot be empty${NC}"
            continue
        fi
        
        if [ ${#MYSQL_NEW_PASS} -lt 6 ]; then
            echo -e "${RED}Password must be at least 6 characters${NC}"
            continue
        fi
        
        read -sp "Confirm password: " MYSQL_NEW_PASS_CONFIRM
        echo ""
        
        if [ "$MYSQL_NEW_PASS" != "$MYSQL_NEW_PASS_CONFIRM" ]; then
            echo -e "${RED}Passwords do not match${NC}"
        else
            break
        fi
    done
    
    # Set debconf for non-interactive install
    echo "mysql-server mysql-server/root_password password $MYSQL_NEW_PASS" | sudo debconf-set-selections
    echo "mysql-server mysql-server/root_password_again password $MYSQL_NEW_PASS" | sudo debconf-set-selections
    
    echo ""
    echo -e "${YELLOW}Installing MySQL server...${NC}"
    sudo apt update -qq
    sudo DEBIAN_FRONTEND=noninteractive apt install -y mysql-server -qq
    
    sudo systemctl start mysql
    sudo systemctl enable mysql
    
    sleep 5
    echo -e "${GREEN}✓ MySQL server installed${NC}"
    
    # If user chose non-root username, create that user
    if [ "$MYSQL_NEW_USER" != "root" ]; then
        echo -e "${YELLOW}Creating user '$MYSQL_NEW_USER'...${NC}"
        mysql -u root -p"$MYSQL_NEW_PASS" -e "CREATE USER IF NOT EXISTS '$MYSQL_NEW_USER'@'localhost' IDENTIFIED BY '$MYSQL_NEW_PASS';" 2>/dev/null
        mysql -u root -p"$MYSQL_NEW_PASS" -e "GRANT ALL PRIVILEGES ON *.* TO '$MYSQL_NEW_USER'@'localhost' WITH GRANT OPTION;" 2>/dev/null
        mysql -u root -p"$MYSQL_NEW_PASS" -e "FLUSH PRIVILEGES;" 2>/dev/null
        echo -e "${GREEN}✓ User '$MYSQL_NEW_USER' created${NC}"
    fi
    
    DB_USER="$MYSQL_NEW_USER"
    DB_PASS="$MYSQL_NEW_PASS"
    mysql_newly_installed=1
fi

# ============================================
# MySQL Authentication
# ============================================
echo -e "\n${BLUE}MySQL User Credentials${NC}"

if [ $mysql_newly_installed -eq 1 ]; then
    echo -e "${GREEN}Using newly created account: $DB_USER${NC}"
else
    echo -e "${YELLOW}Enter your existing MySQL credentials${NC}"
    # Ask for existing MySQL credentials
    attempts=0
    max_attempts=3
    
    while [ $attempts -lt $max_attempts ]; do
        echo ""
        read -p "MySQL username [root]: " DB_USER
        DB_USER=${DB_USER:-root}
        
        read -sp "MySQL password: " DB_PASS
        echo ""
        
        if mysql -u "$DB_USER" -p"$DB_PASS" -e "SELECT 1;" 2>/dev/null; then
            echo -e "${GREEN}✓ Authentication successful${NC}"
            break
        else
            attempts=$((attempts + 1))
            remaining=$((max_attempts - attempts))
            if [ $remaining -gt 0 ]; then
                echo -e "${RED}✗ Invalid credentials ($remaining attempts left)${NC}"
            else
                echo -e "${RED}✗ Maximum attempts reached${NC}"
                exit 1
            fi
        fi
    done
fi

echo -e "${GREEN}✓ Will use MySQL user: $DB_USER${NC}"

# ============================================
# Database Configuration
# ============================================
echo -e "\n${BLUE}Database Configuration${NC}"

read -p "Database name [gbackup-db]: " DB_NAME
DB_NAME=${DB_NAME:-gbackup_db}

read -p "Database host [127.0.0.1]: " DB_HOST
DB_HOST=${DB_HOST:-127.0.0.1}

read -p "Database port [3306]: " DB_PORT
DB_PORT=${DB_PORT:-3306}

# ============================================
# Create Database
# ============================================
echo -e "\n${BLUE}Setting up database...${NC}"

# Create database if not exists
mysql -u "$DB_USER" -p"$DB_PASS" -e "CREATE DATABASE IF NOT EXISTS \`$DB_NAME\`;" 2>/dev/null
echo -e "${GREEN}✓ Database '$DB_NAME' created${NC}"

# Test connection
if mysql -u "$DB_USER" -p"$DB_PASS" -D "$DB_NAME" -e "SELECT 1;" 2>/dev/null; then
    echo -e "${GREEN}✓ Database connection test passed${NC}"
else
    echo -e "${RED}✗ Database connection failed${NC}"
    exit 1
fi

# ============================================
# Rclone Configuration
# ============================================
echo -e "\n${BLUE}Storage Configuration${NC}"

# Auto-detect rclone config path based on current user
CURRENT_USER=$(whoami)
RCLONE_PATH="/home/$CURRENT_USER/.config/rclone/rclone.conf"

# Create rclone config directory if not exists
RCLONE_DIR=$(dirname "$RCLONE_PATH")
if [ ! -d "$RCLONE_DIR" ]; then
    mkdir -p "$RCLONE_DIR"
    echo -e "${YELLOW}Created rclone config directory${NC}"
fi

echo -e "${GREEN}✓ Rclone path: $RCLONE_PATH${NC}"

# ============================================
# Service Port Configuration
# ============================================
echo -e "\n${BLUE}Service Configuration${NC}"

# Function to check if port is available
is_port_available() {
    ! lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1
}

# Find available port
APP_PORT=""
DEFAULT_PORTS=(8080 3000 3001 8081 8082 8083 5173 5174 4000)

echo -n "Finding available port... "
for port in "${DEFAULT_PORTS[@]}"; do
    if is_port_available $port; then
        APP_PORT=$port
        break
    fi
done

# If all default ports are taken, find any available port in range 8000-9000
if [ -z "$APP_PORT" ]; then
    for port in $(seq 8000 9000); do
        if is_port_available $port; then
            APP_PORT=$port
            break
        fi
    done
fi

# If still no port found, ask user
if [ -z "$APP_PORT" ]; then
    echo ""
    echo -e "${YELLOW}All default ports (8000-9000) are in use${NC}"
    while true; do
        read -p "Enter custom port (1024-65535): " APP_PORT
        
        # Validate port number
        if ! [[ "$APP_PORT" =~ ^[0-9]+$ ]]; then
            echo -e "${RED}✗ Invalid port number${NC}"
            continue
        fi
        
        if [ "$APP_PORT" -lt 1024 ] || [ "$APP_PORT" -gt 65535 ]; then
            echo -e "${RED}✗ Port must be between 1024-65535${NC}"
            continue
        fi
        
        if ! is_port_available $APP_PORT; then
            echo -e "${RED}✗ Port $APP_PORT is already in use${NC}"
            continue
        fi
        
        break
    done
else
    echo "Done"
fi

# Generate JWT secret
JWT_SECRET=$(openssl rand -base64 32)

echo -e "${GREEN}✓ Service port: $APP_PORT${NC}"
echo -e "${GREEN}✓ Security token generated${NC}"

# ============================================
# Create Environment File
# ============================================
echo -e "\n${BLUE}Creating configuration file...${NC}"

cat > Backend/.env << EOF
# Database Configuration
DB_USER=$DB_USER
DB_PASS=$DB_PASS
DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_NAME=$DB_NAME

# Security
JWT_SECRET_KEY=$JWT_SECRET

# Storage
RCLONE_CONFIG=$RCLONE_PATH

# Network
SERVER_HOST=0.0.0.0
APP_PORT=$APP_PORT
EOF

echo -e "${GREEN}✓ Configuration created${NC}"

# ============================================
# Build Backend Application
# ============================================
echo -e "\n${BLUE}Building application...${NC}"

cd Backend

# Fix go.mod version format if needed (remove patch version)
if grep -q 'go [0-9]\+\.[0-9]\+\.[0-9]\+' go.mod 2>/dev/null; then
    echo -e "${YELLOW}Fixing go.mod version format...${NC}"
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
    cd ..
    exit 1
fi
echo " Done"

go mod tidy > /dev/null 2>&1

# Build the application
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
    cd ..
    exit 1
fi
echo " Done"

echo -e "${GREEN}✓ Application built${NC}"
cd ..

# ============================================
# Install Frontend (if exists)
# ============================================
if [ -d "Frontend" ] && [ -f "Frontend/package.json" ]; then
    echo -e "${BLUE}Installing frontend...${NC}"
    
    # Load nvm if available
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
    
    # Double-check npm is working
    if ! npm --version &>/dev/null; then
        echo -e "${RED}✗ npm is not working${NC}"
        exit 1
    fi
    
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
        echo -e "${RED}✗ Frontend installation failed${NC}"
        cat /tmp/npm_install.log
        cd ..
        exit 1
    fi
    echo " Done"
    
    echo -e "${GREEN}✓ Frontend ready${NC}"
    cd ..
fi

# ============================================
# Installation Complete
# ============================================

# Get the correct network IP (not NAT/loopback)
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
        return 1
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

echo -e "\n${GREEN}✓ System successfully installed${NC}"
echo ""

# Check if using NAT IP (status = 1)
if [ $IP_STATUS -eq 1 ]; then
    echo -e "${BLUE}Access URL:${NC}"
    echo "  http://$SERVER_IP:$APP_PORT"
    echo ""
    echo -e "${YELLOW}⚠ Network Notice:${NC}"
    echo "  Using NAT IP - may not be accessible from other devices"
    echo "  Fix: VirtualBox → Settings → Network → Bridged Adapter"
    
elif [ $IP_COUNT -gt 1 ]; then
    echo -e "${BLUE}Access URL (multiple networks):${NC}"
    while IFS= read -r ip; do
        if [ "$ip" = "$SERVER_IP" ]; then
            echo -e "  ${GREEN}➜ http://$ip:$APP_PORT${NC} (primary)"
        else
            echo -e "    http://$ip:$APP_PORT"
        fi
    done <<< "$ALL_IPS"
else
    echo -e "${BLUE}Access URL:${NC}"
    echo "  http://$SERVER_IP:$APP_PORT"
fi
echo ""

# Show instructions to reload shell
echo -e "${YELLOW}Note: If Go or Node.js commands are not found, please run:${NC}"
echo -e "${YELLOW}  source ~/.bashrc${NC}"
echo ""