package controllers

import (
	"net/http"

	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/models"
	"golang.org/x/crypto/bcrypt"
)

type ManagerController struct {
	Auth
}

func (c *ManagerController) Prepare() {
	c.Auth.Prepare()
	if m := c.getManager(); m == nil || m.GetModel().Name != "admin" {
		c.NeedLogon()
	}
}

func (c *ManagerController) Edit() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "manager/edit.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/manager"
		c.Data["config"].(map[string]interface{})["actionname"] = "edit"
		c.Data["title"] = "change passwd"
		c.Data["name"] = c.getManager().GetModel().Name
		return
	}

	var form struct {
		Name   string `form:"name"`
		Passwd string `form:"password"`
	}
	if err := c.ParseForm(&form); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	} else if form.Name == "" || form.Passwd == "" {
		c.Error(http.StatusBadRequest, "account or passwd is nil")
	}
	pwd, _ := bcrypt.GenerateFromPassword([]byte(form.Passwd), bcrypt.DefaultCost)
	if err := models.ChangeManagerPasswd(form.Name, string(pwd)); err != nil {
		c.Error(http.StatusBadRequest, "change passwd error:"+err.Error())
	}

	c.Msg = "success"
}
