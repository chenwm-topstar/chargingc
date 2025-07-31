/*
 * Copyright (c) 2018.
 */

package main

import (
	"github.com/astaxie/beego"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/transac"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdinternal/autoupgrade"
	_ "github.com/chenwm-topstar/chargingc/cchome-admin-tsdinternal/template"
	_ "github.com/chenwm-topstar/chargingc/cchome-admin-tsdrouters"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdtransac"
	"github.com/chenwm-topstar/chargingc/utils/flags"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/redigo"
	"github.com/joho/godotenv"
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
