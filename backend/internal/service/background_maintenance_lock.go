package service

import "sync"

var (
	backgroundMaintenanceMu      sync.Mutex
	backgroundMaintenanceRunning bool
	backgroundMaintenanceTask    string
)

func TryAcquireBackgroundMaintenance(task string) (func(), bool, string) {
	backgroundMaintenanceMu.Lock()
	defer backgroundMaintenanceMu.Unlock()

	if backgroundMaintenanceRunning {
		return nil, false, backgroundMaintenanceTask
	}

	backgroundMaintenanceRunning = true
	backgroundMaintenanceTask = task

	return func() {
		backgroundMaintenanceMu.Lock()
		backgroundMaintenanceRunning = false
		backgroundMaintenanceTask = ""
		backgroundMaintenanceMu.Unlock()
	}, true, ""
}
