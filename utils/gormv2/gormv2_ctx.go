package gormv2

import (
	"context"

	"gorm.io/gorm"
)

type dbkey string

const (
	dbctxkey dbkey = "dbkey"
)

func NewDBContext(ctx context.Context, db *gorm.DB) context.Context {
	if db == nil {
		db = DB
	}
	return context.WithValue(ctx, dbctxkey, db)
}

func FromDBContext(ctx context.Context) (db *gorm.DB) {
	if tmp := ctx.Value(dbctxkey); tmp != nil {
		return tmp.(*gorm.DB)
	}
	return DB
}
