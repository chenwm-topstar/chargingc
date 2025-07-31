package controllers

import (
	"github.com/astaxie/beego"
)

type AjaxController struct {
	beego.Controller
}

func (aj *AjaxController) Lang() {
	// aj.CustomAbort(200, "")
	// aj.Abort("200")
	aj.CustomAbort(200, `define({});`)
}
