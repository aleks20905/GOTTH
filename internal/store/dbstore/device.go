package dbstore

import (
	"goth/internal/store"

	"gorm.io/gorm"
)

type DeviceStore struct {
	db *gorm.DB
}

type NewDeviceStoreParams struct {
	DB *gorm.DB
}

func NewDeviceStore(params NewDeviceStoreParams) *DeviceStore {
	return &DeviceStore{
		db: params.DB,
	}
}

func (s *DeviceStore) ListDevices() ([]store.Device, error) {
	var devices []store.Device
	err := s.db.Preload("Group").Order("id").Find(&devices).Error
	return devices, err
}

func (s *DeviceStore) GetDevice(id uint) (*store.Device, error) {
	var device store.Device
	err := s.db.Preload("Group").First(&device, id).Error
	return &device, err
}

func (s *DeviceStore) AddDevice(device *store.Device) (*store.Device, error) {
	if err := s.db.Create(device).Error; err != nil {
		return nil, err
	}
	return device, nil
}

func (s *DeviceStore) AssignGroup(deviceID uint, groupID *uint) error {
	return s.db.Model(&store.Device{}).Where("id = ?", deviceID).Update("group_id", groupID).Error
}

func (s *DeviceStore) UpsertDevice(device *store.Device) (*store.Device, error) {
	var existing store.Device
	err := s.db.Where("url = ?", device.URL).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return s.AddDevice(device)
	}
	if err != nil {
		return nil, err
	}

	existing.Name = device.Name
	existing.Version = device.Version
	existing.OS = device.OS
	existing.Verified = device.Verified
	existing.LastSeen = device.LastSeen
	if err := s.db.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *DeviceStore) DeleteDevice(id uint) error {
	return s.db.Delete(&store.Device{}, id).Error
}
