package controllers

import (
	"encoding/json"
	"net/http"

	"gitlab.goiot.net/chargingc/cchome-admin/internal/lib"
	"gitlab.goiot.net/chargingc/cchome-admin/models"
	"gitlab.goiot.net/chargingc/utils/gormv2"
	"gitlab.goiot.net/chargingc/utils/uuid"
)

type QAAController struct {
	Auth
}

func (c *QAAController) Prepare() {
	c.Auth.Prepare()
}

func (c *QAAController) List() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "qaa/list.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/qaa"
		c.Data["config"].(map[string]interface{})["actionname"] = "list"
		c.Data["title"] = "Q&A"
		return
	}

	f := lib.DataTablesRequest{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &f); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	count, err := gormv2.Count(c.HeaderToContext(), &models.QAA{}, "1=1")
	if err != nil {
		c.Error(http.StatusBadRequest, "count user error: "+err.Error())
	}

	var list []models.QAA
	if err := gormv2.GetDB().Order("created_at desc").Offset(f.Offset).Limit(f.Limit).Find(&list).Error; err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}

	c.Total = int(count)
	c.Rows = make([]interface{}, 0)
	for _, l := range list {
		c.Rows = append(c.Rows, map[string]interface{}{
			"id":         l.ID.String(),
			"q":          l.Q,
			"a":          l.A,
			"updated_at": l.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}

func (c *QAAController) Add() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "qaa/add.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/qaa"
		c.Data["config"].(map[string]interface{})["actionname"] = "add"
		c.Data["title"] = "Q&A add"
		return
	}

	var form struct {
		Language string `form:"language"`
		Q        string `form:"q"`
		A        string `form:"a"`
	}
	if err := c.ParseForm(&form); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	} else if form.Q == "" || form.A == "" {
		c.Error(http.StatusBadRequest, "request param is nil")
	}

	cp := &models.QAA{
		Language: form.Language,
		Q:        form.Q,
		A:        form.A,
	}

	if err := gormv2.Save(c.HeaderToContext(), cp); err != nil {
		c.Error(http.StatusInternalServerError, "save error:"+err.Error())
	}

	c.Msg = "add success"
}

func (c *QAAController) getQAA() *models.QAA {
	idstr := c.GetString(":id")
	if idstr == "" {
		c.Error(http.StatusBadRequest, "id is nil")
	}
	id, _ := uuid.ParseID(idstr)

	qaa := &models.QAA{}
	if err := gormv2.GetByID(c.HeaderToContext(), qaa, id.Uint64()); err != nil {
		c.Error(http.StatusBadRequest, "配置参数获取错误:"+err.Error())
	}
	c.Data["qaa"] = qaa
	return qaa
}
func (c *QAAController) Edit() {
	qaa := c.getQAA()
	if !c.Ctx.Input.IsPost() {
		c.TplName = "qaa/edit.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/qaa"
		c.Data["config"].(map[string]interface{})["actionname"] = "edit"
		c.Data["title"] = "Q&A edit"
		return
	}

	var form struct {
		Language string `form:"language"`
		Q        string `form:"q"`
		A        string `form:"a"`
	}
	if err := c.ParseForm(&form); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	} else if form.Q == "" || form.A == "" {
		c.Error(http.StatusBadRequest, "request param is nil")
	}

	qaa.Language = form.Language
	qaa.Q = form.Q
	qaa.A = form.A
	if err := gormv2.Save(c.HeaderToContext(), qaa); err != nil {
		c.Error(http.StatusInternalServerError, "save error:"+err.Error())
	}

	c.Msg = "add success"
}

func (c *QAAController) Del() {
	if !c.Ctx.Input.IsPost() {
		return
	}
	qaa := c.getQAA()

	if err := gormv2.GetDB().Model(qaa).Delete("id=?", qaa.ID).Error; err != nil {
		c.Error(http.StatusInternalServerError, "delete error:"+err.Error())
	}
	c.Msg = "delete success"
}
