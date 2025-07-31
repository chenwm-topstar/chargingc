package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/singleflight"
	"gopkg.in/gomail.v2"
)

var sg singleflight.Group

var LatestFirmwareVersionConfig = &LatestFirmwareVersion{}
var apps map[string]*APPConfig

func GetAppByName(ctx context.Context, appName string) (*APPConfig, error) {
	if v, ok := apps[appName]; ok {
		return v, nil
	}

	app, err, _ := sg.Do(fmt.Sprintf("%s:app:get", appName), func() (interface{}, error) {
		app := &APPConfig{}
		if err := gormv2.Find(ctx, app, "name=?", appName); err != nil {
			return nil, err
		}
		if app.IsNew() {
			return nil, fmt.Errorf("app %s not support", appName)
		}
		app.mailCH = make(chan *gomail.Message, 1024)
		app.SenderEmailWorker(ctx)
		tmpapps := make(map[string]*APPConfig)
		for k := range apps {
			tmpapps[k] = apps[k]
		}
		tmpapps[appName] = app

		apps = tmpapps
		return app, nil
	})
	if err != nil {
		return nil, err
	}

	return app.(*APPConfig), nil
}

func init() {
	apps = make(map[string]*APPConfig)
	gormv2.RegisterModel(&Manager{}, &Evse{}, &Connector{}, &EvseAutoUpgrade{},
		&User{}, &EvseBind{}, &EvseRecord{}, &Dict{}, &Feedback{}, &QAA{}, &APPConfig{}, &LatestFirmwareVersion{})
	gormv2.RegisterAfterCBS(initAdminMager, ininAppConfig, ininEvseLastVersion)
}

func ininEvseLastVersion(ctx context.Context) (err error) {
	dt, err := GetDict(ctx, KindDictTypeLatestFirmwareVersion)
	if err != nil {
		return err
	}

	if dt.IsExists() {
		err = json.Unmarshal([]byte(dt.Val), LatestFirmwareVersionConfig)
	}

	return err
}

func initAdminMager(ctx context.Context) (err error) {
	m := &Manager{}
	if err = gormv2.Last(context.Background(), m, 1); err != nil {
		return errors.New("check admin marager error: " + err.Error())
	}
	if m.IsNew() {
		pwd, _ := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
		if err := gormv2.Save(context.Background(), &Manager{
			ID:     1,
			Name:   "admin",
			Passwd: string(pwd),
		}); err != nil {
			return errors.New("init admin marager error: " + err.Error())
		}
	}
	return nil
}

func ininAppConfig(ctx context.Context) (err error) {
	{
		app := &APPConfig{}
		if err = gormv2.Last(ctx, app, "name=?", "chargingc"); err != nil {
			return err
		}
		if app.IsNew() {
			app.Name = "chargingc"
			app.LatestVersion = "1.1.14"
			app.ForceUpgrade = false
			app.IOSClientId = "net.goiot.chargingc"
			app.AndroidClientId = "net.goiot.chargingc"
			app.Config = KindConfig{
				SendHost:      "smtp.exmail.qq.com",
				SendPort:      465,
				UserName:      "songshenyang@goiot.net",
				Passwd:        "Breeze!23",
				DefSenderMail: "songshenyang@goiot.net",
			}
			if err = gormv2.Save(ctx, app); err != nil {
				return
			}
		}
	}

	{
		app := &APPConfig{}
		if err = gormv2.Last(ctx, app, "name=?", "chargenius"); err != nil {
			return err
		}
		if app.IsNew() {
			app.Name = "chargenius"
			app.LatestVersion = "1.1.14"
			app.ForceUpgrade = false
			app.IOSClientId = "com.topstar.chargenius"
			app.AndroidClientId = "com.topstar.chargenius"
			app.Config = KindConfig{
				SendHost:      "smtp.exmail.qq.com",
				SendPort:      465,
				UserName:      "songshenyang@goiot.net",
				Passwd:        "Breeze!23",
				DefSenderMail: "songshenyang@goiot.net",
			}
			if err = gormv2.Save(ctx, app); err != nil {
				return
			}
		}
	}

	return err
}
