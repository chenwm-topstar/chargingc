package models

import (
	"context"

	"gitlab.goiot.net/chargingc/utils/gormv2"
	"gitlab.goiot.net/chargingc/utils/uuid"
)

// Manager 管理员
type Manager struct {
	ID     uuid.ID `gorm:"column:id;" json:"id"`
	Name   string  `gorm:"column:name;type:char(20);uniqueIndex" json:"name"`
	Passwd string  `gorm:"column:passwd;type:char(64);" json:"passwd"`

	gormv2.Base
}

func (e Manager) DBName() string {
	return "cchome-admin"
}

func (e Manager) TableName() string {
	return "manages"
}

func GetManagerByName(name string) (ret *Manager, err error) {
	ret = &Manager{}
	if err = gormv2.MustFind(context.Background(), &ret, "name=?", name); err != nil {
		return
	}
	return
}

func ChangeManagerPasswd(name, passwd string) (err error) {
	if err = gormv2.GetDB().Model(&Manager{}).Where("name=?", name).Update("passwd", passwd).Error; err != nil {
		return
	}
	return
}
