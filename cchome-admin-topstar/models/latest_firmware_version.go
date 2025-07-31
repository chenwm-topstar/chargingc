package models

import (
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
)

type LatestFirmwareVersion struct {
	ID             uuid.ID `gorm:"column:id;primary_key;" `
	PN             string  `gorm:"column:pn;type:char(20);uniqueIndex:u_pv;" `
	Vendor         string  `gorm:"column:vendor;type:char(30);uniqueIndex:u_pv;" `
	LastVersion    int     `gorm:"column:last_version;"`
	UpgradeAddress string  `gorm:"column:upgrade_address;"`
	UpgradeDesc    string  `gorm:"column:upgrade_desc;"`

	gormv2.Base
}

func (e LatestFirmwareVersion) DBName() string {
	return "cchome-admin"
}

func (e LatestFirmwareVersion) TableName() string {
	return "latest_firmware_vers"
}
