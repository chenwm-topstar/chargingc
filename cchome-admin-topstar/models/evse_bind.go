package models

import (
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
)

type EvseBind struct {
	ID     uuid.ID `gorm:"column:id"`
	UID    uuid.ID `gorm:"column:uid;uniqueIndex:u_u_e" json:"uid"`
	SN     string  `gorm:"column:sn;type:char(20);uniqueIndex:u_u_e;" json:"sn"`
	EvseID uuid.ID `gorm:"column:evse_id;uniqueIndex:u_u_e;index:i_e;" json:"evse_id"`

	gormv2.Base
}

func (e EvseBind) DBName() string {
	return "cchome-admin"
}

func (e EvseBind) TableName() string {
	return "evse_bind"
}
