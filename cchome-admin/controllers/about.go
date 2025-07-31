package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/chenwm-topstar/chargingc/cchome-admin/models"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
)

type AboutController struct {
	Auth
}

func (c *AboutController) Prepare() {
	c.Auth.Prepare()

	d := &models.Dict{}
	if err := gormv2.GetByID(c.HeaderToContext(), d, uint64(models.KindDictTypeAbout)); err != nil {
		c.Error(http.StatusBadRequest, "get about error:"+err.Error())
	}
	c.Data["about"] = d.Val
	if d.IsExists() && d.Val != "" {
		ac := &models.AboutConfig{}
		if err := json.Unmarshal([]byte(d.Val), ac); err == nil {
			c.Data["about"] = ac.Content
		}
	}

}

func (c *AboutController) Edit() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "about/edit.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/about"
		c.Data["config"].(map[string]interface{})["actionname"] = "edit"
		c.Data["title"] = "about edit"
		return
	}

	var form struct {
		About string `form:"about"`
	}
	if err := c.ParseForm(&form); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	if form.About == "" {
		c.Error(http.StatusBadRequest, "request param is nil")
	}

	if err := models.SetDict(c.HeaderToContext(), models.KindDictTypeAbout, models.AboutConfig{Content: form.About}); err != nil {
		c.Error(http.StatusInternalServerError, "save error:"+err.Error())
	}

	c.Msg = "change success"
}
