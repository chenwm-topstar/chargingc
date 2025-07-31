package models

import (
	"context"

	"github.com/chenwm-topstar/chargingc/pbs/evsepb"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
	"github.com/pkg/errors"
)

type Connector struct {
	ID               uuid.ID               `gorm:"column:id;not null;auto_increment:false;primary_key" `
	EvseID           uuid.ID               `gorm:"column:evse_id;uniqueIndex:u_connectorid;index:i_e_s" `
	CNO              uint8                 `gorm:"column:cno;uniqueIndex:u_connectorid" `
	Desc             string                `gorm:"column:desc;size:50;default:''" `
	CurrentLimit     int16                 `gorm:"column:current_limit;size:2;default:-1"`
	FaultCode        uint16                `gorm:"column:fault_code;size:2;"`
	State            evsepb.ConnectorState `gorm:"column:state;default:0;index:i_e_s;" `
	RecordID         string                `gorm:"column:record_id;default:null;size:32;" `
	Power            uint32                `gorm:"column:power;default:0;"`
	CurrentA         uint32                `gorm:"column:current_a;default:0;"`
	CurrentB         uint32                `gorm:"column:current_b;default:0;"`
	CurrentC         uint32                `gorm:"column:current_c;default:0;"`
	VoltageA         uint32                `gorm:"column:voltage_a;default:0;"`
	VoltageB         uint32                `gorm:"column:voltage_b;default:0;"`
	VoltageC         uint32                `gorm:"column:voltage_c;default:0;"`
	ConsumedElectric uint32                `gorm:"column:consumed_electric;default:0;"`
	ChargingTime     uint16                `gorm:"column:charging_time;default:0;"`

	gormv2.Base
}

func (e Connector) DBName() string {
	return "cchome-admin"
}

func (e Connector) TableName() string {
	return "connectors"
}

func SetConnectorCurrentLimit(evseid uuid.ID, currentLimit int) error {
	if err := gormv2.GetDB().Model(&Connector{}).Where("evse_id=?", evseid).Update("current_limit", currentLimit).Error; err != nil {
		return errors.Wrapf(err, "SetConnectorCurrentLimit[%s]", evseid)
	}
	return nil
}

func GetConnector(evseid uuid.ID) (*Connector, error) {
	key := evseid.String() + ":get:connect"
	v, err, _ := sg.Do(key, func() (interface{}, error) {
		e := &Connector{}
		if err := gormv2.Find(context.Background(), e, "evse_id=? and cno=1", evseid); err != nil {
			return nil, errors.Wrapf(err, "GetConnector[%s]", evseid)
		}
		return e, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*Connector), nil
}
