package models

import (
	"gitlab.goiot.net/chargingc/utils/gormv2"
	"gitlab.goiot.net/chargingc/utils/uuid"
)

type Feedback struct {
	ID        uuid.ID `gorm:"column:id"`                                // bindid
	UID       uuid.ID `gorm:"column:uid;index:i_u" json:"uid"`          // 用户ID
	Content   string  `gorm:"column:content;type:text;" json:"content"` // 反馈内容
	IsProcess bool    `gorm:"column:is_process" json:"is_process"`      // 是否已经处理
	Remark    string  `gorm:"column:remark;type:text;" json:"remark"`   // 备注
	Email     string  `gorm:"column:email;type:char(50);" json:"email"` // 邮箱

	gormv2.Base
}

func (e Feedback) DBName() string {
	return "cchome-admin"
}

func (e Feedback) TableName() string {
	return "feedbacks"
}
