package models

import (
	"context"
	"fmt"

	"github.com/chenwm-topstar/pbs/commonpb"
	"github.com/chenwm-topstar/utils/gormv2"
	"github.com/chenwm-topstar/utils/redigo"
	"github.com/chenwm-topstar/utils/uuid"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

type KindNetWork uint8

const (
	NetWorkWifi KindNetWork = 0
	NetWork4G   KindNetWork = 1
)

// 设备信息
type Evse struct {
	ID                   uuid.ID               `gorm:"column:id;primary_key;" `                   // ID
	SN                   string                `gorm:"column:sn;type:char(20);uniqueIndex:u_sn" ` // 序列号
	PN                   string                `gorm:"column:pn;type:char(20);" `                 // 型号
	Vendor               string                `gorm:"column:vendor;type:char(30);" `             // vendor
	AndroidMac           string                `gorm:"column:android_mac;type:char(30);" `        // android mac地址，一般是蓝牙使用
	IOSMac               string                `gorm:"column:ios_mac;type:char(50);" `            // ios mac地址，一般是蓝牙使用
	CNum                 uint8                 `gorm:"column:cnum;type:char(18);" `               // 枪头数量
	State                commonpb.EvseState    `gorm:"column:state;size:2;" `                     // 状态
	FirmwareVersion      string                `gorm:"column:firmware_version;type:char(10);" `   // 固件版本号
	BTVersion            string                `gorm:"column:bt_version;type:char(10);" `         // 通讯版本号
	LastActivityTime     uint32                `gorm:"column:last_activity_time;default:0" `      // 最后一次保活时间
	LastDisconnectReason string                `gorm:"column:last_disconn_reason;size:128;" `     // 上次链接断开原因
	Standard             commonpb.EvseStandard `gorm:"column:standard;size:2;"`                   // 设备标准
	RatedMinCurrent      int32                 `gorm:"column:rated_min_current"`                  // 额定电流 (单位: A)
	RatedMaxCurrent      int32                 `gorm:"column:rated_max_current"`                  // 额定电流 (单位: A)
	RatedVoltage         int32                 `gorm:"column:rated_voltage"`                      // 额定电压 (单位: V)
	RatedPower           int32                 `gorm:"column:rated_power"`                        // 额定功率 (单位: W)
	WorkMode             uint8                 `gorm:"column:work_mode"`                          // 当前充电模式
	Alias                string                `gorm:"column:alias;type:varchar(100)" `           // 设备别名
	NetWork              KindNetWork           `gorm:"column:network"`                            // 联网方式 0：WIFI 1：4G
	Rssi                 uint8                 `gorm:"column:rssi;"`                              // 20 4G的信号强度 BIN 1 0-31，越大信号越好
	SIM                  string                `gorm:"column:SIM;size:21;"`                       // 21 4G的SIM卡号 BIN 20 4G的SIM卡号

	gormv2.Base
}

func (e Evse) DBName() string {
	return "cchome-admin"
}

func (e Evse) TableName() string {
	return "evses"
}

func GetUserZone(id uuid.ID) (zone int, err error) {
	user := &User{}
	if err := gormv2.GetDB().Model(user).Select("zone").Where("id=?", id).Last(user).Error; err != nil {
		return 0, err
	}
	return user.Zone, nil
}

func GetEvseZone(id uuid.ID) (zone int, err error) {
	user := &User{}
	if err := gormv2.GetDB().Model(user).Select("zone").Where("id in (select uid from evse_bind where evse_id=?)", id).Last(user).Error; err != nil {
		return 0, err
	}
	return user.Zone, nil
}

func EvseOffine(sn string) error {
	if err := gormv2.GetDB().Model(&Evse{}).Where("sn=?", sn).Update("state", commonpb.EvseState_ES_OFFLINE).Error; err != nil {
		return err
	}
	if err := gormv2.GetDB().Model(&Connector{}).Where("evse_id in (select id from evses where sn=?)", sn).Update("state", commonpb.ConnectorState_CS_Unavailable).Error; err != nil {
		return err
	}
	return nil
}

func GetEvseByID(id uint64) (*Evse, error) {
	e := &Evse{}
	if err := gormv2.GetByID(context.Background(), e, id); err != nil {
		return nil, errors.Wrapf(err, "GetEvseByID[%d]", id)
	}
	return e, nil
}

func GetEvseBySN(sn string) (*Evse, error) {
	// todo: 缓存机制
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

func SetAuthCode(id uint64, code string, second int) error {
	key := fmt.Sprintf("%d:auth:evse:cchome", id)
	_, err := redigo.Do("set", key, code, "ex", second)
	return errors.Wrapf(err, "save authcode error")
}

func GetAuthCode(id uint64) (authcode string, ttl int, err error) {
	key := fmt.Sprintf("%d:auth:evse:cchome", id)
	authcode, err = redis.String(redigo.Do("get", key))
	if err != nil {
		return "", 0, errors.Wrapf(err, "get authcode error")
	}
	ttl, err = redis.Int(redigo.Do("ttl", key))
	if err != nil {
		return "", 0, errors.Wrapf(err, "get authcode ttl error")
	}
	return
}
