package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/logging"
	"github.com/sirupsen/logrus"
)

// CloudLogger wraps Google Cloud Logging
type CloudLogger struct {
	client *logging.Client
	logger *logging.Logger
}

// NewCloudLogger creates a new Cloud Logging client
func NewCloudLogger(ctx context.Context, projectID, logName string) (*CloudLogger, error) {
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %w", err)
	}

	logger := client.Logger(logName)

	return &CloudLogger{
		client: client,
		logger: logger,
	}, nil
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Severity string                 `json:"severity"`
	Message  string                 `json:"message"`
	Labels   map[string]string      `json:"labels,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// Log sends a structured log entry to Cloud Logging
func (cl *CloudLogger) Log(entry LogEntry) {
	logEntry := logging.Entry{
		Severity: parseSeverity(entry.Severity),
		Payload:  entry,
		Labels:   entry.Labels,
	}

	cl.logger.Log(logEntry)
}

// Close closes the logging client
func (cl *CloudLogger) Close() error {
	return cl.client.Close()
}

// parseSeverity converts string severity to logging.Severity
func parseSeverity(s string) logging.Severity {
	switch s {
	case "debug":
		return logging.Debug
	case "info":
		return logging.Info
	case "warn", "warning":
		return logging.Warning
	case "error":
		return logging.Error
	case "critical":
		return logging.Critical
	default:
		return logging.Info
	}
}

// CloudLogrusHook is a logrus hook for Cloud Logging
type CloudLogrusHook struct {
	cloudLogger *CloudLogger
}

// NewCloudLogrusHook creates a new logrus hook for Cloud Logging
func NewCloudLogrusHook(cloudLogger *CloudLogger) *CloudLogrusHook {
	return &CloudLogrusHook{
		cloudLogger: cloudLogger,
	}
}

// Fire is called when a log entry is fired
func (hook *CloudLogrusHook) Fire(entry *logrus.Entry) error {
	logEntry := LogEntry{
		Severity: entry.Level.String(),
		Message:  entry.Message,
		Labels: map[string]string{
			"service": "primopoker",
		},
		Data: make(map[string]interface{}),
	}

	// Add fields as data
	for k, v := range entry.Data {
		logEntry.Data[k] = v
	}

	hook.cloudLogger.Log(logEntry)
	return nil
}

// Levels returns the available logging levels
func (hook *CloudLogrusHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
