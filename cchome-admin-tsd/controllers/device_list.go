package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/internal/lib"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/models"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
)

type DeviceListController struct {
	Auth
}

func (c *DeviceListController) Prepare() {
	c.Auth.Prepare()
}

func (c *DeviceListController) List() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "device/list.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/device"
		c.Data["config"].(map[string]interface{})["actionname"] = "list"
		c.Data["title"] = "evse list"
		return
	}

	f := lib.DataTablesRequest{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &f); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	db := gormv2.Model(c.HeaderToContext(), &models.Evse{})
	if f.Search != "" {
		db.Where("sn=?", f.Search)
	}

	evseCount := int64(0)
	if err := db.Count(&evseCount).Error; err != nil {
		c.Error(http.StatusBadRequest, "count evse error: "+err.Error())
	}

	var evses []models.Evse
	if err := db.Order("updated_at desc").Offset(f.Offset).Limit(f.Limit).Find(&evses).Error; err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}

	c.Total = int(evseCount)
	c.Rows = make([]interface{}, 0)
	for _, evse := range evses {
		c.Rows = append(c.Rows, map[string]interface{}{
			"id":                evse.ID.String(),
			"pn":                evse.PN,
			"sn":                evse.SN,
			"state":             evse.State,
			"firmware_version":  evse.FirmwareVersion,
			"protocol_version":  evse.BTVersion,
			"rated_min_current": evse.RatedMinCurrent,
			"rated_max_current": evse.RatedMaxCurrent,
			"android_mac":       evse.AndroidMac,
			"ios_mac":           evse.IOSMac,
		})
	}
}
