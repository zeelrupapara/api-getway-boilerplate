package monitor

import "sync"

type HealthKey string

const (
	Health_SMTP              HealthKey = "SMTP"
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
	HealthMonitorList[Health_SMTP] = HealthStatus_unknown
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
