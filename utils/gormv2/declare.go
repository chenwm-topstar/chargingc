package gormv2

import (
	"context"
	"database/sql"

	"cchome-admin-topstar/utils/lg"

	"gorm.io/gorm"
)

func IgnoreErrRecordNotFound(err error) error {
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	return err
}

func Begin(opts ...*sql.TxOptions) *gorm.DB {
	return DB.Begin(opts...)
}

func Model(ctx context.Context, value interface{}) (tx *gorm.DB) {
	return FromDBContext(ctx).Model(value)
}

func GetByID(ctx context.Context, obj interface{}, id uint64) error {
	return FromDBContext(ctx).Find(obj, "id = ?", id).Error
}

func Count(ctx context.Context, obj interface{}, dest interface{}, conds ...interface{}) (c int64, err error) {
	if err = FromDBContext(ctx).Model(obj).Where(dest, conds...).Count(&c).Error; err != nil {
		return
	}
	return
}

func First(ctx context.Context, dest interface{}, conds ...interface{}) (err error) {
	return IgnoreErrRecordNotFound(FromDBContext(ctx).First(dest, conds...).Error)
}

func Last(ctx context.Context, dest interface{}, conds ...interface{}) (err error) {
	return IgnoreErrRecordNotFound(FromDBContext(ctx).Last(dest, conds...).Error)
}

func Find(ctx context.Context, dest interface{}, conds ...interface{}) (err error) {
	return IgnoreErrRecordNotFound(FromDBContext(ctx).Find(dest, conds...).Error)
}

func Raw(ctx context.Context, dest interface{}, sql string, values ...interface{}) (err error) {
	return IgnoreErrRecordNotFound(FromDBContext(ctx).Raw(sql, values...).Scan(dest).Error)
}

func RawPluck(ctx context.Context, column string, dest interface{}, sql string, values ...interface{}) (err error) {
	return IgnoreErrRecordNotFound(FromDBContext(ctx).Raw(sql, values...).Pluck(column, dest).Error)
}

func MustFind(ctx context.Context, dest interface{}, conds ...interface{}) (err error) {
	return FromDBContext(ctx).Find(dest, conds...).Error
}

func Save(ctx context.Context, obj interface{}) error {
	return FromDBContext(ctx).Save(obj).Error
}

func IsBeginner(tx *gorm.DB) bool {
	switch tx.Statement.ConnPool.(type) {
	case gorm.TxBeginner, gorm.ConnPoolBeginner:
		return true
	}
	return false
}
func CommitWithErr(ctx context.Context, err error) error {
	tx := FromDBContext(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func Saves(ctx context.Context, objs ...interface{}) (err error) {
	if len(objs) <= 0 {
		lg.Warn("saves object is nil")
		return
	}
	tx := FromDBContext(ctx)

	isBeginner := IsBeginner(tx)
	if !isBeginner {
		tx = tx.Begin()
	}

	defer func() {
		if isBeginner {
			return
		}
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit().Error
	}()
	for _, obj := range objs {
		if err = tx.Model(obj).Save(obj).Error; err != nil {
			break
		}
	}
	return
}

func Updates(ctx context.Context, values interface{}) error {
	return FromDBContext(ctx).Updates(values).Error
}

func Create(ctx context.Context, obj interface{}) error {
	return FromDBContext(ctx).Create(obj).Error
}

func Creates(ctx context.Context, objs ...interface{}) (err error) {
	tx := FromDBContext(ctx)
	isBeginner := IsBeginner(tx)
	if !isBeginner {
		tx = tx.Begin()
	}

	defer func() {
		if isBeginner {
			return
		}

		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit().Error
	}()
	for _, obj := range objs {
		if err = tx.Create(obj).Error; err != nil {
			break
		}
	}
	return
}
