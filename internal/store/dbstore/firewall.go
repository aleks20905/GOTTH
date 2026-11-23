package dbstore

import (
	"goth/internal/store"

	"gorm.io/gorm"
)

type FirewallStore struct {
	db *gorm.DB
}

type NewFirewallStoreParams struct {
	DB *gorm.DB
}

func NewFirewallStore(params NewFirewallStoreParams) *FirewallStore {
	return &FirewallStore{
		db: params.DB,
	}
}

func (s *FirewallStore) ListRules(deviceID uint) ([]store.FirewallRule, error) {
	var rules []store.FirewallRule
	err := s.db.Where("device_id = ?", deviceID).Order("id").Find(&rules).Error
	return rules, err
}

func (s *FirewallStore) AddRule(rule *store.FirewallRule) (*store.FirewallRule, error) {
	if err := s.db.Create(rule).Error; err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *FirewallStore) SetRuleEnabled(id uint, enabled bool) error {
	return s.db.Model(&store.FirewallRule{}).Where("id = ?", id).Update("enabled", enabled).Error
}

func (s *FirewallStore) DeleteRule(id uint) error {
	return s.db.Delete(&store.FirewallRule{}, id).Error
}
