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
	Course    uint   `validate:"required,min=1,max=50"`
	Spec      string `validate:"required,min=1,max=50"`
	GroupName string `validate:"required,min=1,max=50"`
}

func NewScheduleStore(params NewScheduleStoreParams) *ScheduleStore {
	return &ScheduleStore{
		db: params.DB,
	}
}

func (s *ScheduleStore) GetSchedule(course uint, spec string, group_name string) (*[]store.Schedule, error) {

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
