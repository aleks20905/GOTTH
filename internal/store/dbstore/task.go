package dbstore

import (
	"goth/internal/store"

	"gorm.io/gorm"
)

type TaskStore struct {
	db *gorm.DB
}

type NewTaskStoreParams struct {
	DB *gorm.DB
}

func NewTaskStore(params NewTaskStoreParams) *TaskStore {
	return &TaskStore{
		db: params.DB,
	}
}

func (s *TaskStore) ListTasks() ([]store.UpdateTask, error) {
	var tasks []store.UpdateTask
	err := s.db.Order("id desc").Find(&tasks).Error
	return tasks, err
}

func (s *TaskStore) ListTasksByStatus(status string) ([]store.UpdateTask, error) {
	var tasks []store.UpdateTask
	err := s.db.Where("status = ?", status).Order("id").Find(&tasks).Error
	return tasks, err
}

func (s *TaskStore) ListTasksForDevice(deviceID uint) ([]store.UpdateTask, error) {
	var tasks []store.UpdateTask
	err := s.db.Where("device_id = ?", deviceID).Order("id desc").Find(&tasks).Error
	return tasks, err
}

func (s *TaskStore) AddTask(task *store.UpdateTask) (*store.UpdateTask, error) {
	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskStore) SetTaskStatus(id uint, status string) error {
	return s.db.Model(&store.UpdateTask{}).Where("id = ?", id).Update("status", status).Error
}
