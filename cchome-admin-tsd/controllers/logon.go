package controllers

import (
	"net/http"

	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/models"
	"golang.org/x/crypto/bcrypt"
)

type LogonController struct {
	Main
}

type Manager struct {
	Name     string `form:"username" valid:"MaxSize(50)"`             // Name 不能为空并且以 Bee 开头
	Password string `form:"password" valid:"MaxSize(20);MinSize(8);"` // 密码长度
}

func (c *LogonController) Index() {
	defaultUrl := "/index/?ref=addtabs"
	redirectUrl := c.GetString("url", defaultUrl)
	if len("/index/logout") == len(redirectUrl) && redirectUrl[0:len("/index/logout")] == "/index/logout" {
		redirectUrl = defaultUrl
	}
	if c.Ctx.Input.IsPost() {
		from := Manager{}
		if err := c.ParseForm(&from); err != nil {
			c.Error(http.StatusBadRequest, err.Error())
		}

		manager, err := models.GetManagerByName(from.Name)
		if err != nil {
			c.Error(http.StatusBadRequest, err.Error())
		}
		if err := bcrypt.CompareHashAndPassword([]byte(manager.Passwd), []byte(from.Password)); err != nil {
			c.Error(http.StatusBadRequest, "password error")
		}

		//注册session
		sess, err := globalSessions.SessionStart(c.Ctx.ResponseWriter, c.Ctx.Request)
		if err != nil {
			return
		}
		defer sess.SessionRelease(c.Ctx.ResponseWriter)
		sess.Set("uname", manager.Name)

		c.JsonData = map[string]interface{}{
			"url": "/index",
		}
		c.Msg = "login success"
	}
	c.TplName = "logon/index.html"
}
