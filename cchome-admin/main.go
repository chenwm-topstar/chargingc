/*
 * Copyright (c) 2018.
 */

package main

import (
	"github.com/astaxie/beego"
	"github.com/joho/godotenv"
	"gitlab.goiot.net/chargingc/cchome-admin/internal/autoupgrade"
	_ "gitlab.goiot.net/chargingc/cchome-admin/internal/template"
	_ "gitlab.goiot.net/chargingc/cchome-admin/routers"
	"gitlab.goiot.net/chargingc/cchome-admin/transac"
	"gitlab.goiot.net/chargingc/utils/flags"
	"gitlab.goiot.net/chargingc/utils/gormv2"
	"gitlab.goiot.net/chargingc/utils/redigo"
)

var (
	autoMigrate = flags.Bool("autoMigrate", true, "rebuild mysql tables")
)

func main() {
	flags.Parse()
	if err := gormv2.Init(gormv2.WithAutoMigrate(autoMigrate())); err != nil {
		panic(err)
	}
	redigo.Init()

	godotenv.Load()

	beego.BConfig.WebConfig.StaticDir["/assets"] = "static/assets"
	beego.BConfig.RouterCaseSensitive = true
	beego.BConfig.AppName = "cchome-admin"
	go autoupgrade.ListenEvseAutoUpgrade()

	go transac.Run("0.0.0.0:2011")
	beego.Run("0.0.0.0:2010")
}
