-- ============================================================================
-- GreenLync API Gateway - Comprehensive Monitoring Database Setup
-- ============================================================================
-- This file sets up MySQL for comprehensive monitoring and observability
-- It enables performance schema, creates monitoring users, and sets up tables

-- ============================================================================
-- 1. PERFORMANCE SCHEMA SETUP (For MySQL Monitoring)
-- ============================================================================

-- Enable performance schema and key instruments for monitoring
-- These settings should also be added to MySQL configuration file (my.cnf)
SET GLOBAL performance_schema = ON;

-- Enable statement monitoring
UPDATE performance_schema.setup_instruments 
SET ENABLED = 'YES', TIMED = 'YES' 
WHERE NAME LIKE 'statement/%';

-- Enable transaction monitoring
UPDATE performance_schema.setup_instruments 
SET ENABLED = 'YES', TIMED = 'YES' 
WHERE NAME LIKE 'transaction';

-- Enable connection monitoring
UPDATE performance_schema.setup_instruments 
SET ENABLED = 'YES', TIMED = 'YES' 
WHERE NAME LIKE 'wait/io/socket/%';

-- Enable table I/O monitoring
UPDATE performance_schema.setup_consumers 
SET ENABLED = 'YES' 
WHERE NAME LIKE 'events_statements_%';

UPDATE performance_schema.setup_consumers 
SET ENABLED = 'YES' 
WHERE NAME LIKE 'events_transactions_%';

-- ============================================================================
-- 2. MONITORING USERS SETUP
-- ============================================================================

-- Create monitoring user for Prometheus MySQL Exporter
CREATE USER IF NOT EXISTS 'monitoring'@'%' IDENTIFIED BY 'monitoring_password_2024!' WITH MAX_USER_CONNECTIONS 5;

-- Grant necessary permissions for monitoring
GRANT PROCESS, REPLICATION CLIENT ON *.* TO 'monitoring'@'%';
GRANT SELECT ON performance_schema.* TO 'monitoring'@'%';
GRANT SELECT ON information_schema.* TO 'monitoring'@'%';
GRANT SELECT ON mysql.* TO 'monitoring'@'%';

-- Create application monitoring user (for health checks)
CREATE USER IF NOT EXISTS 'health_checker'@'%' IDENTIFIED BY 'health_check_2024!' WITH MAX_USER_CONNECTIONS 10;
GRANT SELECT ON greenlync.* TO 'health_checker'@'%';
GRANT PROCESS ON *.* TO 'health_checker'@'%';

-- ============================================================================
-- 3. APPLICATION DATABASE SETUP
-- ============================================================================

-- Create main application database
CREATE DATABASE IF NOT EXISTS greenlync CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE greenlync;

-- ============================================================================
-- 4. MONITORING AND METRICS TABLES
-- ============================================================================

-- System health monitoring table
CREATE TABLE IF NOT EXISTS system_health_events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    service_name VARCHAR(100) NOT NULL,
    status ENUM('running', 'stopped', 'error', 'unknown') NOT NULL,
    previous_status ENUM('running', 'stopped', 'error', 'unknown'),
    error_message TEXT,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_service_status (service_name, status),
    INDEX idx_created_at (created_at),
    INDEX idx_service_time (service_name, created_at)
) ENGINE=InnoDB;

-- API metrics and performance tracking
CREATE TABLE IF NOT EXISTS api_metrics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    endpoint VARCHAR(255) NOT NULL,
    method ENUM('GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD') NOT NULL,
    status_code INT NOT NULL,
    response_time_ms INT NOT NULL,
    request_size_bytes INT DEFAULT 0,
    response_size_bytes INT DEFAULT 0,
    user_id VARCHAR(100),
    ip_address VARCHAR(45),
    user_agent TEXT,
    trace_id VARCHAR(100),
    span_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_endpoint_method (endpoint, method),
    INDEX idx_status_code (status_code),
    INDEX idx_response_time (response_time_ms),
    INDEX idx_created_at (created_at),
    INDEX idx_user_metrics (user_id, created_at),
    INDEX idx_trace (trace_id)
) ENGINE=InnoDB;

-- Database performance metrics
CREATE TABLE IF NOT EXISTS database_metrics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(20,4) NOT NULL,
    metric_unit VARCHAR(50),
    tags JSON,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_metric_name (metric_name),
    INDEX idx_recorded_at (recorded_at),
    INDEX idx_metric_time (metric_name, recorded_at)
) ENGINE=InnoDB;

