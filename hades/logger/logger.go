package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger manages logging to files and console
type Logger struct {
	runID      string
	planName   string
	stdoutFile *os.File
	stderrFile *os.File
	stdout     io.Writer // Console output
	stderr     io.Writer // Console output
	mu         sync.Mutex
}

// New creates a new logger for a plan run on a specific host
func New(runID, planName, hostName string, consoleOut, consoleErr io.Writer) (*Logger, error) {
	// Create logs directory structure
	logDir := filepath.Join("logs", runID)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create/open stdout log file with host name (append mode to accumulate all jobs)
	stdoutPath := filepath.Join(logDir, fmt.Sprintf("%s.%s.out.log", planName, hostName))
	stdoutFile, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout log: %w", err)
	}

	// Create/open stderr log file with host name (append mode to accumulate all jobs)
	stderrPath := filepath.Join(logDir, fmt.Sprintf("%s.%s.err.log", planName, hostName))
	stderrFile, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		stdoutFile.Close()
		return nil, fmt.Errorf("failed to create stderr log: %w", err)
	}

	return &Logger{
		runID:      runID,
		planName:   planName,
		stdoutFile: stdoutFile,
		stderrFile: stderrFile,
		stdout:     consoleOut,
		stderr:     consoleErr,
	}, nil
}

// Close closes the log files
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var errs []error

	// Sync before closing to ensure all data is written
	if err := l.stdoutFile.Sync(); err != nil {
		errs = append(errs, err)
	}
	if err := l.stderrFile.Sync(); err != nil {
		errs = append(errs, err)
	}

	if err := l.stdoutFile.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := l.stderrFile.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing log files: %v", errs)
	}
	return nil
}

// Stdout returns a writer that writes to both stdout log and console
func (l *Logger) Stdout() io.Writer {
	return &logWriter{
		logger:  l,
		logFile: l.stdoutFile,
		console: l.stdout,
	}
}

// Stderr returns a writer that writes to both stderr log and console
func (l *Logger) Stderr() io.Writer {
	return &logWriter{
		logger:  l,
		logFile: l.stderrFile,
		console: l.stderr,
	}
}

// WriteJobDelimiter writes a job delimiter to the stdout log
func (l *Logger) WriteJobDelimiter(jobName string, actionType string, actionName string, actionIndex int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Format action with optional name
	actionDesc := fmt.Sprintf("[%d] %s", actionIndex, actionType)
	if actionName != "" {
		actionDesc = fmt.Sprintf("[%d] %s // %s", actionIndex, actionType, actionName)
	}

	delimiter := fmt.Sprintf("\n====================\nJOB: %s, ACTION: %s\nSTARTED: %s\n--------------------\n\n",
		jobName, actionDesc, timestamp)

	if _, err := l.stdoutFile.WriteString(delimiter); err != nil {
		return err
	}
	// Sync to ensure delimiter is written to disk
	return l.stdoutFile.Sync()
}

// logWriter is an io.Writer that writes to both log file and console
type logWriter struct {
	logger  *Logger
	logFile *os.File
	console io.Writer
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.logger.mu.Lock()
	defer w.logger.mu.Unlock()

	// Write to log file only (not to console)
	// Console should only show UI status messages, not command output
	n, err = w.logFile.Write(p)
	if err != nil {
		return 0, fmt.Errorf("file write failed: %w", err)
	}

	// Sync to ensure data is flushed to disk
	if err := w.logFile.Sync(); err != nil {
		return 0, fmt.Errorf("file sync failed: %w", err)
	}

	return n, nil
}
