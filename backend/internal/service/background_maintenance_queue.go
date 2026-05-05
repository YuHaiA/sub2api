package service

import "sync"

type BackgroundMaintenanceRunMode string

const (
	BackgroundMaintenanceRunNow    BackgroundMaintenanceRunMode = "run_now"
	BackgroundMaintenanceQueued    BackgroundMaintenanceRunMode = "queued"
	BackgroundMaintenanceDuplicate BackgroundMaintenanceRunMode = "duplicate"
)

type BackgroundMaintenanceTask struct {
	Name string
	Run  func()
}

type backgroundMaintenanceState struct {
	running string
	pending *BackgroundMaintenanceTask
}

var (
	backgroundMaintenanceMu sync.Mutex
	backgroundMaintenance   backgroundMaintenanceState
)

func EnqueueBackgroundMaintenance(task BackgroundMaintenanceTask) (BackgroundMaintenanceRunMode, string) {
	backgroundMaintenanceMu.Lock()
	defer backgroundMaintenanceMu.Unlock()

	if backgroundMaintenance.running == "" {
		backgroundMaintenance.running = task.Name
		go runBackgroundMaintenanceTask(task)
		return BackgroundMaintenanceRunNow, ""
	}

	if backgroundMaintenance.running == task.Name {
		return BackgroundMaintenanceDuplicate, task.Name
	}

	if backgroundMaintenance.pending != nil {
		if backgroundMaintenance.pending.Name == task.Name {
			return BackgroundMaintenanceDuplicate, backgroundMaintenance.pending.Name
		}
		return BackgroundMaintenanceQueued, backgroundMaintenance.pending.Name
	}

	pending := task
	backgroundMaintenance.pending = &pending
	return BackgroundMaintenanceQueued, backgroundMaintenance.running
}

func runBackgroundMaintenanceTask(task BackgroundMaintenanceTask) {
	task.Run()

	backgroundMaintenanceMu.Lock()
	backgroundMaintenance.running = ""
	next := backgroundMaintenance.pending
	backgroundMaintenance.pending = nil
	if next != nil {
		backgroundMaintenance.running = next.Name
		go runBackgroundMaintenanceTask(*next)
	}
	backgroundMaintenanceMu.Unlock()
}
