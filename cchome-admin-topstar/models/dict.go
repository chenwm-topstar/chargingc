package models

import (
	"context"
	"encoding/json"

	"github.com/chenwm-topstar/chargingc/utils/gormv2"
)

type KindDictType int

const (
	KindDictTypeEmail                 KindDictType = 1
	KindDictTypeAbout                 KindDictType = 2
	KindDictTypeLatestFirmwareVersion KindDictType = 3
)

type Dict struct {
	ID  KindDictType `gorm:"column:id;primary_key" `
	Val string       `gorm:"column:val;type:text;" `

	gormv2.Base
}

func (e Dict) DBName() string {
	return "cchome-admin"
}

func (e Dict) TableName() string {
	return "dicts"
}

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
