package models

import (
	"context"
	"encoding/json"

	"github.com/chenwm-topstar/chargingc/utils/gormv2"
)

type KindDictType int

const (
	KindDictTypeEmail                 KindDictType = 1 // 邮件配置
	KindDictTypeAbout                 KindDictType = 2 // 关于配置
	KindDictTypeLatestFirmwareVersion KindDictType = 3 // 设备最新版本配置
)

type Dict struct {
	ID  KindDictType `gorm:"column:id;primary_key" ` //ID
	Val string       `gorm:"column:val;type:text;" ` // 配置内容

	gormv2.Base
}

func (e Dict) DBName() string {
	return "cchome-admin"
}

func (e Dict) TableName() string {
	return "dicts"
}

// type EmailConfig struct {
// 	SendHost      string `json:"send_host"`       // 发送服务器服务器. e.g.: smtp.exmail.qq.com
// 	SendPort      int    `json:"send_port"`       // 发送服务器服务器端口. e.g.: 465
// 	UserName      string `json:"user_name"`       // 发送授权账号. 一般是邮箱账号
// 	Passwd        string `json:"passwd"`          // 发送授权账号密码. 一般是邮箱密码
// 	DefSenderMail string `json:"def_sender_mail"` // 默认发送邮件
// }

type AboutConfig struct {
	Content string `json:"content"`
}

func GetDict(ctx context.Context, dt KindDictType) (*Dict, error) {
	ret := &Dict{}
	if err := gormv2.GetByID(context.Background(), ret, uint64(dt)); err != nil {
		return nil, err
	}
	return ret, nil
}

func SetDict(ctx context.Context, dt KindDictType, val interface{}) error {
	buf, err := json.Marshal(val)
	if err != nil {
		return err
	}
	d, err := GetDict(ctx, dt)
	if err != nil {
		return err
	}

	d.ID = dt
	d.Val = string(buf)
	return gormv2.Save(context.Background(), d)
}