-- Application errors and incidents
CREATE TABLE IF NOT EXISTS error_events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    service VARCHAR(100) NOT NULL,
    error_type VARCHAR(100) NOT NULL,
    error_message TEXT NOT NULL,
    stack_trace TEXT,
    context JSON,
    severity ENUM('low', 'medium', 'high', 'critical') NOT NULL DEFAULT 'medium',
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP NULL,
    resolved_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_service_error (service, error_type),
    INDEX idx_severity (severity),
    INDEX idx_resolved (resolved),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB;

-- Cache performance metrics
CREATE TABLE IF NOT EXISTS cache_metrics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    cache_type ENUM('redis', 'memory', 'other') NOT NULL DEFAULT 'redis',
    operation ENUM('get', 'set', 'delete', 'exists', 'expire') NOT NULL,
    key_pattern VARCHAR(255),
    hit BOOLEAN,
    execution_time_ms DECIMAL(10,3),
    data_size_bytes INT,
    ttl_seconds INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_cache_operation (cache_type, operation),
    INDEX idx_hit_rate (hit),
    INDEX idx_execution_time (execution_time_ms),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB;

-- Business metrics (Cannabis compliance specific)
CREATE TABLE IF NOT EXISTS business_metrics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    metric_category ENUM('compliance', 'operations', 'users', 'performance') NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(20,4) NOT NULL,
    metric_unit VARCHAR(50),
    jurisdiction VARCHAR(10),
    license_number VARCHAR(100),
    user_id VARCHAR(100),
    metadata JSON,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_category_name (metric_category, metric_name),
    INDEX idx_jurisdiction (jurisdiction),
    INDEX idx_license (license_number),
    INDEX idx_recorded_at (recorded_at)
) ENGINE=InnoDB;

-- Audit trail for monitoring configuration changes
CREATE TABLE IF NOT EXISTS monitoring_audit (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    action ENUM('create', 'update', 'delete', 'configure') NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(100),
    old_value JSON,
    new_value JSON,
    changed_by VARCHAR(100) NOT NULL,
    change_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_action_resource (action, resource_type),
    INDEX idx_changed_by (changed_by),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB;

-- ============================================================================
-- 5. MONITORING VIEWS FOR EASY QUERYING
-- ============================================================================

-- Current system health view
CREATE OR REPLACE VIEW current_system_health AS
SELECT 
    service_name,
    status,
    error_message,
    created_at as last_updated,
    JSON_EXTRACT(metadata, '$.version') as service_version,
    JSON_EXTRACT(metadata, '$.uptime') as uptime
FROM system_health_events she1
WHERE she1.created_at = (
    SELECT MAX(she2.created_at) 
    FROM system_health_events she2 
    WHERE she2.service_name = she1.service_name
)
ORDER BY service_name;

-- API performance summary view
CREATE OR REPLACE VIEW api_performance_summary AS
SELECT 
    endpoint,
    method,
    COUNT(*) as request_count,
    AVG(response_time_ms) as avg_response_time,
    MIN(response_time_ms) as min_response_time,
    MAX(response_time_ms) as max_response_time,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY response_time_ms) as p95_response_time,
    COUNT(CASE WHEN status_code >= 400 THEN 1 END) as error_count,
    COUNT(CASE WHEN status_code >= 400 THEN 1 END) / COUNT(*) * 100 as error_rate,
    DATE(created_at) as date
FROM api_metrics 
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 24 HOURS)
GROUP BY endpoint, method, DATE(created_at)
ORDER BY request_count DESC;

-- Error rate by service view
CREATE OR REPLACE VIEW error_rate_by_service AS
SELECT 
    service,
    COUNT(*) as total_errors,
    COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_errors,
    COUNT(CASE WHEN severity = 'high' THEN 1 END) as high_errors,
    COUNT(CASE WHEN resolved = TRUE THEN 1 END) as resolved_errors,
    AVG(CASE WHEN resolved = TRUE THEN TIMESTAMPDIFF(MINUTE, created_at, resolved_at) END) as avg_resolution_time_minutes,
    DATE(created_at) as date
FROM error_events 
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAYS)
GROUP BY service, DATE(created_at)
ORDER BY total_errors DESC;

-- ============================================================================
-- 6. STORED PROCEDURES FOR MONITORING
-- ============================================================================

DELIMITER //

