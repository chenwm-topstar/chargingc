/*
 * Copyright (c) 2018.
 */
package controllers

// import (
// 	_ "gitlab.goiot.net/sde-base/golib/web/beego"
// )

import (
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"time"

	"github.com/astaxie/beego"
	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

var (
	sg     singleflight.Group
	ccache *cache.Cache
)

func init() {
	ccache = cache.New(1*time.Minute, 5*time.Minute)

	//json_encode
	_ = beego.AddFuncMap("json_encode", func(v interface{}) string {
		b, _ := json.Marshal(v)
		return string(b)
	})

	// 在模板对象t中注册unescaped
	_ = beego.AddFuncMap("unescaped", func(x string) template.HTML {
		return template.HTML(x)
	})

	_ = beego.AddFuncMap("DeRefToString", func(s interface{}) string {
		if reflect.ValueOf(s).IsNil() {
			return "-"
		}
		return fmt.Sprintf("%v", reflect.ValueOf(s).Elem())
	})
	_ = beego.AddFuncMap("Round", func(s interface{}, i ...int) string {
		f := 2
		if len(i) > 0 {
			f = i[0]
		}
		return fmt.Sprintf(fmt.Sprintf("%%.%df", f), s)
		// return ""
	})
	_ = beego.AddFuncMap("addone", func(i int) int {
		return i + 1
	})
	_ = beego.AddFuncMap("timestamp2string", func(timestamp int64) string {
		if timestamp == 0 {
			return "-"
		}
		return time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
	})

}

type DataTablesRequest struct {
	Order  string `json:"order"`
	Sort   string `json:"sort"`
	Limit  int    `json:"limit"`
	Search string `json:"search"`
	Offset int    `json:"offset"`
}
