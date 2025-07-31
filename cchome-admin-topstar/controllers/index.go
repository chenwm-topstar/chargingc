package controllers

import (
	"net/http"
	"sync"
)

type IndexController struct {
	Auth
	l sync.Mutex
}

func (c *IndexController) Prepare() {
	c.Auth.Prepare()
	c.Data["nav"] = "dashboard"

}

type SystemRuntime struct {
	Name string
	Val  interface{}
}

func (c *IndexController) Index() {
	url := c.GetString("url", "/device/list")

	siderbar, err := c.getManager().GetSiderbar(c.Ctx, url)
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	c.Data["menulist"] = siderbar["menulist"]
	c.Data["navlist"] = siderbar["navlist"]
	c.Data["config"].(map[string]interface{})["referer"] = url
	c.TplName = "index/index.html"
	c.Layout = "layout.html"
}

func (c *IndexController) Logout() {
	c.Layout = ""
	defer c.sess.SessionRelease(c.Ctx.ResponseWriter)

	c.getManager().Release()
	c.sess.Flush()
	c.Data["title"] = "logout success"
	c.Data["target"] = "/logon"
	c.TplName = "components/notice.html"
}
