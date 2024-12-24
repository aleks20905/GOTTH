package dbstore

import (
	"goth/internal/store"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ScheduleStore struct {
	db *gorm.DB
}

type NewScheduleStoreParams struct {
	DB *gorm.DB
}

type valiedRequest struct {
	Course    uint   `validate:"max=2"`
	Spec      string `validate:"max=5"`
	GroupName string `validate:"max=5"`
}

func NewScheduleStore(params NewScheduleStoreParams) *ScheduleStore {
	return &ScheduleStore{
		db: params.DB,
	}
}

func (s *ScheduleStore) GetGroupNames() (*[]store.Schedule, error) {
	var shedule []store.Schedule

	err := s.db.Distinct("group_name").Order("group_name").Find(&shedule).Error

	if err != nil {
		return nil, err
	}
	return &shedule, err
}

func (s *ScheduleStore) GetSpecs() (*[]store.Schedule, error) {
	var shedule []store.Schedule

	err := s.db.Distinct("spec").Order("spec").Find(&shedule).Error

	if err != nil {
		return nil, err
	}
	return &shedule, err
}

func (s *ScheduleStore) GetCourses() (*[]store.Schedule, error) {
	var shedule []store.Schedule

	err := s.db.Distinct("course").Find(&shedule).Error

	if err != nil {
		return nil, err
	}
	return &shedule, err
}

func (s *ScheduleStore) GetAllscheduelsUrl() (*[]store.Schedule, error) {
	var shedule []store.Schedule

	err := s.db.Distinct("course", "spec", "group_name").Order("course").Find(&shedule).Error

	if err != nil {
		return nil, err
	}
	return &shedule, err
}

func (s *ScheduleStore) GetSchedules(course uint, spec string, group_name string) (*[]store.Schedule, error) {

	validate := validator.New()
	req := valiedRequest{
		Course:    course,
		Spec:      spec,
		GroupName: group_name,
	}

	err := validate.Struct(req)
	if err != nil {
		return nil, err // Return validation errors to the user
	}

	var shedule []store.Schedule

	err = s.db.Where("course = ? AND spec = ? AND group_name = ?", course, spec, group_name).Order("start ").Find(&shedule).Error

	if err != nil {
		return nil, err
	}
	return &shedule, err
}
