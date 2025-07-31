package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/redis"
	"github.com/astaxie/beego/validation"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
	"github.com/sirupsen/logrus"

	pHttp "github.com/chenwm-topstar/chargingc/cchome-admin-topstar/internal/http"
	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/internal/log"
	"google.golang.org/grpc/metadata"
)

type Main struct {
	beego.Controller
	Code     int
	Msg      string
	Wait     int
	JsonData map[string]interface{}
	Total    int
	Rows     []interface{}
}

var globalSessions *session.Manager
var bootTime time.Time

func init() {
	sessionConfig := &session.ManagerConfig{
		CookieName:      "sessid",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     2592000,
		Secure:          false,
		CookieLifeTime:  2592000,
		ProviderConfig:  "redis:6379,100,",
	}
	var err error
	globalSessions, err = session.NewManager("redis", sessionConfig)
	if err != nil {
		panic(err)
	}
	go globalSessions.GC()
	bootTime = time.Now()
}

func (c *Main) SetAddtags(url, icon string) {
	c.Data["config"].(map[string]interface{})["addtabs"] = map[string]interface{}{
		"sider_url": url,
		"icon":      icon,
	}
}

func (c *Main) AddBreadCrumb(href, title string) {
	c.Data["breadcrumb"] = append(c.Data["breadcrumb"].([]map[string]string), map[string]string{
		"href":  href,
		"title": title,
	})
}

func (c *Main) TrimedRefererPath() string {
	return ""
}

func (c *Main) Prepare() {
	c.TplExt = "html"
	if c.Ctx.Input.AcceptsJSON() || c.Ctx.Input.IsAjax() {
		c.EnableRender = false
		c.JsonData = make(map[string]interface{})
		c.Wait = 3
	}

	{
		if requestID := c.Ctx.Input.Header("RequestID"); requestID != "" {
			c.Ctx.Input.SetData("requestID", requestID)
		} else {
			c.Ctx.Input.SetData("requestID", time.Now().Unix())
		}
	}

	controller, action := c.GetControllerAndAction()
	controllerName := strings.Replace(strings.ToLower(controller), "controller", "", 1)
	actionName := strings.ToLower(action)
	c.TplName = fmt.Sprintf("%s/%s.html", controllerName, actionName)
	c.Data["AppWebName"] = "Core cms"
	if ok, _ := c.GetBool("dialog", false); ok {
		c.Data["isDialog"] = true
	} else {
		c.Data["isDialog"] = false
	}

	var jsVersion int64
	if beego.BConfig.RunMode == "dev" {
		c.Data["AppWebName"] = "[dev]" + c.Data["AppWebName"].(string)
		jsVersion = time.Now().UnixNano()
	} else {
		jsVersion = bootTime.UnixNano()
	}

	c.Data["config"] = map[string]interface{}{
		"site": map[string]interface{}{
			"name":     "GoIoT",
			"cdnurl":   "",
			"version":  jsVersion,
			"timezone": "Asia/Shanghai",
			"languages": map[string]interface{}{
				"backend":  "zh-cn",
				"frontend": "zh-cn",
			},
			"logo": "/assets/img/ic_512.png",
		},
		"upload": map[string]interface{}{
			"cdnurl":    "",
			"uploadurl": "",
			"bucket":    "local",
			"maxsize":   "10mb",
			"mimetype":  "jpg,png,bmp,jpeg,gif,zip,rar,xls,xlsx",
			"multipart": []string{},
			"multiple":  false,
		},
		"modulename":     "",
		"controllername": controllerName,
		"actionname":     actionName,
		"jsname":         fmt.Sprintf("backend/%s", controllerName),
		"moduleurl":      "",
		"language":       "zh-cn",
		"fastadmin": map[string]interface{}{
			"usercenter":          true,
			"login_captcha":       false,
			"login_failure_retry": true,
			"login_unique":        false,
			"login_background":    "/assets/img/loginbg.jpg",
			"multiplenav":         false,
			"checkupdate":         false,
			"version":             "1.0.0.20180911_beta",
			"api_url":             "",
		},
		"addtabs": map[string]interface{}{
			"sider_url": "",
			"icon":      "",
		},
		"referer":    c.TrimedRefererPath(),
		"__PUBLIC__": "/",
		"__ROOT__":   "/",
		"__CDN__":    "",
	}

}

func (c *Main) GetAbsoluteURL(s string) string {
	return fmt.Sprintf("http://%s/%s", c.Ctx.Request.Host, s)
}

func (c *Main) HeaderToContext(kv ...string) context.Context {
	md := metadata.Pairs(kv...)
	if v, ok := c.Data["requestID"]; ok {
		md.Append("requestID", fmt.Sprintf("%v", v))
	} else {
		md.Append("requestID", uuid.GetID().String())
	}
	return metadata.NewOutgoingContext(context.Background(), md)
}

func (c *Main) Json(key string, value interface{}) {
	c.JsonData[key] = value
}

func (c *Main) Error(code int, errMsg ...string) {
	logger := c.GetLogger()
	logger.Data["uri"] = c.Ctx.Input.URI()
	logger.Data["ip"] = c.Ctx.Input.IP()

	_code := int(code)
	_msg := ""
	if len(errMsg) > 0 {
		_msg = errMsg[0]
	}

	logger.Error(_msg)
	if c.Ctx.Input.AcceptsJSON() || c.Ctx.Input.IsAjax() || !c.EnableRender {
		resp := &pHttp.Resp{
			Code:  _code,
			Msg:   _msg,
			Data:  c.JsonData,
			Wait:  c.Wait,
			Total: c.Total,
			Rows:  c.Rows,
		}
		c.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
		respJson, _ := json.Marshal(resp)
		c.CustomAbort(http.StatusOK, string(respJson))
	} else {
		c.Data["msg"] = _msg
		beego.Exception(uint64(_code), c.Ctx)
		c.CustomAbort(_code, _msg)
		c.StopRun()
	}
}

func (c *Main) Finish() {
	if !c.EnableRender {
		if c.Ctx.Input.AcceptsJSON() || c.Ctx.Input.IsUpload() {
			if c.Data["json"] == nil {
				if c.Msg == "" {
					c.Msg = "done"
				}
				resp := &pHttp.Resp{
					Code:  c.Code,
					Msg:   c.Msg,
					Data:  c.JsonData,
					Wait:  c.Wait,
					Total: c.Total,
					Rows:  c.Rows,
				}
				c.Ctx.Output.JSON(resp, false, true)
			}
			c.ServeJSON()
		}
	}
}

func (c Main) GetLogger() *logrus.Entry {
	return log.FromBeegoContext(c.Ctx)
}

type ValidateRequired struct {
	Obj interface{}
	Key string
}

func (c Main) CheckAndValidRequest(obj interface{}, f func(*validation.Validation) error, required ...ValidateRequired) error {
	valid := &validation.Validation{}
	if len(required) > 0 {
		for _, f := range required {
			valid.Required(f.Obj, f.Key+".required")
		}
		if err := f(valid); err != nil {
			c.Error(http.StatusBadRequest, err.Error())
		} else if valid.HasErrors() {
			c.Error(http.StatusBadRequest, fmt.Sprintf("%v", valid.ErrorsMap))
		}
	}

	b, err := valid.Valid(obj)
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	} else if !b {
		c.Error(http.StatusBadRequest, fmt.Sprintf("%v", valid.ErrorsMap))
	}
	return nil
}
