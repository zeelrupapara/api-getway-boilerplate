package monitor

import (
	"sync"
	"time"
)

type HealthKey string

const (
	// Current application services
	Health_SMTP     HealthKey = "smtp"
	Health_Database HealthKey = "database"
	Health_Redis    HealthKey = "redis"
	Health_NATS     HealthKey = "nats"
	Health_Cache    HealthKey = "cache"
	Health_API      HealthKey = "api"
	
	// Legacy services (for backward compatibility)
	Health_MarketFeed        HealthKey = "market_feed"
	Health_TraderConfig      HealthKey = "trader_config"
	Health_SymbolsParamaters HealthKey = "symbols_parameters"
	Health_RSS               HealthKey = "RSS"
)

type HealthStatus int32

const (
	HealthStatus_running HealthStatus = 0
	HealthStatus_stopped HealthStatus = 1
	HealthStatus_error   HealthStatus = 2
	HealthStatus_unknown HealthStatus = 3
)

// Enum value maps for HealthStatus.
var (
	HealthStatus_name = map[HealthStatus]string{
		0: "running",
		1: "stopped",
		2: "error",
		3: "unknown",
	}
	HealthStatus_value = map[string]HealthStatus{
		"running": 0,
		"stopped": 1,
		"error":   2,
		"unknown": 3,
	}
)

var (
	HealthMonitorList = make(map[HealthKey]HealthStatus)
	Chan              = make(chan struct{})
	locker            = &sync.RWMutex{}
)

func InitMonitor() {
	// Initialize current services
	HealthMonitorList[Health_SMTP] = HealthStatus_unknown
	HealthMonitorList[Health_Database] = HealthStatus_unknown
	HealthMonitorList[Health_Redis] = HealthStatus_unknown
	HealthMonitorList[Health_NATS] = HealthStatus_unknown
	HealthMonitorList[Health_Cache] = HealthStatus_unknown
	HealthMonitorList[Health_API] = HealthStatus_running
	
	// Initialize legacy services (for backward compatibility)
	HealthMonitorList[Health_MarketFeed] = HealthStatus_unknown
	HealthMonitorList[Health_TraderConfig] = HealthStatus_unknown
	HealthMonitorList[Health_SymbolsParamaters] = HealthStatus_unknown
	HealthMonitorList[Health_RSS] = HealthStatus_unknown
}

type SystemHealth map[HealthKey]string

func GetHealthStatus() SystemHealth {
	sh := make(SystemHealth)
	for k := range HealthMonitorList {
		sh[k] = HealthStatus_name[HealthMonitorList[k]]
	}
	return sh
}

func ChangeStatus(h HealthKey, val HealthStatus) {
	locker.Lock()
	defer locker.Unlock()

	if _, ok := HealthMonitorList[h]; ok {
		HealthMonitorList[h] = val
	}
}

// DetailedHealthStatus provides more comprehensive health information
type DetailedHealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Services  map[string]ServiceInfo `json:"services"`
	Uptime    string                 `json:"uptime,omitempty"`
	Version   string                 `json:"version,omitempty"`
}

type ServiceInfo struct {
	Status      string    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
	Error       string    `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

var (
	serviceDetails = make(map[HealthKey]*ServiceInfo)
	startTime      = time.Now()
)

// GetDetailedHealth returns comprehensive health information
func GetDetailedHealth() DetailedHealthStatus {
	locker.RLock()
	defer locker.RUnlock()

	services := make(map[string]ServiceInfo)
	overallHealthy := true

	for key, status := range HealthMonitorList {
		statusStr := HealthStatus_name[status]
		if status == HealthStatus_error || status == HealthStatus_stopped {
			overallHealthy = false
		}

		info := ServiceInfo{
			Status:      statusStr,
			LastChecked: time.Now(),
		}

		if detail, exists := serviceDetails[key]; exists {
			info = *detail
		}

		services[string(key)] = info
	}

	overallStatus := "healthy"
	if !overallHealthy {
		overallStatus = "unhealthy"
	}

	return DetailedHealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Services:  services,
		Uptime:    time.Since(startTime).String(),
		Version:   "1.0.0", // This should come from build info
	}
}

// UpdateServiceDetail updates detailed information for a service
func UpdateServiceDetail(key HealthKey, status HealthStatus, err error, metadata map[string]interface{}) {
	locker.Lock()
	defer locker.Unlock()

	info := &ServiceInfo{
		Status:      HealthStatus_name[status],
		LastChecked: time.Now(),
		Metadata:    metadata,
	}

	if err != nil {
		info.Error = err.Error()
	}

	serviceDetails[key] = info
	HealthMonitorList[key] = status
}

// IsHealthy returns true if all critical services are healthy
func IsHealthy() bool {
	locker.RLock()
	defer locker.RUnlock()

	criticalServices := []HealthKey{
		Health_Database,
		Health_Redis,
		Health_NATS,
		Health_API,
	}

	for _, service := range criticalServices {
		if status, exists := HealthMonitorList[service]; exists {
			if status == HealthStatus_error || status == HealthStatus_stopped {
				return false
			}
		}
	}

	return true
}

// GetServiceStatus returns the status of a specific service
func GetServiceStatus(key HealthKey) (HealthStatus, bool) {
	locker.RLock()
	defer locker.RUnlock()

	status, exists := HealthMonitorList[key]
	return status, exists
}
