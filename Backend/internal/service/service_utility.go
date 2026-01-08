package service

import "sync"

// LogMutex adalah Mutex (Lock) yang melindungi akses ke LogRepository
// saat Goroutine (Job) menulis ke database secara bersamaan.
var LogMutex sync.Mutex
