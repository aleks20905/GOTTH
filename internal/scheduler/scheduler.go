package scheduler

import (
	"goth/internal/store"
	"time"
)

// Updater sends an install command to one agent. agent.Client implements it.
type Updater interface {
	StartUpdate(baseURL string, pkg string, version string) error
}

const staleRunningTimeout = 10 * time.Minute

// Tick runs one scheduler pass: dispatches due tasks, fails stale running ones.
func Tick(deviceStore store.DeviceStore, taskStore store.TaskStore, logStore store.LogStore, updater Updater, now time.Time) {
	pending, err := taskStore.ListTasksByStatus("pending")
	if err != nil {
		return
	}
	for _, task := range pending {
		if task.ScheduledAt.After(now) {
			continue
		}
		Dispatch(deviceStore, taskStore, logStore, updater, &task)
	}

	running, err := taskStore.ListTasksByStatus("running")
	if err != nil {
		return
	}
	for _, task := range running {
		if now.Sub(task.UpdatedAt) > staleRunningTimeout {
			taskStore.SetTaskStatus(task.ID, "failed")
			writeLog(logStore, "error", "updates", "Task timed out.", task.PackageName+" "+task.PackageVersion, task.CreatedBy)
		}
	}
}

// Dispatch sends one task to its target devices and marks it running.
func Dispatch(deviceStore store.DeviceStore, taskStore store.TaskStore, logStore store.LogStore, updater Updater, task *store.UpdateTask) {
	devices, err := targetDevices(deviceStore, task)
	if err != nil || len(devices) == 0 {
		taskStore.SetTaskStatus(task.ID, "failed")
		writeLog(logStore, "error", "updates", "Task has no reachable devices.", task.PackageName+" "+task.PackageVersion, task.CreatedBy)
		return
	}

	sent := false
	for _, device := range devices {
		if err := updater.StartUpdate(device.URL, task.PackageName, task.PackageVersion); err == nil {
			sent = true
		}
	}
	if !sent {
		taskStore.SetTaskStatus(task.ID, "failed")
		writeLog(logStore, "error", "updates", "No device accepted the task.", task.PackageName+" "+task.PackageVersion, task.CreatedBy)
		return
	}

	taskStore.SetTaskStatus(task.ID, "running")
	writeLog(logStore, "info", "updates", "Task started.", task.PackageName+" "+task.PackageVersion, task.CreatedBy)
}

func targetDevices(deviceStore store.DeviceStore, task *store.UpdateTask) ([]store.Device, error) {
	devices, err := deviceStore.ListDevices()
	if err != nil {
		return nil, err
	}

	out := []store.Device{}
	for _, device := range devices {
		if task.DeviceID != nil && device.ID == *task.DeviceID {
			out = append(out, device)
		}
		if task.GroupID != nil && device.GroupID != nil && *device.GroupID == *task.GroupID {
			out = append(out, device)
		}
	}
	return out, nil
}

func writeLog(s store.LogStore, level string, category string, message string, details string, user string) {
	_ = s.AddEntry(&store.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Category:  category,
		Message:   message,
		Details:   details,
		UserEmail: user,
	})
}
