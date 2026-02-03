# G-Backup

***Web-based automated data backup management solution that bridges servers with Google Drive.***

---

## Key Features

- ***Script Runner Pipeline***: Execute backups through 3 phases (Pre-Script, Rclone Execution, Post-Script)
- ***Automated Scheduling***: CRON-based backup scheduling with Golang background worker
- ***Proactive Monitoring***: Visual dashboard for GDrive connection status, storage metrics, and execution logs
- ***Simple Restore***: Data recovery mechanism with automatic path inversion

---

## System Architecture
![Architecture Diagram](./HLA.png)

---

## Technology Stack

### Backend
- Golang (>= 1.20)
- Echo Framework
- GORM

### Frontend
- Vue.js 
- Vite (>= 5.0 requires Node.js >= 18)
- Axios
- Pinia

### Infrastructure
- MariaDB/MySQL
- Rclone CLI
- Linux (Debian/Ubuntu or RHEL/Fedora)

---

---

## Installation

### Automatic Installation (Recommended)

The installer will automatically detect your Linux distribution and use the appropriate package manager.

#### 1. Clone Repository
```bash
git clone https://github.com/username/G-Backup.git
cd g-backup
```

#### 2. Set File Permissions
```bash
chmod +x install.sh
```

#### 3. Execute Installation
```bash
./install.sh
```

The script will:
- Auto-detect your Linux distribution (Debian/Ubuntu or RHEL/Fedora)
- Install all required dependencies
- Configure MySQL database
- Build backend and frontend applications
- Generate secure configuration files

---

### Manual Installation

Use this method if you want to preserve existing server configurations or have specific requirements.

#### Prerequisites Check
```bash
# Check Go version (requires >= 1.20)
go version

# Check Node.js version (requires >= 18 for Vite 5)
node -v

# Check npm
npm -v

# Check MySQL/MariaDB
systemctl status mysql
# or
systemctl status mysqld
```

---

### Manual Installation - Debian/Ubuntu

#### 1. Install Required Dependencies

**Install Go (if not installed or version < 1.20)**
```bash
# Download Go 1.23.5 (adjust architecture if needed)
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz

# Remove old version and install new
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

**Install Node.js (if not installed or version < 18)**
```bash
# Using nvm (recommended)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash
source ~/.bashrc
nvm install --lts
nvm use --lts

# Verify installation
node -v
npm -v
```

**Install MySQL Client**
```bash
sudo apt update
sudo apt install -y default-mysql-client
```

**Install Rclone**
```bash
curl https://rclone.org/install.sh | sudo bash
```

#### 2. Configure MySQL Database

**Option A: Use Existing MySQL Installation**
```bash
# Login to MySQL with your existing credentials
mysql -u your_username -p

# Create database for G-Backup
CREATE DATABASE IF NOT EXISTS gbackup_db;

# Verify connection
USE gbackup_db;
SELECT DATABASE();
exit;
```

**Option B: Install Fresh MySQL Server**
```bash
sudo apt update
sudo apt install -y mysql-server

# Start and enable MySQL
sudo systemctl start mysql
sudo systemctl enable mysql

# Secure installation
sudo mysql_secure_installation

# Create database
mysql -u root -p
CREATE DATABASE IF NOT EXISTS gbackup_db;
exit;
```

#### 3. Configure Rclone

**Preserve Existing Rclone Configuration**
```bash
# Check existing rclone config location
rclone config file

# Typical location: /home/username/.config/rclone/rclone.conf
# Note this path for the .env file
```

**Setup New Rclone for Google Drive (if needed)**
```bash
rclone config

# Follow prompts:
# n) New remote
# name> gdrive
# Storage> drive (Google Drive)
# client_id> (press Enter for default)
# client_secret> (press Enter for default)
# scope> 1 (Full access)
# root_folder_id> (press Enter)
# service_account_file> (press Enter)
# Edit advanced config? n
# Use auto config? y (follow browser authentication)
# Configure as team drive? n
# Confirm? y

# Verify configuration
rclone listremotes
```

#### 4. Create Backend Environment File
```bash
cd Backend

# Create .env file
nano .env
```

Add the following configuration (adjust values to match your setup):
```env
# Database Configuration (use your existing MySQL credentials)
DB_USER=your_mysql_username
DB_PASS=your_mysql_password
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=gbackup_db

# Security (generate random JWT secret)
JWT_SECRET_KEY=your_random_secret_key_here

# Storage (use your existing rclone config path)
RCLONE_CONFIG=/home/your_username/.config/rclone/rclone.conf

# Network
SERVER_HOST=0.0.0.0
APP_PORT=8080
```

**Generate JWT Secret Key:**
```bash
openssl rand -base64 32
```

#### 5. Build Backend Application
```bash
cd Backend

