package controllers

import (
	"github.com/astaxie/beego"
)

type AjaxController struct {
	beego.Controller
}

func (aj *AjaxController) Lang() {
	aj.CustomAbort(200, `define({});`)
}
