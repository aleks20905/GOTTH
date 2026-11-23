package dbstore

import (
	"goth/internal/store"

	"gorm.io/gorm"
)

type GroupStore struct {
	db *gorm.DB
}

type NewGroupStoreParams struct {
	DB *gorm.DB
}

func NewGroupStore(params NewGroupStoreParams) *GroupStore {
	return &GroupStore{
		db: params.DB,
	}
}

func (s *GroupStore) ListGroups() ([]store.DeviceGroup, error) {
	var groups []store.DeviceGroup
	err := s.db.Order("id").Find(&groups).Error
	return groups, err
}

func (s *GroupStore) AddGroup(group *store.DeviceGroup) (*store.DeviceGroup, error) {
	if err := s.db.Create(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupStore) DeleteGroup(id uint) error {
	if err := s.db.Model(&store.Device{}).Where("group_id = ?", id).Update("group_id", nil).Error; err != nil {
		return err
	}
	return s.db.Delete(&store.DeviceGroup{}, id).Error
}
