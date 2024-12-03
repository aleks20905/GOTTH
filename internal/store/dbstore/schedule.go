package dbstore

import (
	"goth/internal/store"

	"gorm.io/gorm"
)

type ScheduleStore struct {
	db *gorm.DB
}

type NewScheduleStoreParams struct {
	DB *gorm.DB
}

func NewScheduleStore(params NewScheduleStoreParams) *ScheduleStore {
	return &ScheduleStore{
		db: params.DB,
	}
}

func (s *ScheduleStore) GetSchedule(course uint, spec string, group_name string) (*[]store.Schedule, error) {

	var shedule []store.Schedule
	err := s.db.Where("course = ? AND spec = ? AND group_name = ?", course, spec, group_name).Find(&shedule).Error

	if err != nil {
		return nil, err
	}
	return &shedule, err
}
