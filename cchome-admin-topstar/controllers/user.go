package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/internal/lib"
	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/models"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
)

type UserListController struct {
	Auth
}

func (c *UserListController) Prepare() {
	c.Auth.Prepare()
}

func (c *UserListController) List() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "user/list.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/user"
		c.Data["config"].(map[string]interface{})["actionname"] = "list"
		c.Data["title"] = "user list"
		return
	}

	f := lib.DataTablesRequest{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &f); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	userCount, err := gormv2.Count(c.HeaderToContext(), &models.User{}, "1=1")
	if err != nil {
		c.Error(http.StatusBadRequest, "count user error: "+err.Error())
	}

	var users []models.User
	if err := gormv2.GetDB().Order("created_at desc").Offset(f.Offset).Limit(f.Limit).Find(&users).Error; err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}

	apps := make(map[uuid.ID]*models.APPConfig)

	getapp := func(appid uuid.ID) *models.APPConfig {
		if v, ok := apps[appid]; ok {
			return v
		}
		appConfig := &models.APPConfig{}
		if err := gormv2.GetByID(c.HeaderToContext(), appConfig, appid.Uint64()); err != nil {
			c.Error(http.StatusBadRequest, "get app config error: "+err.Error())
		}
		apps[appid] = appConfig
		return appConfig
	}

	c.Total = int(userCount)
	c.Rows = make([]interface{}, 0)
	for _, user := range users {
		appconf := getapp(user.APPID)
		c.Rows = append(c.Rows, map[string]interface{}{
			"id":         user.ID.String(),
			"app_client": appconf.Name,
			"name":       user.Name,
			"email":      user.Email,
		})
	}
}
