package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Logger provides minimal, human-readable logging
type Logger struct {
	mu     sync.Mutex
	writer io.Writer
}

var defaultLogger = &Logger{
	writer: os.Stderr,
}

// SetLogPath configures the logger to write to a file
func (l *Logger) SetLogPath(logPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logPath, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file (append mode)
	logFile := filepath.Join(logPath, "pipeline.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Write to both file and stderr
	l.writer = io.MultiWriter(os.Stderr, file)
	return nil
}

// Log writes a formatted log line
// Format: TIMESTAMP | LEVEL | COMPONENT | MESSAGE | CONTEXT
func (l *Logger) Log(level, component, message, context string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339)
	line := fmt.Sprintf("%s | %s | %s | %s", timestamp, level, component, message)
	if context != "" {
		line += " | " + context
	}
	line += "\n"

	l.writer.Write([]byte(line))
}

// Info logs at INFO level
func (l *Logger) Info(component, message, context string) {
	l.Log("INFO", component, message, context)
}

// Warn logs at WARN level
func (l *Logger) Warn(component, message, context string) {
	l.Log("WARN", component, message, context)
}

// Error logs at ERROR level
func (l *Logger) Error(component, message, context string) {
	l.Log("ERROR", component, message, context)
}

// Fatal logs at FATAL level and exits
func (l *Logger) Fatal(component, message, context string) {
	l.Log("FATAL", component, message, context)
	os.Exit(1)
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	return defaultLogger
}

// HealthMetrics holds system health metrics
type HealthMetrics struct {
	MemoryRSS      uint64 // Resident Set Size in bytes
	MemoryHeap     uint64 // Heap size in bytes
	BufferBytes    int64  // Current buffer size in bytes
	BufferPercent  int    // Buffer usage percentage
	ThroughputEPS  int64  // Events per second
	FlushCount     int64  // Total flush count
	ErrorCount     int64  // Total error count
	KafkaLag       int64  // Kafka consumer lag (if available)
}

// GetMemoryStats collects memory usage statistics
func GetMemoryStats() (rss, heap uint64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	heap = m.Alloc

	// Try to get RSS from /proc/self/status (Linux)
	if f, err := os.Open("/proc/self/status"); err == nil {
		defer f.Close()
		buf := make([]byte, 4096)
		n, _ := f.Read(buf)
		content := string(buf[:n])
		// Parse VmRSS line
		var rssKB uint64
		fmt.Sscanf(content, "VmRSS: %d kB", &rssKB)
		rss = rssKB * 1024
	}

	return rss, heap
}

