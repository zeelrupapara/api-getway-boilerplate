package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DatabaseMonitor handles database-specific monitoring operations
type DatabaseMonitor struct {
	db *gorm.DB
}

// NewDatabaseMonitor creates a new database monitor instance
func NewDatabaseMonitor(db *gorm.DB) *DatabaseMonitor {
	return &DatabaseMonitor{db: db}
}

// HealthEvent represents a system health event for database storage
type HealthEvent struct {
	ServiceName     string                 `json:"service_name" gorm:"column:service_name"`
	Status          string                 `json:"status" gorm:"column:status"`
	PreviousStatus  *string                `json:"previous_status,omitempty" gorm:"column:previous_status"`
	ErrorMessage    *string                `json:"error_message,omitempty" gorm:"column:error_message"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" gorm:"column:metadata;type:json"`
	CreatedAt       time.Time              `json:"created_at" gorm:"column:created_at"`
}

// TableName returns the table name for HealthEvent
func (HealthEvent) TableName() string {
	return "system_health_events"
}

// APIMetric represents API performance metrics
type APIMetric struct {
	ID                uint   `json:"id" gorm:"primaryKey"`
	Endpoint          string `json:"endpoint" gorm:"column:endpoint"`
	Method            string `json:"method" gorm:"column:method"`
	StatusCode        int    `json:"status_code" gorm:"column:status_code"`
	ResponseTimeMs    int    `json:"response_time_ms" gorm:"column:response_time_ms"`
	RequestSizeBytes  int    `json:"request_size_bytes" gorm:"column:request_size_bytes"`
	ResponseSizeBytes int    `json:"response_size_bytes" gorm:"column:response_size_bytes"`
	UserID            string `json:"user_id,omitempty" gorm:"column:user_id"`
	IPAddress         string `json:"ip_address,omitempty" gorm:"column:ip_address"`
	UserAgent         string `json:"user_agent,omitempty" gorm:"column:user_agent"`
	TraceID           string `json:"trace_id,omitempty" gorm:"column:trace_id"`
	SpanID            string `json:"span_id,omitempty" gorm:"column:span_id"`
	CreatedAt         time.Time `json:"created_at" gorm:"column:created_at"`
}

// TableName returns the table name for APIMetric
func (APIMetric) TableName() string {
	return "api_metrics"
}

// CacheMetric represents cache performance metrics
type CacheMetric struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	CacheType       string    `json:"cache_type" gorm:"column:cache_type"`
	Operation       string    `json:"operation" gorm:"column:operation"`
	KeyPattern      string    `json:"key_pattern,omitempty" gorm:"column:key_pattern"`
	Hit             *bool     `json:"hit,omitempty" gorm:"column:hit"`
	ExecutionTimeMs float64   `json:"execution_time_ms" gorm:"column:execution_time_ms"`
	DataSizeBytes   int       `json:"data_size_bytes,omitempty" gorm:"column:data_size_bytes"`
	TTLSeconds      int       `json:"ttl_seconds,omitempty" gorm:"column:ttl_seconds"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at"`
}

// TableName returns the table name for CacheMetric
func (CacheMetric) TableName() string {
	return "cache_metrics"
}

// ErrorEvent represents application errors
type ErrorEvent struct {
	ID           uint                   `json:"id" gorm:"primaryKey"`
	Service      string                 `json:"service" gorm:"column:service"`
	ErrorType    string                 `json:"error_type" gorm:"column:error_type"`
	ErrorMessage string                 `json:"error_message" gorm:"column:error_message"`
	StackTrace   string                 `json:"stack_trace,omitempty" gorm:"column:stack_trace"`
	Context      map[string]interface{} `json:"context,omitempty" gorm:"column:context;type:json"`
	Severity     string                 `json:"severity" gorm:"column:severity"`
	Resolved     bool                   `json:"resolved" gorm:"column:resolved"`
	ResolvedAt   *time.Time             `json:"resolved_at,omitempty" gorm:"column:resolved_at"`
	ResolvedBy   string                 `json:"resolved_by,omitempty" gorm:"column:resolved_by"`
	CreatedAt    time.Time              `json:"created_at" gorm:"column:created_at"`
}

// TableName returns the table name for ErrorEvent
func (ErrorEvent) TableName() string {
	return "error_events"
}

// RecordHealthEvent records a health status change in the database
func (dm *DatabaseMonitor) RecordHealthEvent(serviceName string, status HealthStatus, err error, metadata map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statusStr := HealthStatus_name[status]
	
	// Get previous status
	var previousEvent HealthEvent
	dm.db.WithContext(ctx).Where("service_name = ?", serviceName).
		Order("created_at DESC").First(&previousEvent)
	
	event := HealthEvent{
		ServiceName: serviceName,
		Status:      statusStr,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
	}
	
	if previousEvent.ServiceName != "" && previousEvent.Status != statusStr {
		event.PreviousStatus = &previousEvent.Status
	}
	
	if err != nil {
		errMsg := err.Error()
		event.ErrorMessage = &errMsg
	}
	
	// Use raw SQL to call the stored procedure if it exists, otherwise insert directly
	var dbErr error
	if metadata != nil {
		metadataJSON, _ := json.Marshal(metadata)
		dbErr = dm.db.WithContext(ctx).Exec(
			"CALL RecordSystemHealth(?, ?, ?, ?)",
			serviceName, statusStr, event.ErrorMessage, string(metadataJSON),
		).Error
	} else {
		dbErr = dm.db.WithContext(ctx).Exec(
			"CALL RecordSystemHealth(?, ?, ?, NULL)",
			serviceName, statusStr, event.ErrorMessage,
		).Error
	}
	
	// If stored procedure doesn't exist, fall back to direct insert
	if dbErr != nil {
		return dm.db.WithContext(ctx).Create(&event).Error
	}
	
	return nil
}

