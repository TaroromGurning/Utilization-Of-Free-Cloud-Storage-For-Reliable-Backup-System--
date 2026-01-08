# G-Backup

Solusi manajemen backup data otomatis berbasis web yang menjembatani server dengan Google Drive.

## Fitur Utama

- **Script Runner Pipeline**: Eksekusi backup melalui 3 fase (Pre-Script, Rclone Execution, Post-Script)
- **Automated Scheduling**: Penjadwalan backup berbasis CRON dengan background worker Golang
- **Proactive Monitoring**: Dashboard visual untuk status koneksi GDrive, metrik storage, dan log eksekusi
- **Simple Restore**: Mekanisme pengembalian data dengan path inversion otomatis

## Arsitektur Sistem

Sistem membagi tanggung jawab ke dalam 3 lapisan:

- **Presentation Tier**: Vue.js 3 + Pinia untuk UI dan state management
- **Application Tier**: Golang + Echo Framework dengan pola MSC (Model-Service-Controller)
- **Data Tier**: MariaDB untuk metadata dan Rclone CLI untuk operasi file ke Google Drive

## Teknologi

**Backend**
- Golang 1.18+
- Echo Framework
- GORM

**Frontend**
- Vue.js 3
- Vite
- Axios
- Pinia

**Infrastructure**
- MariaDB/MySQL
- Rclone CLI
- Ubuntu/Linux

## Prasyarat

- Go (v1.18+)
- Node.js (v16+)
- MariaDB/MySQL
- Rclone (sudah dikonfigurasi dengan Google Drive)

### Setup Rclone
```bash
curl https://rclone.org/install.sh | sudo bash
rclone config
rclone ls gdrive:
```

## Instalasi

### 1. Clone Repository
```bash
git clone https://github.com/username/g-backup.git
cd g-backup
```

### 2. Setup Backend
```bash
cd backend
cp .env.example .env
nano .env
go mod tidy
go run main.go
```

Contoh `.env`:
```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=gbackup

JWT_SECRET=your-super-secret-key
JWT_EXPIRY=24h

RCLONE_REMOTE=gdrive
RCLONE_PATH=/backups

APP_PORT=8080
APP_ENV=development
```

### 3. Setup Frontend
```bash
cd frontend
npm install
cp .env.example .env
nano .env
npm run dev
```

Contoh `.env` Frontend:
```env
VITE_API_URL=http://localhost:8080/api/v1
VITE_APP_NAME=G-Backup
```

### 4. Build Production
```bash
# Backend
cd backend
go build -o gbackup-server

# Frontend
cd frontend
npm run build
```

## Quick Start

1. Akses aplikasi: `http://localhost:5173`
2. Login dengan kredensial default (admin/admin123)
3. Ubah password setelah login pertama
4. Tambahkan job backup pertama
5. Monitor eksekusi melalui dashboard

## Author
Yehezkiel-Rumapea - Lead Developer