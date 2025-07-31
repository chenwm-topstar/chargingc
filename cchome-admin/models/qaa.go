package models

import (
	"gitlab.goiot.net/chargingc/utils/gormv2"
	"gitlab.goiot.net/chargingc/utils/uuid"
)

// Q&A
type QAA struct {
	ID       uuid.ID `gorm:"column:id"`                              //
	Language string  `gorm:"column:language;index;" json:"language"` //
	Q        string  `gorm:"column:q;type:text;" json:"q"`           //
	A        string  `gorm:"column:a;type:text;" json:"a"`           //

	gormv2.Base
}

func (e QAA) DBName() string {
	return "cchome-admin"
}

func (e QAA) TableName() string {
	return "qaas"
}
