package models

import (
	"context"

	"github.com/chenwm-topstar/chargingc/pbs/evsepb"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
	"github.com/pkg/errors"
)

// 设备信息
type Evse struct {
	ID                   uuid.ID             `gorm:"column:id;primary_key;" `
	SN                   string              `gorm:"column:sn;type:char(20);uniqueIndex:u_sn" `
	PN                   string              `gorm:"column:pn;type:char(20);" `
	Vendor               string              `gorm:"column:vendor;type:char(30);" `
	Mac                  string              `gorm:"column:mac;type:char(18);" `
	CNum                 uint8               `gorm:"column:cnum;type:char(18);" `
	State                evsepb.EvseState    `gorm:"column:state;size:2;" `
	FirmwareVersion      string              `gorm:"column:firmware_version;type:char(10);" `
	BTVersion            string              `gorm:"column:bt_version;type:char(10);" `
	LastActivityTime     uint32              `gorm:"column:last_activity_time;default:0" `
	LastDisconnectReason string              `gorm:"column:last_disconn_reason;size:128;" `
	Standard             evsepb.EvseStandard `gorm:"column:standard;size:2;"`
	RatedMinCurrent      int32               `gorm:"column:rated_min_current"`
	RatedMaxCurrent      int32               `gorm:"column:rated_max_current"`
	RatedVoltage         int32               `gorm:"column:rated_voltage"`
	RatedPower           int32               `gorm:"column:rated_power"`
	WorkMode             uint8               `gorm:"column:work_mode"`
	Alias                string              `gorm:"column:alias;type:varchar(100)" `

	gormv2.Base
}

func (e Evse) DBName() string {
	return "cchome-admin"
}

func (e Evse) TableName() string {
	return "evses"
}

func EvseOffine(sn string) error {
	if err := gormv2.GetDB().Model(&Evse{}).Where("sn=?", sn).Update("state", evsepb.EvseState_ES_OFFLINE).Error; err != nil {
		return err
	}
	if err := gormv2.GetDB().Model(&Connector{}).Where("evse_id in (select id from evses where sn=?)", sn).Update("state", evsepb.ConnectorState_CS_Unavailable).Error; err != nil {
		return err
	}
	return nil
}

func GetEvseByID(id uint64) (*Evse, error) {
	e := &Evse{}
	if err := gormv2.GetByID(context.Background(), e, id); err != nil {
		return nil, errors.Wrapf(err, "GetEvseByID[%s]", id)
	}
	return e, nil
}

func GetEvseBySN(sn string) (*Evse, error) {
	key := sn + ":get:evse"
	v, err, _ := sg.Do(key, func() (interface{}, error) {
		e := &Evse{}
		if err := gormv2.Find(context.Background(), e, "sn=?", sn); err != nil {
			return nil, errors.Wrapf(err, "GetEvseBySN[%s]", sn)
		}
		return e, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*Evse), nil
}
