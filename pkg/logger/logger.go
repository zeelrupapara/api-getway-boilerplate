package logger

import (
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps both logrus and zap loggers for flexibility
type Logger struct {
	Logrus *logrus.Logger
	Zap    *zap.Logger
	Sugar  *zap.SugaredLogger
}

// LogConfig holds logger configuration
type LogConfig struct {
	Level       string
	Format      string
	Output      string
	Structured  bool
	FileOutput  string
	MaxFileSize int
	MaxBackups  int
	MaxAge      int
}

// NewLogger creates a new logger instance
func NewLogger(config *LogConfig) (*Logger, error) {
	// Create logrus logger
	logrusLogger := logrus.New()
	
	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrusLogger.SetLevel(level)

	// Set formatter
	if config.Structured || config.Format == "json" {
		logrusLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		logrusLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Set output
	var output io.Writer = os.Stdout
	if config.Output == "stderr" {
		output = os.Stderr
	}
	
	// Add file output if specified
	if config.FileOutput != "" {
		file, err := os.OpenFile(config.FileOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			output = io.MultiWriter(output, file)
		}
	}
	
	logrusLogger.SetOutput(output)

	// Create zap logger
	zapConfig := zap.NewProductionConfig()
	
	// Set zap level
	switch config.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Configure zap encoding
	if config.Format == "console" || !config.Structured {
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		zapConfig.Encoding = "json"
		zapConfig.EncoderConfig = zap.NewProductionEncoderConfig()
	}

	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Set zap output paths
	zapConfig.OutputPaths = []string{"stdout"}
	if config.FileOutput != "" {
		zapConfig.OutputPaths = append(zapConfig.OutputPaths, config.FileOutput)
	}

	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		Logrus: logrusLogger,
		Zap:    zapLogger,
		Sugar:  zapLogger.Sugar(),
	}, nil
}

// Cannabis-specific logging methods

// LogCannabisAudit logs cannabis-related audit events
func (l *Logger) LogCannabisAudit(userID, action, resource string, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"audit_type":   "cannabis_compliance",
		"user_id":      userID,
		"action":       action,
		"resource":     resource,
		"timestamp":    time.Now().UTC(),
		"compliance":   true,
	}
	
	// Add metadata fields
	for k, v := range metadata {
		fields[k] = v
	}
	
	l.Logrus.WithFields(fields).Info("Cannabis audit event")
}

// LogAgeVerification logs age verification events
func (l *Logger) LogAgeVerification(userID string, verified bool, method string, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"audit_type":      "age_verification",
		"user_id":         userID,
		"age_verified":    verified,
		"verification_method": method,
		"timestamp":       time.Now().UTC(),
		"compliance":      true,
	}
	
	for k, v := range metadata {
		fields[k] = v
	}
	
	l.Logrus.WithFields(fields).Info("Age verification event")
}

// LogStateCompliance logs state compliance checks
func (l *Logger) LogStateCompliance(userID, state string, compliant bool, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"audit_type":    "state_compliance",
		"user_id":       userID,
		"state":         state,
		"compliant":     compliant,
		"timestamp":     time.Now().UTC(),
		"compliance":    true,
	}
	
	for k, v := range metadata {
		fields[k] = v
	}
	
	l.Logrus.WithFields(fields).Info("State compliance check")
}

// LogSessionActivity logs session-related activities
func (l *Logger) LogSessionActivity(sessionID, userID, action string, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"audit_type":  "session_activity",
		"session_id":  sessionID,
		"user_id":     userID,
		"action":      action,
		"timestamp":   time.Now().UTC(),
	}
	
	for k, v := range metadata {
		fields[k] = v
	}
	
	l.Logrus.WithFields(fields).Info("Session activity")
}

// LogWebSocketEvent logs WebSocket events
func (l *Logger) LogWebSocketEvent(sessionID, userID, eventType string, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"event_type":  "websocket",
		"session_id":  sessionID,
		"user_id":     userID,
		"ws_event":    eventType,
		"timestamp":   time.Now().UTC(),
	}
	
	for k, v := range metadata {
		fields[k] = v
	}
	
	l.Logrus.WithFields(fields).Info("WebSocket event")
}

// LogHTTPRequest logs HTTP requests with cannabis compliance context
func (l *Logger) LogHTTPRequest(method, path, userID, sessionID string, statusCode int, duration time.Duration, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"event_type":   "http_request",
		"method":       method,
		"path":         path,
		"user_id":      userID,
		"session_id":   sessionID,
		"status_code":  statusCode,
		"duration_ms":  duration.Milliseconds(),
		"timestamp":    time.Now().UTC(),
	}
	
	for k, v := range metadata {
		fields[k] = v
	}
	
	l.Logrus.WithFields(fields).Info("HTTP request")
}

// Standard logging methods

// Debug logs debug messages
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.Sugar.Debugw(msg, fields...)
}

// Info logs info messages
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.Sugar.Infow(msg, fields...)
}

// Warn logs warning messages
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.Sugar.Warnw(msg, fields...)
}

// Error logs error messages
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.Sugar.Errorw(msg, fields...)
}

// Fatal logs fatal messages and exits
func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.Sugar.Fatalw(msg, fields...)
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	logrusFields := logrus.Fields{}
	for k, v := range fields {
		logrusFields[k] = v
	}
	return l.Logrus.WithFields(logrusFields)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Zap.Sync()
}