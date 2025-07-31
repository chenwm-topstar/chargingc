package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/chenwm-topstar/chargingc/cchome-admin/internal/lib"
	"github.com/chenwm-topstar/chargingc/cchome-admin/models"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
)

type DeviceLatestVersionListController struct {
	Auth
}

func (c *DeviceLatestVersionListController) Prepare() {
	c.Auth.Prepare()
}

func (c *DeviceLatestVersionListController) List() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "device_last_ver/list.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/device_last_ver"
		c.Data["config"].(map[string]interface{})["actionname"] = "list"
		c.Data["title"] = "device_last_ver list"
		return
	}

	f := lib.DataTablesRequest{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &f); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	db := gormv2.Model(c.HeaderToContext(), &models.LatestFirmwareVersion{})

	count := int64(0)
	if err := db.Count(&count).Error; err != nil {
		c.Error(http.StatusBadRequest, "count evse error: "+err.Error())
	}

	var verstions []models.LatestFirmwareVersion
	if err := db.Offset(f.Offset).Limit(f.Limit).Find(&verstions).Error; err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}

	c.Total = int(count)
	c.Rows = make([]interface{}, 0)
	for _, v := range verstions {
		c.Rows = append(c.Rows, map[string]interface{}{
			"id":              v.ID.String(),
			"pn":              v.PN,
			"vendor":          v.Vendor,
			"last_version":    v.LastVersion,
			"upgrade_address": v.UpgradeAddress,
			"updated_at":      v.UpdatedAt.Local().Format("2006-01-02 15:04:05"),
			"created_at":      v.CreatedAt.Local().Format("2006-01-02 15:04:05"),
		})
	}
}

func (c *DeviceLatestVersionListController) Add() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "device_last_ver/add.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/device_last_ver"
		c.Data["config"].(map[string]interface{})["actionname"] = "add"
		c.Data["title"] = "evse version"
		return
	}

	var form struct {
		Vendor           string `form:"vendor"`
		Pn               string `form:"pn"`
		LatestOtaVersion int    `form:"latest_ota_version"`
		LatestOtaAddress string `form:"latest_ota_address"`
		LatestOtaDesc    string `form:"latest_ota_desc"`
	}
	if err := c.ParseForm(&form); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	if form.LatestOtaVersion <= 0 || form.LatestOtaAddress == "" {
		c.Error(http.StatusBadRequest, "req address is nil")
	}

	lfv := &models.LatestFirmwareVersion{}
	if err := gormv2.Find(c.HeaderToContext(), lfv, "pn=? and vendor=?", form.Pn, form.Vendor); err != nil {
		c.Error(http.StatusInternalServerError, "check pn error: "+err.Error())
	}
	if lfv.IsExists() {
		c.Error(http.StatusBadRequest, "pn is exists")
	}

	lfv.PN = form.Pn
	lfv.Vendor = form.Vendor
	lfv.LastVersion = form.LatestOtaVersion
	lfv.UpgradeAddress = form.LatestOtaAddress
	lfv.UpgradeDesc = form.LatestOtaDesc
	if err := gormv2.Save(c.HeaderToContext(), lfv); err != nil {
		c.Error(http.StatusInternalServerError, "save error: "+err.Error())
	}

	c.Msg = "request success"
}

type DeviceLatestVersionController struct {
	Auth
}

func (c *DeviceLatestVersionController) Prepare() {
	c.Auth.Prepare()

	idstr := c.GetString(":id")
	if idstr == "" {
		idstr := c.GetString("id")
		if idstr == "" {
			c.Error(http.StatusBadRequest, "id is nil")
		}
	}
	id, _ := uuid.ParseID(idstr)

	lv := &models.LatestFirmwareVersion{}
	if err := gormv2.GetByID(c.HeaderToContext(), lv, id.Uint64()); err != nil {
		c.Error(http.StatusBadRequest, "配置参数获取错误:"+err.Error())
	}
	c.Data["lv"] = lv
}

func (c *DeviceLatestVersionController) Edit() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "device_last_ver/edit.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/device_last_ver"
		c.Data["config"].(map[string]interface{})["actionname"] = "edit"
		c.Data["title"] = "evse version"
		return
	}

	var form struct {
		LatestOtaVersion int    `form:"latest_ota_version"`
		LatestOtaAddress string `form:"latest_ota_address"`
		LatestOtaDesc    string `form:"latest_ota_desc"`
	}
	if err := c.ParseForm(&form); err != nil {
		c.Error(http.StatusBadRequest, "ParseForm error: "+err.Error())
	}
	if form.LatestOtaVersion <= 0 || form.LatestOtaAddress == "" {
		c.Error(http.StatusBadRequest, "req address is nil")
	}

	lv := c.Data["lv"].(*models.LatestFirmwareVersion)
	lv.LastVersion = form.LatestOtaVersion
	lv.UpgradeAddress = form.LatestOtaAddress
	lv.UpgradeDesc = form.LatestOtaDesc
	if err := gormv2.Save(c.HeaderToContext(), lv); err != nil {
		c.Error(http.StatusInternalServerError, "save error: "+err.Error())
	}

	c.Msg = "request success"
}
