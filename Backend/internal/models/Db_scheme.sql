-- =========================================================
-- 1. Tabel users (Otentikasi dan Manajemen Akses)
-- =========================================================
CREATE TABLE users (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- =========================================================
-- 2. Tabel scheduled_jobs (Konfigurasi Job "Script Runner")
-- =========================================================
CREATE TABLE scheduled_jobs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL,
    job_name VARCHAR(100) NOT NULL,
    
    -- [PERBAIKAN 1]: Tambahkan OperationMode
    operation_mode ENUM('BACKUP', 'RESTORE') NOT NULL, 
    
    -- [PERBAIKAN 2]: RcloneMode - Ubah ke huruf kecil
    rclone_mode ENUM('copy', 'sync') NOT NULL DEFAULT 'copy', 
    source_path VARCHAR(255) NOT NULL, 
    remote_name VARCHAR(100) NOT NULL,
    destination_path VARCHAR(255) NOT NULL, 

    -- Script Kustom
    pre_script TEXT NULL, 
    post_script TEXT NULL, 
    
    -- Penjadwalan dan Status
    schedule_cron VARCHAR(50) NULL, 
    priority INT DEFAULT 5,
    
    -- [PERBAIKAN 3]: StatusQueue - Sesuaikan dengan label panjang di kode Go
    status_queue ENUM(
        'PENDING',
        'RUNNING',
        'COMPLETED',
        'FAIL_PRE_SCRIPT',
        'FAIL_RCLONE',
        'FAIL_POST_SCRIPT'
    ) DEFAULT 'PENDING',
    
    is_active BOOLEAN DEFAULT TRUE,
    last_run_at DATETIME NULL,
    
    -- FOREIGN KEY (user_id) REFERENCES users(id) -- Asumsi tabel users sudah ada
);

-- =========================================================
-- 3. Tabel logs (Riwayat Eksekusi)
-- =========================================================
CREATE TABLE logs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    job_id INT UNSIGNED NULL, 
    
    -- Status lebih detail untuk Eksekusi 3 Fase
    status ENUM('SUCCESS', 'FAIL_PRE_SCRIPT', 'FAIL_RCLONE', 'FAIL_POST_SCRIPT', 'ERROR') NOT NULL,
    
    config_snapshot JSON NULL, -- (Tetap ada untuk Job Manual)
    message TEXT, -- (Output mentah dari bash -c)
    duration_sec INT NULL,                      
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (job_id) REFERENCES scheduled_jobs(id) ON DELETE SET NULL
);

-- =========================================================
-- 4. Tabel monitoring (Status Cloud Storage)
-- =========================================================
CREATE TABLE monitoring (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    remote_name VARCHAR(100) UNIQUE NOT NULL,   
    status_connect ENUM('CONNECTED', 'DISCONNECTED') DEFAULT 'DISCONNECTED',
    total_storage_gb FLOAT NULL,
    used_storage_gb FLOAT NULL,
    free_storage_gb FLOAT NULL,
    last_checked_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- =========================================================
-- 5. Tabel remotes (Fitur "Add Remote" - Opsional)
-- =========================================================
CREATE TABLE remotes (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    remote_type VARCHAR(50) NOT NULL,
    config_json JSON NULL, -- Menyimpan config non-sensitif (opsional)
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);