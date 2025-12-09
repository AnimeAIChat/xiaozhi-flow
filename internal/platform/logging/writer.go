package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	LogRetentionDays = 7
)

// RotatableFileWriter 支持按日期轮转的文件写入器
type RotatableFileWriter struct {
	dir         string
	filename    string
	file        *os.File
	currentDate string
	mu          sync.Mutex
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

func NewRotatableFileWriter(dir, filename string) (*RotatableFileWriter, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %v", err)
	}

	w := &RotatableFileWriter{
		dir:      dir,
		filename: filename,
		stopCh:   make(chan struct{}),
	}

	if err := w.openFile(); err != nil {
		return nil, err
	}

	w.startRotationChecker()
	return w, nil
}

func (w *RotatableFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return 0, os.ErrClosed
	}
	return w.file.Write(p)
}

func (w *RotatableFileWriter) Close() error {
	close(w.stopCh)
	w.wg.Wait()

	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *RotatableFileWriter) openFile() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	logPath := filepath.Join(w.dir, w.filename)
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	w.file = file
	w.currentDate = time.Now().Format("2006-01-02")
	return nil
}

func (w *RotatableFileWriter) startRotationChecker() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.checkAndRotate()
			case <-w.stopCh:
				return
			}
		}
	}()
}

func (w *RotatableFileWriter) checkAndRotate() {
	today := time.Now().Format("2006-01-02")
	w.mu.Lock()
	shouldRotate := today != w.currentDate
	w.mu.Unlock()

	if shouldRotate {
		w.rotate(today)
		w.cleanOldLogs()
	}
}

func (w *RotatableFileWriter) rotate(newDate string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 关闭当前文件
	if w.file != nil {
		w.file.Close()
	}

	// 重命名旧文件
	logPath := filepath.Join(w.dir, w.filename)
	baseName := strings.TrimSuffix(w.filename, filepath.Ext(w.filename))
	ext := filepath.Ext(w.filename)
	// 使用旧日期归档
	archiveName := fmt.Sprintf("%s-%s%s", baseName, w.currentDate, ext)
	archivePath := filepath.Join(w.dir, archiveName)

	if _, err := os.Stat(logPath); err == nil {
		os.Rename(logPath, archivePath)
	}

	// 打开新文件
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		// 如果打开失败，尝试输出到 stderr，但不要 panic
		fmt.Fprintf(os.Stderr, "Failed to open new log file: %v\n", err)
		w.file = nil // 标记为不可用
		return
	}

	w.file = file
	w.currentDate = newDate
}

func (w *RotatableFileWriter) cleanOldLogs() {
	cutoffDate := time.Now().AddDate(0, 0, -LogRetentionDays)
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return
	}

	baseName := strings.TrimSuffix(w.filename, filepath.Ext(w.filename))
	ext := filepath.Ext(w.filename)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// 匹配 pattern: baseName-YYYY-MM-DD.ext
		if strings.HasPrefix(name, baseName+"-") && strings.HasSuffix(name, ext) {
			datePart := strings.TrimSuffix(strings.TrimPrefix(name, baseName+"-"), ext)
			fileDate, err := time.Parse("2006-01-02", datePart)
			if err == nil && fileDate.Before(cutoffDate) {
				os.Remove(filepath.Join(w.dir, name))
			}
		}
	}
}
