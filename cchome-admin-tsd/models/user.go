package models

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/pkg/errors"
	"gitlab.goiot.net/chargingc/utils/gormv2"
	"gitlab.goiot.net/chargingc/utils/uuid"
)

type UserOAuthType int

const (
	UserOAuthTypeEmail   UserOAuthType = 0 // 0 邮箱登录
	UserOAuthTypePhone   UserOAuthType = 1 // 1 手机号登录
	UserOAuthTypeAppleid UserOAuthType = 2 // 2 appleid登录
	UserOAuthTypeGoogle  UserOAuthType = 3 // 3 google登录
	UserOAuthTypeWX      UserOAuthType = 4 // 4 小程序登录
)
const (
	UserFuncAlertNotify = "AlertNotify"
	UserFuncStopNotify  = "StopNotify"
)

type User struct {
	ID                uuid.ID            `gorm:"column:id"`                                                                    // id 自增
	APPID             uuid.ID            `gorm:"column:appid;uniqueIndex:u_oa_t_f;uniqueIndex:u_aa;"`                          // APPID
	OAuthType         UserOAuthType      `gorm:"column:oauth_type;uniqueIndex:u_oa_t_f;" json:"oauth_type"`                    // 授权类型
	OAuthFlag         string             `gorm:"column:oauth_flag;type:char(64);uniqueIndex:u_oa_t_f;" json:"oauth_flag"`      // 授权标识
	Account           string             `gorm:"column:account;type:char(64);uniqueIndex:u_aa;" json:"account"`                // 账号名称
	Name              string             `gorm:"column:name;type:char(64) COLLATE utf8mb4_unicode_ci;" json:"name"`            // 用户名
	Email             string             `gorm:"column:email;type:char(64);" json:"email"`                                     // 邮件
	Passwd            string             `gorm:"column:passwd;type:char(64);" json:"passwd"`                                   // 用户密码
	IsLogoff          bool               `gorm:"column:is_logoff;type:int(4);" json:"is_logoff"`                               // 是否已经注销, 废弃
	IsACT             *bool              `gorm:"column:is_act;default:1;uniqueIndex:u_oa_t_f;uniqueIndex:u_aa;" json:"is_act"` // 激活用户
	Zone              int                `gorm:"column:zone;type:int(4);" json:"zone"`                                         // 时区
	ManualCtrlConnect bool               `gorm:"column:manual_ctrl_connect;" json:"manual_ctrl_connect"`                       // 手动控制连接方式
	Price             uint32             `gorm:"column:price;" json:"price"`                                                   // 家用电价
	Funcs             gormv2.KindStrings `gorm:"column:funcs;" json:"funcs"`                                                   // 功能列表
	EX                KindUserEx         `gorm:"column:ex;type:text;" json:"ex"`                                               // 扩展数据

	gormv2.Base
}

func (e User) DBName() string {
	return "cchome-admin"
}

func (e User) TableName() string {
	return "users"
}

func GetUser(idstr string) (*User, error) {
	id, err := uuid.ParseID(idstr)
	if err != nil {
		return nil, err
	}
	return GetUserByID(id.Uint64())
}

func GetUserByID(id uint64) (*User, error) {
	key := fmt.Sprintf("%d:user", id)
	v, err, _ := sg.Do(key, func() (interface{}, error) {
		e := &User{}
		if err := gormv2.GetByID(context.Background(), e, id); err != nil {
			return nil, errors.Wrapf(err, "GetUserByID[%d]", id)
		}
		return e, nil
	})
	return v.(*User), err
}

type KindUserEx struct {
	SessionKey string `json:"session_key"`
}

func (KindUserEx) GormDataType() string {
	return "text"
}
func (c KindUserEx) Value() (driver.Value, error) {
	return gormv2.JsonValue(c)
}

func (c *KindUserEx) Scan(input interface{}) error {
	return gormv2.JsonScan(input, c)
}
