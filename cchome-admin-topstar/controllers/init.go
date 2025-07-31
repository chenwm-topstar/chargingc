package controllers

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

	_ = beego.AddFuncMap("json_encode", func(v interface{}) string {
		b, _ := json.Marshal(v)
		return string(b)
	})

	_ = beego.AddFuncMap("unescaped", func(x string) template.HTML {
		return template.HTML(x)
	})

	_ = beego.AddFuncMap("DeRefToString", func(s interface{}) string {
		if reflect.ValueOf(s).IsNil() {
			return "-"
		}
		return fmt.Sprintf("%v", reflect.ValueOf(s).Elem())
	})
}

type DataTablesRequest struct {
	Order  string `json:"order"`
	Sort   string `json:"sort"`
	Limit  int    `json:"limit"`
	Search string `json:"search"`
	Offset int    `json:"offset"`
}
