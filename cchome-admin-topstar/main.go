/*
 * Copyright (c) 2018.
 */

package main

import (
	"github.com/astaxie/beego"
	_ "github.com/chenwm-topstar/chargingc/cchome-admin-topstar/internal/template"
	_ "github.com/chenwm-topstar/chargingc/cchome-admin-topstar/routers"
	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/transac"
	"github.com/chenwm-topstar/chargingc/utils/flags"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/redigo"
	"github.com/joho/godotenv"
)

func main() {
	flags.Parse()
	if err := gormv2.Init(true); err != nil {
		panic(err)
	}
	redigo.Init()

	godotenv.Load()

	beego.BConfig.WebConfig.StaticDir["/assets"] = "static/assets"
	beego.BConfig.RouterCaseSensitive = true
	beego.BConfig.AppName = "cchome-admin"

	go transac.Run("0.0.0.0:2011")
	beego.Run("0.0.0.0:2010")
}
