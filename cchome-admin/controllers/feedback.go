package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/chenwm-topstar/chargingc/cchome-admin/internal/lib"
	"github.com/chenwm-topstar/chargingc/cchome-admin/models"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
)

type FeedbackController struct {
	Auth
}

func (c *FeedbackController) Prepare() {
	c.Auth.Prepare()
}

func (c *FeedbackController) List() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "feedback/list.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/feedback"
		c.Data["config"].(map[string]interface{})["actionname"] = "list"
		c.Data["title"] = "feedback"
		return
	}

	f := lib.DataTablesRequest{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &f); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	count, err := gormv2.Count(c.HeaderToContext(), &models.Feedback{}, "1=1")
	if err != nil {
		c.Error(http.StatusBadRequest, "count user error: "+err.Error())
	}

	var list []models.Feedback
	if err := gormv2.GetDB().Order("created_at desc").Offset(f.Offset).Limit(f.Limit).Find(&list).Error; err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}

	c.Total = int(count)
	c.Rows = make([]interface{}, 0)
	for _, l := range list {
		user, err := models.GetUserByID(l.UID.Uint64())
		if err != nil {
			c.Error(http.StatusBadRequest, err.Error())
		}
		c.Rows = append(c.Rows, map[string]interface{}{
			"id":         l.ID.String(),
			"user_name":  user.Name,
			"user_email": l.Email,
			"is_process": l.IsProcess,
			"content":    l.Content,
			"remark":     l.Remark,
			"updated_at": l.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}

func (c *FeedbackController) getFeedback() *models.Feedback {
	idstr := c.GetString(":id")
	if idstr == "" {
		c.Error(http.StatusBadRequest, "id is nil")
	}
	id, _ := uuid.ParseID(idstr)

	feedback := &models.Feedback{}
	if err := gormv2.GetByID(c.HeaderToContext(), feedback, id.Uint64()); err != nil {
		c.Error(http.StatusBadRequest, "配置参数获取错误:"+err.Error())
	}
	c.Data["feedback"] = feedback
	return feedback
}
func (c *FeedbackController) Edit() {
	feedback := c.getFeedback()
	if !c.Ctx.Input.IsPost() {
		c.TplName = "feedback/edit.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/feedback"
		c.Data["config"].(map[string]interface{})["actionname"] = "edit"
		c.Data["title"] = "feedback edit"
		return
	}

	var form struct {
		IsProcess bool   `form:"is_process"`
		Remark    string `form:"remark"`
	}
	if err := c.ParseForm(&form); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	feedback.IsProcess = form.IsProcess
	feedback.Remark = form.Remark
	if err := gormv2.Save(c.HeaderToContext(), feedback); err != nil {
		c.Error(http.StatusInternalServerError, "save error:"+err.Error())
	}

	c.Msg = "edit success"
}
