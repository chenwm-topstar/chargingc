package models

import (
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
)

type EvseRecord struct {
	ID               uuid.ID `gorm:"column:id"`                                                          // bindid
	UID              uuid.ID `gorm:"column:uid;index:i_u" json:"uid"`                                    // 用户ID
	EvseID           uuid.ID `gorm:"column:evse_id;uniqueIndex:u_e_r;" json:"evse_id"`                   // 设备ID
	SN               string  `gorm:"column:sn;type:char(32);index:i_sn" json:"sn"`                       // 设备编号
	AuthID           string  `gorm:"column:auth_id;" json:"auth_id"`                                     // 授权ID
	RecordID         string  `gorm:"column:record_id;type:char(64);uniqueIndex:u_e_r;" json:"record_id"` // 充电流水号
	AuthMode         uint8   `gorm:"column:auth_mode;" json:"auth_mode"`                                 // 充电开始时间
	StartTime        uint32  `gorm:"column:start_time;" json:"start_time"`                               // 充电时长 秒
	ChargeTime       uint32  `gorm:"column:charge_time;" json:"charge_time"`                             // 本次充电时间
	TotalElectricity uint32  `gorm:"column:total_electricity;" json:"total_electricity"`                 // 本次充电电量
	StopReason       uint8   `gorm:"column:stop_reason;" json:"stop_reason"`                             // 充电停止原因
	FaultCode        uint8   `gorm:"column:fault_code;" json:"fault_code"`                               // 故障码

	gormv2.Base
}

func (e EvseRecord) DBName() string {
	return "cchome-admin"
}

func (e EvseRecord) TableName() string {
	return "evse_records"
}
