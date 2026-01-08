package models

// FileItem merepresentasikan file/folder di cloud storage
type FileItem struct {
	Name     string `json:"name"`      // Nama file/folder
	Path     string `json:"path"`      // Full path
	IsDir    bool   `json:"is_dir"`    // Apakah folder?
	Size     int64  `json:"size"`      // Ukuran dalam bytes
	ModTime  string `json:"mod_time"`  // Waktu modifikasi
	MimeType string `json:"mime_type"` // MIME type (untuk file)
}

// BrowserResponse response dari browse API
type BrowserResponse struct {
	Path      string     `json:"path"`       // Path saat ini
	Files     []FileItem `json:"files"`      // List file/folder
	TotalSize int64      `json:"total_size"` // Total size
}
