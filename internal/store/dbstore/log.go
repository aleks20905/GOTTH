package dbstore

import (
	"goth/internal/store"

	"gorm.io/gorm"
)

type LogStore struct {
	db *gorm.DB
}

type NewLogStoreParams struct {
	DB *gorm.DB
}

func NewLogStore(params NewLogStoreParams) *LogStore {
	return &LogStore{
		db: params.DB,
	}
}

func (s *LogStore) ListEntries(limit int) ([]store.LogEntry, error) {
	var entries []store.LogEntry
	err := s.db.Order("id desc").Limit(limit).Find(&entries).Error
	return entries, err
}

func (s *LogStore) AddEntry(entry *store.LogEntry) error {
	return s.db.Create(entry).Error
}