# Fix go.mod version format if needed
sed -i 's/go \([0-9]\+\.[0-9]\+\)\.[0-9]\+/go \1/' go.mod

# Download dependencies
go mod download
go mod tidy

# Build application
go build -o app ./cmd/main.go

# Verify build
ls -lh app
```

#### 6. Install Frontend Dependencies
```bash
cd ../Frontend

# Install packages
npm install

# Verify installation
npm list --depth=0
```

#### 7. Verify Installation
```bash
# Test database connection
mysql -u your_mysql_username -p -D gbackup_db -e "SELECT 1;"

# Test rclone
rclone listremotes

# Test backend build
cd ../Backend
./app --version || echo "Backend binary ready"

# Test frontend
cd ../Frontend
npm run build
```

---

### Manual Installation - RHEL/Fedora/CentOS

#### 1. Install Required Dependencies

**Install Go (if not installed or version < 1.20)**
```bash
# Download Go 1.23.5 (adjust architecture if needed)
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz

# Remove old version and install new
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

**Install Node.js (if not installed or version < 18)**
```bash
# Using nvm (recommended)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash
source ~/.bashrc
nvm install --lts
nvm use --lts

# Alternative: Using NodeSource repository
curl -fsSL https://rpm.nodesource.com/setup_lts.x | sudo bash -
sudo dnf install -y nodejs

# Verify installation
node -v
npm -v
```

**Install MySQL Client**
```bash
sudo dnf install -y mysql
```

**Install Rclone**
```bash
curl https://rclone.org/install.sh | sudo bash
```

#### 2. Configure MySQL Database

**Option A: Use Existing MySQL Installation**
```bash
# Login to MySQL with your existing credentials
mysql -u your_username -p

# Create database for G-Backup
CREATE DATABASE IF NOT EXISTS gbackup_db;

# Verify connection
USE gbackup_db;
SELECT DATABASE();
exit;
```

**Option B: Install Fresh MySQL Server**
```bash
sudo dnf install -y mysql-server

# Start and enable MySQL
sudo systemctl start mysqld
sudo systemctl enable mysqld

# Get temporary root password (MySQL 8.0+)
sudo grep 'temporary password' /var/log/mysqld.log

# Secure installation
sudo mysql_secure_installation

# Login and create database
mysql -u root -p
CREATE DATABASE IF NOT EXISTS gbackup_db;
exit;
```

#### 3. Configure Rclone

**Preserve Existing Rclone Configuration**
```bash
# Check existing rclone config location
rclone config file

# Typical location: /home/username/.config/rclone/rclone.conf
# Note this path for the .env file
```

**Setup New Rclone for Google Drive (if needed)**
```bash
rclone config

# Follow prompts:
# n) New remote
# name> gdrive
# Storage> drive (Google Drive)
# client_id> (press Enter for default)
# client_secret> (press Enter for default)
# scope> 1 (Full access)
# root_folder_id> (press Enter)
# service_account_file> (press Enter)
# Edit advanced config? n
# Use auto config? y (follow browser authentication)
# Configure as team drive? n
# Confirm? y

# Verify configuration
rclone listremotes
```

#### 4. Create Backend Environment File
```bash
cd Backend

# Create .env file
nano .env
```

Add the following configuration (adjust values to match your setup):
```env
# Database Configuration (use your existing MySQL credentials)
DB_USER=your_mysql_username
DB_PASS=your_mysql_password
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=gbackup_db

# Security (generate random JWT secret)
JWT_SECRET_KEY=your_random_secret_key_here

# Storage (use your existing rclone config path)
RCLONE_CONFIG=/home/your_username/.config/rclone/rclone.conf

# Network
SERVER_HOST=0.0.0.0
APP_PORT=8080
```

**Generate JWT Secret Key:**
```bash
openssl rand -base64 32
```

#### 5. Build Backend Application
```bash
cd Backend

# Fix go.mod version format if needed
sed -i 's/go \([0-9]\+\.[0-9]\+\)\.[0-9]\+/go \1/' go.mod

# Download dependencies
go mod download
go mod tidy

# Build application
go build -o app ./cmd/main.go

# Verify build
ls -lh app
```

#### 6. Install Frontend Dependencies
```bash
cd ../Frontend

# Install packages
npm install

# Verify installation
npm list --depth=0
```

#### 7. Verify Installation
```bash
# Test database connection
mysql -u your_mysql_username -p -D gbackup_db -e "SELECT 1;"

# Test rclone
rclone listremotes

# Test backend build
cd ../Backend
./app --version || echo "Backend binary ready"

# Test frontend
cd ../Frontend
npm run build
```

