package controllers

import (
	"net/http"
)

type Error struct {
	Main
}

func (c *Error) Prepare() {
	c.Main.Prepare()
	c.Data["code"] = c.Ctx.Output.Status
	c.Data["desc"] = http.StatusText(c.Data["code"].(int))
	c.Data["isShowGoBack"] = true
	if isDialog, _ := c.GetInt("dialog"); isDialog == 1 {
		c.Data["isShowGoBack"] = false
	}
	c.TplName = "err.html"

}

func (c *Error) Error404() {

}

func (c *Error) Error400() {

}

func (c *Error) Error403() {

}

func (c *Error) Error500() {

}
