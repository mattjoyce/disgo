package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

// DisgoWriter implements io.Writer to pipe logs to disgo
type DisgoWriter struct {
	Config     string
	ThreadName string
	Tags       string
	Level      string // Log level for conditional tagging
}

// Write satisfies io.Writer interface
func (w *DisgoWriter) Write(p []byte) (n int, err error) {
	cmd := exec.Command("disgo", "--config", w.Config)
	
	if w.ThreadName != "" {
		cmd.Args = append(cmd.Args, "--thread", w.ThreadName)
	}
	
	// Base tags
	tags := w.Tags
	
	// Add level-based tags if applicable
	if w.Level == "ERROR" || w.Level == "FATAL" {
		if tags != "" {
			tags += ","
		}
		tags += "error,critical"
	} else if w.Level == "WARN" || w.Level == "WARNING" {
		if tags != "" {
			tags += ","
		}
		tags += "warning"
	}
	
	if tags != "" {
		cmd.Args = append(cmd.Args, "--tags", tags)
	}
	
	// Create stdin pipe to send log content
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return 0, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	
	// Start the command before writing to stdin
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start disgo: %w", err)
	}
	
	// Write to stdin and close
	n, err = stdin.Write(p)
	stdin.Close()
	if err != nil {
		return n, fmt.Errorf("failed to write to disgo: %w", err)
	}
	
	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		return n, fmt.Errorf("disgo command failed: %w", err)
	}
	
	return n, nil
}

// LevelLogger wraps a logger with level information
type LevelLogger struct {
	InfoLogger  *log.Logger
	WarnLogger  *log.Logger
	ErrorLogger *log.Logger
}

// NewLevelLogger creates a new LevelLogger with stdout and disgo outputs
func NewLevelLogger(appName string) *LevelLogger {
	// Create base disgo writer
	baseDisgoWriter := &DisgoWriter{
		Config:     "default",
		ThreadName: appName + " Logs",
		Tags:       "golang," + appName,
	}
	
	// Create level-specific disgo writers
	infoDisgoWriter := &DisgoWriter{
		Config:     baseDisgoWriter.Config,
		ThreadName: baseDisgoWriter.ThreadName,
		Tags:       baseDisgoWriter.Tags,
		Level:      "INFO",
	}
	
	warnDisgoWriter := &DisgoWriter{
		Config:     baseDisgoWriter.Config,
		ThreadName: baseDisgoWriter.ThreadName,
		Tags:       baseDisgoWriter.Tags,
		Level:      "WARN",
	}
	
	errorDisgoWriter := &DisgoWriter{
		Config:     baseDisgoWriter.Config,
		ThreadName: baseDisgoWriter.ThreadName,
		Tags:       baseDisgoWriter.Tags,
		Level:      "ERROR",
	}
	
	// Create multi-writers for each level
	infoWriter := io.MultiWriter(os.Stdout, infoDisgoWriter)
	warnWriter := io.MultiWriter(os.Stdout, warnDisgoWriter)
	errorWriter := io.MultiWriter(os.Stdout, errorDisgoWriter)
	
	// Create loggers with level prefixes
	infoLogger := log.New(infoWriter, "INFO: ", log.Ldate|log.Ltime)
	warnLogger := log.New(warnWriter, "WARN: ", log.Ldate|log.Ltime)
	errorLogger := log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime)
	
	return &LevelLogger{
		InfoLogger:  infoLogger,
		WarnLogger:  warnLogger,
		ErrorLogger: errorLogger,
	}
}

// Convenience methods
func (l *LevelLogger) Info(format string, v ...interface{}) {
	l.InfoLogger.Printf(format, v...)
}

func (l *LevelLogger) Warn(format string, v ...interface{}) {
	l.WarnLogger.Printf(format, v...)
}

func (l *LevelLogger) Error(format string, v ...interface{}) {
	l.ErrorLogger.Printf(format, v...)
}

// simulateActivity demonstrates the logger with various message types
func simulateActivity(logger *LevelLogger) {
	logger.Info("Starting application simulation")
	
	// Generate some info messages
	for i := 0; i < 3; i++ {
		logger.Info("Processing item %d", i)
		time.Sleep(1 * time.Second)
	}
	
	// Generate a warning
	logger.Warn("Resource usage is high (80%%)")
	time.Sleep(1 * time.Second)
	
	// Generate an error
	logger.Error("Failed to connect to external API")
	time.Sleep(1 * time.Second)
	
	// Simulate error handling
	err := fmt.Errorf("division by zero")
	if err != nil {
		logger.Error("An unexpected error occurred: %v", err)
	}
	
	logger.Info("Simulation complete")
}

func main() {
	// Create a new logger for our application
	logger := NewLevelLogger("ExampleApp")
	
	// Run the simulation
	simulateActivity(logger)
}