-- Procedure to record system health events
CREATE PROCEDURE RecordSystemHealth(
    IN p_service_name VARCHAR(100),
    IN p_status VARCHAR(20),
    IN p_error_message TEXT,
    IN p_metadata JSON
)
BEGIN
    DECLARE v_previous_status VARCHAR(20);
    
    -- Get the current status
    SELECT status INTO v_previous_status 
    FROM system_health_events 
    WHERE service_name = p_service_name 
    ORDER BY created_at DESC 
    LIMIT 1;
    
    -- Only insert if status changed or if error message is provided
    IF v_previous_status != p_status OR p_error_message IS NOT NULL THEN
        INSERT INTO system_health_events (
            service_name, 
            status, 
            previous_status, 
            error_message, 
            metadata
        ) VALUES (
            p_service_name, 
            p_status, 
            v_previous_status, 
            p_error_message, 
            p_metadata
        );
    END IF;
END //

-- Procedure to clean old monitoring data
CREATE PROCEDURE CleanOldMonitoringData(IN retention_days INT)
BEGIN
    DECLARE v_cutoff_date TIMESTAMP;
    SET v_cutoff_date = DATE_SUB(NOW(), INTERVAL retention_days DAY);
    
    -- Clean old API metrics (keep last 30 days by default)
    DELETE FROM api_metrics WHERE created_at < v_cutoff_date;
    
    -- Clean old system health events (keep last 90 days)
    DELETE FROM system_health_events 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY);
    
    -- Clean old cache metrics (keep last 7 days)
    DELETE FROM cache_metrics 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 7 DAY);
    
    -- Clean old database metrics (keep last 30 days)
    DELETE FROM database_metrics 
    WHERE recorded_at < v_cutoff_date;
    
    -- Don't clean error_events and business_metrics (keep for compliance)
END //

DELIMITER ;

-- ============================================================================
-- 7. TRIGGERS FOR AUTOMATIC MONITORING
-- ============================================================================

-- Trigger to automatically update cache hit rates
DELIMITER //
CREATE TRIGGER after_cache_metrics_insert
    AFTER INSERT ON cache_metrics
    FOR EACH ROW
BEGIN
    -- Update aggregated cache statistics
    INSERT INTO database_metrics (metric_name, metric_value, metric_unit, tags, recorded_at)
    VALUES (
        'cache_hit_rate',
        (
            SELECT (COUNT(CASE WHEN hit = TRUE THEN 1 END) / COUNT(*)) * 100
            FROM cache_metrics 
            WHERE created_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
              AND cache_type = NEW.cache_type
        ),
        'percentage',
        JSON_OBJECT('cache_type', NEW.cache_type, 'period', '1hour'),
        NOW()
    );
END //
DELIMITER ;

-- ============================================================================
-- 8. INITIAL DATA AND CONFIGURATION
-- ============================================================================

-- Insert initial system health status for all services
INSERT IGNORE INTO system_health_events (service_name, status, metadata) VALUES
('api', 'running', JSON_OBJECT('version', '1.0.0', 'startup_time', NOW())),
('database', 'unknown', JSON_OBJECT('version', '8.0')),
('redis', 'unknown', JSON_OBJECT('version', '7.0')),
('nats', 'unknown', JSON_OBJECT('version', '2.10')),
('smtp', 'unknown', JSON_OBJECT('provider', 'gmail'));

-- Grant permissions on new tables to application user
GRANT SELECT, INSERT, UPDATE ON greenlync.system_health_events TO 'greenlync'@'%';
GRANT SELECT, INSERT, UPDATE ON greenlync.api_metrics TO 'greenlync'@'%';
GRANT SELECT, INSERT, UPDATE ON greenlync.database_metrics TO 'greenlync'@'%';
GRANT SELECT, INSERT, UPDATE ON greenlync.error_events TO 'greenlync'@'%';
GRANT SELECT, INSERT, UPDATE ON greenlync.cache_metrics TO 'greenlync'@'%';
GRANT SELECT, INSERT, UPDATE ON greenlync.business_metrics TO 'greenlync'@'%';
GRANT SELECT, INSERT, UPDATE ON greenlync.monitoring_audit TO 'greenlync'@'%';

-- Grant permissions to monitoring user
GRANT SELECT ON greenlync.* TO 'monitoring'@'%';

-- Flush privileges to ensure all changes take effect
FLUSH PRIVILEGES;

-- ============================================================================
-- 9. OPTIMIZATION AND MAINTENANCE
-- ============================================================================

-- Create event scheduler job to clean old data (runs daily at 2 AM)
SET GLOBAL event_scheduler = ON;

CREATE EVENT IF NOT EXISTS clean_monitoring_data
ON SCHEDULE EVERY 1 DAY
STARTS TIMESTAMP(CURRENT_DATE + INTERVAL 1 DAY, '02:00:00')
DO
  CALL CleanOldMonitoringData(30);

-- ============================================================================
-- END OF MONITORING SETUP
-- ============================================================================

SELECT 'Monitoring database setup completed successfully!' as status;