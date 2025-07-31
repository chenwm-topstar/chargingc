package models

import (
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
)

// Q&A
type QAA struct {
	ID uuid.ID `gorm:"column:id"`
	Q  string  `gorm:"column:q;type:text;" json:"q"`
	A  string  `gorm:"column:a;type:text;" json:"a"`

	gormv2.Base
}

func (e QAA) DBName() string {
	return "cchome-admin"
}

func (e QAA) TableName() string {
	return "qaas"
}