// RecordAPIMetric records API performance metrics
func (dm *DatabaseMonitor) RecordAPIMetric(metric APIMetric) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	metric.CreatedAt = time.Now()
	return dm.db.WithContext(ctx).Create(&metric).Error
}

// RecordCacheMetric records cache performance metrics
func (dm *DatabaseMonitor) RecordCacheMetric(metric CacheMetric) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	metric.CreatedAt = time.Now()
	return dm.db.WithContext(ctx).Create(&metric).Error
}

// RecordError records application errors
func (dm *DatabaseMonitor) RecordError(service, errorType, errorMessage, stackTrace string, 
	errContext map[string]interface{}, severity string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	event := ErrorEvent{
		Service:      service,
		ErrorType:    errorType,
		ErrorMessage: errorMessage,
		StackTrace:   stackTrace,
		Context:      errContext,
		Severity:     severity,
		Resolved:     false,
		CreatedAt:    time.Now(),
	}
	
	return dm.db.WithContext(ctx).Create(&event).Error
}

// GetServiceHealth returns current health status from database
func (dm *DatabaseMonitor) GetServiceHealth() (map[string]ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	var events []HealthEvent
	
	// Get latest health event for each service
	subquery := dm.db.Model(&HealthEvent{}).
		Select("service_name, MAX(created_at) as max_created_at").
		Group("service_name")
	
	err := dm.db.WithContext(ctx).
		Joins("JOIN (?) as latest ON system_health_events.service_name = latest.service_name AND system_health_events.created_at = latest.max_created_at", subquery).
		Find(&events).Error
	
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]ServiceInfo)
	for _, event := range events {
		info := ServiceInfo{
			Status:      event.Status,
			LastChecked: event.CreatedAt,
			Metadata:    event.Metadata,
		}
		
		if event.ErrorMessage != nil {
			info.Error = *event.ErrorMessage
		}
		
		result[event.ServiceName] = info
	}
	
	return result, nil
}

// GetAPIMetrics returns API performance metrics for the last period
func (dm *DatabaseMonitor) GetAPIMetrics(since time.Duration) ([]APIMetric, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	var metrics []APIMetric
	err := dm.db.WithContext(ctx).
		Where("created_at >= ?", time.Now().Add(-since)).
		Order("created_at DESC").
		Limit(1000).
		Find(&metrics).Error
	
	return metrics, err
}

// GetErrorEvents returns recent error events
func (dm *DatabaseMonitor) GetErrorEvents(since time.Duration, severity string) ([]ErrorEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	query := dm.db.WithContext(ctx).Where("created_at >= ?", time.Now().Add(-since))
	
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	
	var events []ErrorEvent
	err := query.Order("created_at DESC").Limit(100).Find(&events).Error
	
	return events, err
}

// GetCacheStats returns cache performance statistics
func (dm *DatabaseMonitor) GetCacheStats(since time.Duration) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	var result struct {
		TotalOps  int64   `gorm:"column:total_ops"`
		HitCount  int64   `gorm:"column:hit_count"`
		HitRate   float64 `gorm:"column:hit_rate"`
		AvgTimeMs float64 `gorm:"column:avg_time_ms"`
	}
	
	err := dm.db.WithContext(ctx).
		Model(&CacheMetric{}).
		Select(`
			COUNT(*) as total_ops,
			COUNT(CASE WHEN hit = true THEN 1 END) as hit_count,
			(COUNT(CASE WHEN hit = true THEN 1 END) * 100.0 / COUNT(*)) as hit_rate,
			AVG(execution_time_ms) as avg_time_ms
		`).
		Where("created_at >= ?", time.Now().Add(-since)).
		Scan(&result).Error
	
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"total_operations": result.TotalOps,
		"cache_hits":       result.HitCount,
		"hit_rate":         result.HitRate,
		"avg_time_ms":      result.AvgTimeMs,
	}, nil
}

// TestDatabaseConnection tests if the database is accessible
func (dm *DatabaseMonitor) TestDatabaseConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	sqlDB, err := dm.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	
	return sqlDB.PingContext(ctx)
}

// UpdateServiceStatus updates both in-memory and database status
func (dm *DatabaseMonitor) UpdateServiceStatus(service HealthKey, status HealthStatus, err error, metadata map[string]interface{}) {
	// Update in-memory status
	UpdateServiceDetail(service, status, err, metadata)
	
	// Record in database (async to avoid blocking)
	go func() {
		if dbErr := dm.RecordHealthEvent(string(service), status, err, metadata); dbErr != nil {
			// Log error but don't fail the health update
			fmt.Printf("Failed to record health event in database: %v\n", dbErr)
		}
	}()
}