package models

import (
	"github.com/chenwm-topstar/utils/gormv2"
	"github.com/chenwm-topstar/utils/uuid"
)

type EvseBind struct {
	ID       uuid.ID `gorm:"column:id"`                                                          // bindid
	UID      uuid.ID `gorm:"column:uid;uniqueIndex:u_u_e" json:"uid"`                            // 用户ID
	SN       string  `gorm:"column:sn;type:char(20);uniqueIndex:u_u_e;" json:"sn"`               // 设备SN
	APPID    uuid.ID `gorm:"column:appid;uniqueIndex:u_e_m" json:"appid"`                        // 用户ID
	EvseID   uuid.ID `gorm:"column:evse_id;uniqueIndex:u_u_e;uniqueIndex:u_e_m;" json:"evse_id"` // 设备ID
	IsMaster *bool   `gorm:"column:is_master;uniqueIndex:u_e_m;" json:"is_master"`               // 设备绑定的第一个用户是拥有者

	gormv2.Base
}

func (e EvseBind) DBName() string {
	return "cchome-admin"
}

func (e EvseBind) TableName() string {
	return "evse_bind"
}
