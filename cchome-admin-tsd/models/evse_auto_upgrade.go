package models

import "gitlab.goiot.net/chargingc/utils/gormv2"

var OTACH chan EvseAutoUpgrade

func init() {
	OTACH = make(chan EvseAutoUpgrade, 1024)
}

type EvseAutoUpgrade struct {
	ID                     string `gorm:"column:id;autoIncrement;"`
	SN                     string `gorm:"column:sn;index:i_sn;"`
	UpgradeFirmwareVersion int    `gorm:"column:upgrade_firmware_version;"`
	Address                string `gorm:"column:address;"`
	IsUpgrade              bool   `gorm:"column:is_upgrade;"`
	Result                 string `gorm:"column:result;"`

	gormv2.Base
}

func (e EvseAutoUpgrade) DBName() string {
	return "cchome-admin"
}

func (e EvseAutoUpgrade) TableName() string {
	return "evse_auto_upgrades"
}
