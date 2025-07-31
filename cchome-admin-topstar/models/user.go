package models

import (
	"context"
	"fmt"

	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
	"github.com/pkg/errors"
)

type UserOAuthType int

const (
	UserOAuthTypeEmail   UserOAuthType = 0 // 0 邮箱登录
	UserOAuthTypePhone   UserOAuthType = 1 // 1 手机号登录
	UserOAuthTypeAppleid UserOAuthType = 2 // 2 appleid登录
	UserOAuthTypeGoogle  UserOAuthType = 3 // 3 google登录
)

type User struct {
	ID        uuid.ID       `gorm:"column:id"`
	APPID     uuid.ID       `gorm:"column:appid;uniqueIndex:u_oa_t_f;uniqueIndex:u_aa;"`
	OAuthType UserOAuthType `gorm:"column:oauth_type;uniqueIndex:u_oa_t_f;" json:"oauth_type"`
	OAuthFlag string        `gorm:"column:oauth_flag;type:char(64);uniqueIndex:u_oa_t_f;" json:"oauth_flag"`
	Account   string        `gorm:"column:account;type:char(64);uniqueIndex:u_aa;" json:"account"`
	Name      string        `gorm:"column:name;type:char(64);" json:"name"`
	Email     string        `gorm:"column:email;type:char(64);" json:"email"`
	Passwd    string        `gorm:"column:passwd;type:char(64);" json:"passwd"`
	IsLogoff  bool          `gorm:"column:is_logoff;type:int(4);" json:"is_logoff"`

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
			return nil, errors.Wrapf(err, "GetUserByID[%s]", id)
		}
		return e, nil
	})
	return v.(*User), err
}
