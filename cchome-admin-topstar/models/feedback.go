package models

import (
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
)

type Feedback struct {
	ID        uuid.ID `gorm:"column:id"`
	UID       uuid.ID `gorm:"column:uid;index:i_u" json:"uid"`
	Content   string  `gorm:"column:content;type:text;" json:"content"`
	IsProcess bool    `gorm:"column:is_process" json:"is_process"`
	Remark    string  `gorm:"column:remark;type:text;" json:"remark"`
	Email     string  `gorm:"column:email;type:char(50);" json:"email"`

	gormv2.Base
}

func (e Feedback) DBName() string {
	return "cchome-admin"
}

func (e Feedback) TableName() string {
	return "feedbacks"
}
