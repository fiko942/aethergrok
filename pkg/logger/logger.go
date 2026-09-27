package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogEntry struct {
	ID        string      `json:"id"`
	Timestamp int64       `json:"timestamp"`
	Level     string      `json:"level"`
	Category  string      `json:"category"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
}

const (
	maxLogFileSize = 5 * 1024 * 1024 // 5 MB per file
	maxLogBackups  = 3               // Keep up to 3 rotated archives (aethergrok.1.log .. aethergrok.3.log)
)

type DiskLogger struct {
	mu      sync.Mutex
	logDir  string
	logFile string
	file    *os.File
}

var (
	defaultLogger *DiskLogger
	once          sync.Once
)

func GetDiskLogger() *DiskLogger {
	once.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil {
			home = os.TempDir()
		}
		dir := filepath.Join(home, ".grok", "logs")
		_ = os.MkdirAll(dir, 0755)

		filePath := filepath.Join(dir, "aethergrok.log")
		defaultLogger = &DiskLogger{
			logDir:  dir,
			logFile: filePath,
		}
		_ = defaultLogger.openFile()
	})
	return defaultLogger
}

func (l *DiskLogger) openFile() error {
	f, err := os.OpenFile(l.logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	l.file = f
	return nil
}

func (l *DiskLogger) Append(entry LogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if entry.Timestamp == 0 {
		entry.Timestamp = time.Now().UnixMilli()
	}

	bytes, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')

	// Check file size for rotation
	if l.file != nil {
		fi, err := l.file.Stat()
		if err == nil && fi.Size()+int64(len(bytes)) > maxLogFileSize {
			l.rotate()
		}
	} else {
		if err := l.openFile(); err != nil {
			return err
		}
	}

	if l.file != nil {
		_, err = l.file.Write(bytes)
		return err
	}
	return nil
}

func (l *DiskLogger) rotate() {
	if l.file != nil {
		_ = l.file.Close()
		l.file = nil
	}

	// Rotate backups: .2 -> .3, .1 -> .2, active -> .1
	for i := maxLogBackups - 1; i >= 1; i-- {
		src := filepath.Join(l.logDir, fmt.Sprintf("aethergrok.%d.log", i))
		dst := filepath.Join(l.logDir, fmt.Sprintf("aethergrok.%d.log", i+1))
		if _, err := os.Stat(src); err == nil {
			_ = os.Rename(src, dst)
		}
	}

	firstBackup := filepath.Join(l.logDir, "aethergrok.1.log")
	if _, err := os.Stat(l.logFile); err == nil {
		_ = os.Rename(l.logFile, firstBackup)
	}

	_ = l.openFile()
}

func (l *DiskLogger) ReadRecentEntries(limit int) ([]LogEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 {
		limit = 1000
	}

	// We read from active log file and, if needed, fallback to .1.log
	var entries []LogEntry
	filesToRead := []string{
		l.logFile,
		filepath.Join(l.logDir, "aethergrok.1.log"),
	}

	// Read lines backwards or gather all then take last `limit`
	var allRawEntries []LogEntry
	for _, fp := range filesToRead {
		if _, err := os.Stat(fp); err != nil {
			continue
		}
		f, err := os.Open(fp)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		// Allocate buffer up to 1MB per line
		buf := make([]byte, 1024*1024)
		scanner.Buffer(buf, 1024*1024)

		var fileEntries []LogEntry
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			var entry LogEntry
			if err := json.Unmarshal(line, &entry); err == nil {
				fileEntries = append(fileEntries, entry)
			}
		}
		_ = f.Close()

		if fp == l.logFile {
			allRawEntries = append(fileEntries, allRawEntries...)
		} else {
			// Older backup logs precede current logs
			allRawEntries = append(fileEntries, allRawEntries...)
		}
	}

	total := len(allRawEntries)
	if total <= limit {
		entries = allRawEntries
	} else {
		entries = allRawEntries[total-limit:]
	}

	return entries, nil
}

func (l *DiskLogger) Clear() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		_ = l.file.Close()
		l.file = nil
	}

	// Remove active and rotated logs
	_ = os.Remove(l.logFile)
	for i := 1; i <= maxLogBackups; i++ {
		_ = os.Remove(filepath.Join(l.logDir, fmt.Sprintf("aethergrok.%d.log", i)))
	}

	return l.openFile()
}
