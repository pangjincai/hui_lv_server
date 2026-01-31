package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DailyWriter implements io.Writer to support daily log rotation
type DailyWriter struct {
	LogName  string // Base log name (e.g., "server" or "/var/log/app")
	file     *os.File
	lastDate string
	mu       sync.Mutex
}

// NewDailyWriter creates a new DailyWriter
func NewDailyWriter(logName string) *DailyWriter {
	return &DailyWriter{
		LogName: logName,
	}
}

func (w *DailyWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	currentDate := time.Now().Format("2006-01-02")

	// If date changed or file is not open, rotate/open file
	if currentDate != w.lastDate || w.file == nil {
		if err := w.rotate(currentDate); err != nil {
			// Fallback to stderr if cannot open file, or return error
			fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
			return 0, err
		}
	}

	return w.file.Write(p)
}

func (w *DailyWriter) rotate(date string) error {
	// Close existing file if open
	if w.file != nil {
		w.file.Close()
	}

	// Construct new filename: LogName_YYYY-MM-DD.log
	// If LogName has extension, we might want to handle it, but for simplicity
	// we assume LogName is just the base path/name.
	// Example: LogName="server" -> "server_2026-01-30.log"
	
	// Check if LogName has an extension and insert date before it?
	// Or just append. User example: "hui_lv_server_2026-01-30.log"
	// implying input "hui_lv_server" -> "hui_lv_server_DATE.log"
	
	ext := filepath.Ext(w.LogName)
	base := w.LogName[:len(w.LogName)-len(ext)]
	if ext == "" {
		ext = ".log"
	}
	
	filename := fmt.Sprintf("%s_%s%s", base, date, ext)

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	w.file = f
	w.lastDate = date
	
	// Clean old logs asynchronously
	go w.cleanOldLogs(base, ext)
	
	return nil
}

func (w *DailyWriter) cleanOldLogs(base, ext string) {
	// Keep logs for 7 days
	const maxAge = 7 * 24 * time.Hour
	cutoff := time.Now().Add(-maxAge)

	// Find matching files
	pattern := fmt.Sprintf("%s_*%s", base, ext)
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error cleaning logs (Glob): %v\n", err)
		return
	}

	for _, file := range files {
		// Extract date from filename: base_YYYY-MM-DD.ext
		// file is full path
		filename := filepath.Base(file)
		baseName := filepath.Base(base)
		
		// Expected format: baseName_YYYY-MM-DD.ext
		// Length check: baseName + 1 + 10 (date) + ext
		expectedLen := len(baseName) + 1 + 10 + len(ext)
		if len(filename) != expectedLen {
			continue
		}

		dateStr := filename[len(baseName)+1 : len(baseName)+11]
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		if fileDate.Before(cutoff) {
			// Delete file
			if err := os.Remove(file); err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting old log %s: %v\n", file, err)
			} else {
				fmt.Printf("Deleted old log file: %s\n", file)
			}
		}
	}
}

// Close closes the underlying file
func (w *DailyWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
