package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"gitlab.goiot.net/chargingc/cchome-admin/internal/lib"
	"gitlab.goiot.net/chargingc/cchome-admin/models"
	"gitlab.goiot.net/chargingc/utils/gormv2"
)

type OrderListController struct {
	Auth
}

func (c *OrderListController) Prepare() {
	c.Auth.Prepare()
}

func (c *OrderListController) List() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "order/list.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/order"
		c.Data["config"].(map[string]interface{})["actionname"] = "list"
		c.Data["title"] = "order list"
		return
	}

	f := lib.DataTablesRequest{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &f); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}

	db := gormv2.Model(c.HeaderToContext(), &models.EvseRecord{})
	if f.Search != "" {
		db.Where("sn=?", f.Search)
	}

	count := int64(0)
	if err := db.Count(&count).Error; err != nil {
		c.Error(http.StatusBadRequest, "count record error: "+err.Error())
	}

	var records []models.EvseRecord
	if err := db.Order("created_at desc").Offset(f.Offset).Limit(f.Limit).Find(&records).Error; err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}

	c.Total = int(count)
	c.Rows = make([]interface{}, 0)
	for _, record := range records {
		c.Rows = append(c.Rows, map[string]interface{}{
			"id":                record.ID.String(),
			"uid":               record.UID.String(),
			"record_id":         record.RecordID,
			"evse_id":           record.EvseID.String(),
			"sn":                record.SN,
			"auth_id":           record.AuthID,
			"auth_mode":         record.AuthMode,
			"start_time":        time.Unix(int64(record.StartTime), 0).Local().Format(time.RFC3339),
			"charge_time":       record.ChargeTime,
			"total_electricity": record.TotalElectricity,
			"stop_reason":       record.StopReason,
		})
	}
}
