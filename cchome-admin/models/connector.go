package models

import (
	"context"

	"github.com/pkg/errors"
	"gitlab.goiot.net/chargingc/pbs/commonpb"
	"gitlab.goiot.net/chargingc/utils/gormv2"
	"gitlab.goiot.net/chargingc/utils/uuid"
)

type Connector struct {
	ID           uuid.ID `gorm:"column:id;not null;auto_increment:false;primary_key" `  //ID
	EvseID       uuid.ID `gorm:"column:evse_id;uniqueIndex:u_connectorid;index:i_e_s" ` //设备ID
	CNO          uint8   `gorm:"column:cno;uniqueIndex:u_connectorid" `                 //枪头编号
	Desc         string  `gorm:"column:desc;size:50;default:''" `                       //枪头描述
	CurrentLimit int16   `gorm:"column:current_limit;size:2;default:-1"`                // 电流限制, 有序充电时使用

	// 枪头实时信息
	FaultCode        uint16                  `gorm:"column:fault_code;size:2;"`               // 故障码
	State            commonpb.ConnectorState `gorm:"column:state;default:0;index:i_e_s;" `    // 枪当前状态
	RecordID         string                  `gorm:"column:record_id;default:null;size:32;" ` // 设备上送的订单号
	Power            uint32                  `gorm:"column:power;default:0;"`                 // 功率, 单位:kW 精度:0.01kW
	CurrentA         uint32                  `gorm:"column:current_a;default:0;"`             // A相电流, 单位:A 精度:0.1A
	CurrentB         uint32                  `gorm:"column:current_b;default:0;"`             // B相电流, 单位:A 精度:0.1A
	CurrentC         uint32                  `gorm:"column:current_c;default:0;"`             // C相电流, 单位:A 精度:0.1A
	VoltageA         uint32                  `gorm:"column:voltage_a;default:0;"`             // A相电压, 单位:V 精度:0.1V
	VoltageB         uint32                  `gorm:"column:voltage_b;default:0;"`             // B相电压, 单位:V 精度:0.1V
	VoltageC         uint32                  `gorm:"column:voltage_c;default:0;"`             // C相电压, 单位:V 精度:0.1V
	ConsumedElectric uint32                  `gorm:"column:consumed_electric;default:0;"`     // 本次充电电量	BIN	4	分辨率：0.001kW·h
	ChargingTime     uint16                  `gorm:"column:charging_time;default:0;"`         // 本次充电时长	BIN	2	分

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
	// todo: 缓存机制
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
