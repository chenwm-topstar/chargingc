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

	pHttp "github.com/chenwm-topstar/chargingc/cchome-admin/internal/http"
	"github.com/chenwm-topstar/chargingc/cchome-admin/internal/log"
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
	// for _, v := range vals {
	c.Data["config"].(map[string]interface{})["addtabs"] = map[string]interface{}{
		"sider_url": url,
		"icon":      icon,
	}
	// }
}

// AddBreadCrumb 增加面包屑
func (c *Main) AddBreadCrumb(href, title string) {
	c.Data["breadcrumb"] = append(c.Data["breadcrumb"].([]map[string]string), map[string]string{
		"href":  href,
		"title": title,
	})
}

// TrimedRefererPath 截取来路路径，给fastadmin使用
func (c *Main) TrimedRefererPath() string {
	// r := c.Ctx.Request.Referer()
	// if r != "" {
	// 	pr, _ := url.Parse(r)
	// 	requestURL := c.Ctx.Request.URL.String()
	// 	idx := strings.Index(requestURL, "?")
	// 	if idx >= 0 && pr.Path == requestURL[0:strings.Index(requestURL, "?")] {
	// 		return ""
	// 	}
	// 	return pr.Path
	// }
	// return r
	return ""
}

func (c *Main) Prepare() {
	c.TplExt = "html"
	if c.Ctx.Input.AcceptsJSON() || c.Ctx.Input.IsAjax() {
		c.EnableRender = false
		c.JsonData = make(map[string]interface{})
		c.Wait = 3
		//c.Data["json"] = make(map[string]interface{})
	}

	{ //requestID
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
	// c.code = 0
	c.Data["AppWebName"] = "Core 管理后台"
	// c.Data["autoSaves"] = make(map[string]orm.IBase)
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

	// //根据域名找到对应的运营商
	// {
	// 	hostArr := strings.Split(c.Ctx.Request.Host, ".")
	// 	if c.Ctx.Request.Host != "groupadmin:8040" {
	// 		id, err := strconv.Atoi(hostArr[0])
	// 		if err == nil {
	// 			req := &api.GetOperatorByIDReq{
	// 				Id: uint64(id),
	// 			}
	// 			var resp api.Operator
	// 			if err := grpc.Invoke(context.TODO(), api.AdminServiceServer.GetOperatorByID, req, &resp); err != nil {
	// 				c.Error(http.StatusInternalServerError, err.Error())
	// 			}
	// 			c.Data["config"].(map[string]interface{})["site"].(map[string]interface{})["logo"] = resp.LogoUrl
	// 			c.Data["config"].(map[string]interface{})["site"].(map[string]interface{})["name"] = resp.Name
	// 			c.Data["optr"] = resp
	// 			c.Data["title"] = "集团客户 - " + resp.Name
	// 		} else {
	// 			//c.CustomAbort(http.StatusBadRequest, "aaaaaaa")
	// 			c.Error(http.StatusBadRequest, err.Error())
	// 		}
	// 	}
	// }
}

func (c *Main) GetAbsoluteURL(s string) string {
	return fmt.Sprintf("http://%s/%s", c.Ctx.Request.Host, s)
}

//	func (c *Main) AddToAutoSave(obj orm.IBase) {
//		c.Data["autoSaves"].(map[string]orm.IBase)[obj.InstanceName()] = obj
//	}
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

// func (c *Main) GetOpentracingSpan() opentracing.Span {
// 	return c.Ctx.Input.GetData("span").(opentracing.Span)
// }

// func (c *Main) NewOpentracingSpan(name string) opentracing.Span {
// 	parent := c.GetOpentracingSpan()
// 	// c.Ctx.Input.Context
// 	return opentracing.StartSpan(name, opentracing.ChildOf(parent.Context()))
// }

func (c *Main) Error(code int, errMsg ...string) {
	logger := c.GetLogger()
	logger.Data["uri"] = c.Ctx.Input.URI()
	logger.Data["ip"] = c.Ctx.Input.IP()
	// span := c.GetOpentracingSpan().SetTag(string(ext.Error), true)
	// span.LogFields(
	// 	tracinglog.Error(errors.New(fmt.Sprintf("code=%d err=%s", code, errMsg))),
	// )
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
		// c.Ctx.Output.SetStatus(code)
		//c.Data["json"] = resp
		// c.ServeJSON(true)
		respJson, _ := json.Marshal(resp)
		c.CustomAbort(http.StatusOK, string(respJson))
		// if !c.ReturnHttpStatusCodeAsCode {
		// 	c.CustomAbort(http.StatusOK, string(respJson))
		// } else {
		// c.CustomAbort(_code, string(respJson))
		// }
	} else {
		// fmt.Println(errMsg)
		c.Data["msg"] = _msg
		// c.Ctx.Input.SetData("errMsg", errMsg)
		// fmt.Println(errMsg)
		// c.Ctx.WriteString(errMsg)
		// c.CustomAbort(code, errMsg)
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

// type ValidationRequred
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
		// return fmt.Errorf("%v", valid.ErrorsMap)
		c.Error(http.StatusBadRequest, fmt.Sprintf("%v", valid.ErrorsMap))
		// for _, err := range valid.Errors {
		// 	// c.GetLogger().Error(err.Key, err.Message)
		// 	// log.Println(err.Key, err.Message)
		// 	m.Error(http.StatusBadRequest, fmt.Sprintf("%s:%s", err.Key, err.Message))
		// }
	}
	return nil
	// if err := validationCallback(&valid); err != nil {
	// 	m.Error(http.StatusBadRequest, err.Error())
	// }
	// if valid.HasErrors() {
	// 	return fmt.Errorf("%v", valid.ErrorsMap)
	// }
	// // valid.Required(obj.Code, "注册码")
}