#### 8. Configure Firewall (if needed)
```bash
# Allow application port through firewall
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=5173/tcp
sudo firewall-cmd --reload

# Verify firewall rules
sudo firewall-cmd --list-all
```

---

### Important Notes for Manual Installation

1. **Preserve Existing Data:**
   - Database: Use existing MySQL credentials instead of creating new ones
   - Rclone: Reference your existing `rclone.conf` path in `.env`
   - No existing configurations will be overwritten

2. **Version Requirements:**
   - Go >= 1.20 (required for modern Go features)
   - Node.js >= 18 (required for Vite 5)
   - MySQL/MariaDB (any recent version)

3. **Security Considerations:**
   - Keep your `.env` file secure (add to `.gitignore`)
   - Use strong MySQL passwords (min 8 characters, mixed case, numbers, symbols)
   - Generate unique JWT secret keys
   - For RHEL/Fedora: Configure SELinux policies if needed

4. **Troubleshooting:**

   **Debian/Ubuntu:**
```bash
   # If Go commands not found after installation
   source ~/.bashrc
   
   # If npm commands not found
   source ~/.nvm/nvm.sh
   nvm use --lts
   
   # If MySQL connection fails
   sudo systemctl status mysql
   mysql -u root -p -e "SELECT user, host FROM mysql.user;"
```

   **RHEL/Fedora/CentOS:**
```bash
   # If Go commands not found after installation
   source ~/.bashrc
   
   # If npm commands not found
   source ~/.nvm/nvm.sh
   nvm use --lts
   
   # If MySQL connection fails
   sudo systemctl status mysqld
   mysql -u root -p -e "SELECT user, host FROM mysql.user;"
   
   # If SELinux blocks MySQL connections
   sudo setsebool -P httpd_can_network_connect_db 1
   
   # Check firewall status
   sudo firewall-cmd --list-all
```

5. **Port Configuration:**
```bash
   # Check if default port 8080 is available
   # Debian/Ubuntu
   lsof -i :8080
   
   # RHEL/Fedora
   ss -tuln | grep :8080
   
   # If port is in use, modify APP_PORT in Backend/.env
   # Choose an available port (e.g., 3000, 8081, etc.)
```

---

## Starting the Application

### 1. Set File Permissions
```bash
chmod +x start.sh
```

### 2. Run the Application
```bash
./start.sh
```

---

## Application Access (Example)
```bash
✓ Application is up to date 
✓ System started successfully 

Access URL: 
➜ API Service: http://192.168.1.100:8080 (primary) 
➜ Web Interface: http://192.168.1.100:5173 (primary) 

Press Ctrl+C to stop
```

---

## Configuration Reference

### Environment Variables (.env)
| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| DB_USER | MySQL username | root | Yes |
| DB_PASS | MySQL password | your_password | Yes |
| DB_HOST | Database host | 127.0.0.1 | Yes |
| DB_PORT | Database port | 3306 | Yes |
| DB_NAME | Database name | gbackup_db | Yes |
| JWT_SECRET_KEY | JWT signing key | random_base64_string | Yes |
| RCLONE_CONFIG | Rclone config path | /home/user/.config/rclone/rclone.conf | Yes |
| SERVER_HOST | Server bind address | 0.0.0.0 | Yes |
| APP_PORT | Application port | 8080 | Yes |

### Default Ports
- **Backend API:** 8080
- **Frontend Dev Server:** 5173
- **MySQL:** 3306

### Rclone Remote Names
When configuring Rclone, use meaningful remote names:
- **gdrive** - For Google Drive
- **gdrive-work** - For work Google Drive account
- **gdrive-personal** - For personal Google Drive account

---

## Distribution-Specific Notes

### Debian/Ubuntu
- Uses `apt` package manager
- MySQL service name: `mysql`
- Default shell: `bash`
- SELinux: Not enabled by default

### RHEL/Fedora/CentOS
- Uses `dnf` or `yum` package manager
- MySQL service name: `mysqld`
- Default shell: `bash`
- SELinux: Enabled by default (may require policy configuration)
- Firewall: `firewalld` active by default

---

## Quick Start Commands by Distribution

### Debian/Ubuntu
```bash
# Clone and install
git clone https://github.com/username/G-Backup.git
cd g-backup
chmod +x install.sh
./install.sh

# Start application
chmod +x start.sh
./start.sh
```

### RHEL/Fedora/CentOS
```bash
# Clone and install
git clone https://github.com/username/G-Backup.git
cd g-backup
chmod +x install.sh
./install.sh

# Configure firewall (if needed)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=5173/tcp
sudo firewall-cmd --reload

# Start application
chmod +x start.sh
./start.sh
```

---

## Support

For issues, questions, or contributions:
- **Email:** rumapeayehezkiel@gmail.com

---

## License

[Your License Here]
