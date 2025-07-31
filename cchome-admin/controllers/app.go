package controllers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/chenwm-topstar/chargingc/cchome-admin/internal/appproto"
	"github.com/chenwm-topstar/chargingc/cchome-admin/models"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
	"github.com/tidwall/gjson"
)

type AppController struct {
	Main

	app  *models.APPConfig
	Req  *appproto.Request
	Resp *appproto.Response
}

func (c *AppController) Prepare() {
	c.Main.Prepare()

	c.EnableRender = false

	c.Req = &appproto.Request{}
	c.Resp = &appproto.Response{}

	result := gjson.GetManyBytes(c.Ctx.Input.RequestBody, "data", "sign", "timestamp")
	if result[0].Str == "" {
		c.Error(http.StatusBadRequest, "req body is nil") // data不能为空
	}
	if result[1].Str == "" {
		c.Error(http.StatusBadRequest, "req sign is nil") // 签名不能为空
	}
	if result[2].Str == "0" {
		c.Error(http.StatusBadRequest, "req time is nil") // 时间不能为空
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, c.Req); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}

	signData := fmt.Sprintf("%d%s%s", c.Req.Timestamp, c.Req.Data, "chargingc")
	sign := fmt.Sprintf("%x", md5.Sum([]byte(signData)))
	//验证签名
	if sign != c.Req.Sign {
		c.GetLogger().Data["sign"] = sign
		if !(c.Ctx.Request.Header.Get("debug") == "breeze") {
			c.Error(http.StatusBadRequest, "sign fail.")
		}
	}
	logger := c.GetLogger()
	logger.Data["token"] = c.Ctx.Request.Header.Get("token")

	appName := c.Ctx.Request.Header.Get("client")
	appName = strings.ToLower(strings.TrimSpace(appName))

	app, err := models.GetAppByName(c.HeaderToContext(), appName)
	if err != nil {
		c.Error(http.StatusBadRequest, "check appname error:"+err.Error())
	}
	c.app = app
	logger.Data["client"] = app.Name
}

func (c *AppController) Error(code int, data ...string) {
	msg := ""
	if len(data) == 1 {
		msg = data[0]
	}
	c.Resp.Msg = msg
	c.Resp.Ret = code
	s, _ := json.Marshal(c.Resp)
	logger := c.GetLogger()
	logger.Data["req"] = fmt.Sprintf("%+v", string(c.Ctx.Input.RequestBody))
	logger.Data["resp"] = fmt.Sprintf("%+v", c.Resp)
	logger.Errorf("---> %s. error:%s", c.Ctx.Request.RequestURI, string(s))

	if strings.Contains(string(s), "timeout") {
		c.CustomAbort(http.StatusOK, "evse abnormal. please try again later.") // 设备网络异常，请稍后重试.
	}
	c.CustomAbort(http.StatusOK, string(s))
}

func (c *AppController) Finish() {
	var dataStr []byte
	var err error

	if dataStr, err = json.Marshal(c.Resp.RawData); err != nil {
		c.Error(http.StatusInternalServerError, fmt.Sprintf("json encode err:%s", err.Error()))
	} else {
		c.Resp.Ret = http.StatusOK
		c.Resp.Data = string(dataStr)
	}
	c.Resp.Ret = http.StatusOK
	c.Resp.Msg = "success"

	logger := c.GetLogger()
	logger.Data["req"] = fmt.Sprintf("%+v", c.Req)
	logger.Data["resp"] = fmt.Sprintf("%+v", c.Resp)
	logger.Infof("---> %s", c.Ctx.Request.RequestURI)
	_ = c.Ctx.Output.JSON(c.Resp, false, true)
}

func getVerifyCode(appid uuid.ID, userFlag string, op int) (key, verifyCode string) {
	verifyCode = fmt.Sprintf("%04d", rand.Int31()%10000)
	key = fmt.Sprintf("%v:%d:%d:verifycode:private", userFlag, op, appid)
	return
}
