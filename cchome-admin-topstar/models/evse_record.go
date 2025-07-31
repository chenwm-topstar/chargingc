package models

import (
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
)

type EvseRecord struct {
	ID               uuid.ID `gorm:"column:id"`
	UID              uuid.ID `gorm:"column:uid;index:i_u" json:"uid"`
	EvseID           uuid.ID `gorm:"column:evse_id;uniqueIndex:u_e_r;" json:"evse_id"`
	SN               string  `gorm:"column:sn;type:char(32);index:i_sn" json:"sn"`
	AuthID           string  `gorm:"column:auth_id;" json:"auth_id"`
	RecordID         string  `gorm:"column:record_id;type:char(64);uniqueIndex:u_e_r;" json:"record_id"`
	AuthMode         uint8   `gorm:"column:auth_mode;" json:"auth_mode"`
	StartTime        uint32  `gorm:"column:start_time;" json:"start_time"`
	ChargeTime       uint32  `gorm:"column:charge_time;" json:"charge_time"`
	TotalElectricity uint32  `gorm:"column:total_electricity;" json:"total_electricity"`
	StopReason       uint8   `gorm:"column:stop_reason;" json:"stop_reason"`
	FaultCode        uint8   `gorm:"column:fault_code;" json:"fault_code"`

	gormv2.Base
}

func (e EvseRecord) DBName() string {
	return "cchome-admin"
}

func (e EvseRecord) TableName() string {
	return "evse_records"
}
