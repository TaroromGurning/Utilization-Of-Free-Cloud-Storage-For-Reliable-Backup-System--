package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ✅ BARU: Parse rclone output untuk extract TransferredBytes
func parseTransferredBytes(output string) int64 {
	// Regex: Mencari angka dan unit (e.g., 12.5 GiB)
	re := regexp.MustCompile(`Transferred:\s+([\d\.]+)\s*(\w+i?B)`)

	// Ambil SEMUA match
	matches := re.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return 0
	}

	// Ambil match TERAKHIR (status final)
	lastMatch := matches[len(matches)-1]

	valueStr := lastMatch[1]
	unit := strings.ToUpper(lastMatch[2])

	value, _ := strconv.ParseFloat(valueStr, 64)

	// Konversi Unit
	switch unit {
	case "B":
		return int64(value)
	case "KB", "KIB":
		return int64(value * 1024)
	case "MB", "MIB":
		return int64(value * 1024 * 1024)
	case "GB", "GIB":
		return int64(value * 1024 * 1024 * 1024)
	default:
		return int64(value)
	}
}

// ✅ HELPER: Extract file size dari string format rclone
func extractFirstFileSize(text string) int64 {
	pattern := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(MiB|Mi|GiB|Gi|KiB|Ki|MB|GB|KB|Bytes?)`)
	matches := pattern.FindStringSubmatch(text)

	if len(matches) < 3 {
		return 0
	}

	// Parse nilai (misal "2.5")
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	// Parse unit (misal "GBytes")
	unit := strings.ToLower(matches[2])

	var multiplier int64 = 1
	switch unit {
	case "GiB", "GB":
		return int64(value * 1024 * 1024 * 1024)
	case "MiB", "MB":
		return int64(value * 1024 * 1024)
	case "KiB", "KB":
		return int64(value * 1024)
	case "Bytes", "Byte":
		return int64(value)
	}

	return int64(value * float64(multiplier))
}

// ✅ BARU: Parse rclone output untuk extract statistik lengkap
func parseRcloneStats(output string) RcloneStats {
	stats := RcloneStats{}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Transferred
		if strings.HasPrefix(line, "Transferred:") {
			stats.TransferredBytes = extractFirstFileSize(line)
			stats.TransferredLine = line
		}

		// Checks
		if strings.HasPrefix(line, "Checks:") {
			stats.ChecksLine = line
		}

		// Errors
		if strings.HasPrefix(line, "Errors:") {
			stats.ErrorCount = extractNumber(line)
		}

		// Speed
		if strings.Contains(line, "MBytes/s") || strings.Contains(line, "KBytes/s") {
			stats.Speed = extractSpeed(line)
		}
	}

	return stats
}

// ✅ HELPER: Extract number dari string
func extractNumber(text string) int {
	pattern := regexp.MustCompile(`\d+`)
	matches := pattern.FindString(text)
	if matches == "" {
		return 0
	}
	num, _ := strconv.Atoi(matches)
	return num
}

// ✅ HELPER: Extract speed dari string
func extractSpeed(text string) string {
	// Contoh: "10 MBytes/s" atau "500 KBytes/s"
	pattern := regexp.MustCompile(`(\d+\.?\d*)\s*(MBytes|KBytes|Bytes)/s`)
	matches := pattern.FindStringSubmatch(text)

	if len(matches) < 3 {
		return ""
	}

	return fmt.Sprintf("%s %s/s", matches[1], matches[2])
}

// ✅ STRUCT: Untuk hold rclone statistics
type RcloneStats struct {
	TransferredBytes int64  // Total bytes yang di-transfer
	TransferredLine  string // Full line (untuk logging)
	ChecksLine       string
	ErrorCount       int
	Speed            string
}
