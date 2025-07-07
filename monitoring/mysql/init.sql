-- ============================================================================
-- GreenLync API Gateway - Docker Initialization Script
-- ============================================================================
-- This script runs when the MySQL container starts for the first time
-- It sets up basic users and databases for immediate use

-- Create the main application database
CREATE DATABASE IF NOT EXISTS greenlync CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create application user with secure password
CREATE USER IF NOT EXISTS 'greenlync'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON greenlync.* TO 'greenlync'@'%';

-- Create monitoring user for Prometheus MySQL exporter  
CREATE USER IF NOT EXISTS 'exporter'@'%' IDENTIFIED BY 'exporter_password' WITH MAX_USER_CONNECTIONS 5;
GRANT PROCESS, REPLICATION CLIENT, SELECT ON *.* TO 'exporter'@'%';
GRANT SELECT ON performance_schema.* TO 'exporter'@'%';
GRANT SELECT ON information_schema.* TO 'exporter'@'%';

-- Grant privileges for root user from any host (development only - change in production!)
GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' IDENTIFIED BY 'password' WITH GRANT OPTION;

-- Enable performance schema monitoring (if not already enabled in my.cnf)
-- SET GLOBAL performance_schema = ON;

-- ============================================================================
-- Quick Health Check Table for Container Startup
-- ============================================================================
USE greenlync;

-- Simple health check table that can be used immediately
CREATE TABLE IF NOT EXISTS container_health (
    id INT PRIMARY KEY AUTO_INCREMENT,
    service VARCHAR(50) NOT NULL,
    status ENUM('initializing', 'ready', 'healthy', 'unhealthy') NOT NULL DEFAULT 'initializing',
    last_check TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    details JSON,
    
    UNIQUE KEY unique_service (service)
) ENGINE=InnoDB;

-- Insert initial health status
INSERT INTO container_health (service, status, details) VALUES 
('mysql', 'ready', JSON_OBJECT('version', VERSION(), 'init_time', NOW()))
ON DUPLICATE KEY UPDATE 
status = 'ready', 
details = JSON_OBJECT('version', VERSION(), 'restart_time', NOW());

-- Create simple view for health checks
CREATE OR REPLACE VIEW service_health AS
SELECT service, status, last_check, 
       JSON_UNQUOTE(JSON_EXTRACT(details, '$.version')) as version
FROM container_health;

-- Basic monitoring table for immediate use
CREATE TABLE IF NOT EXISTS quick_metrics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(20,4) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_metric_time (metric_name, timestamp)
) ENGINE=InnoDB;

-- Grant permissions to users
GRANT SELECT, INSERT, UPDATE, DELETE ON greenlync.container_health TO 'greenlync'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON greenlync.quick_metrics TO 'greenlync'@'%';
GRANT SELECT ON greenlync.service_health TO 'greenlync'@'%';
GRANT SELECT ON greenlync.container_health TO 'exporter'@'%';
GRANT SELECT ON greenlync.quick_metrics TO 'exporter'@'%';
GRANT SELECT ON greenlync.service_health TO 'exporter'@'%';

-- Flush privileges to ensure changes take effect
FLUSH PRIVILEGES;

SELECT 'MySQL container initialization completed successfully!' as status;   