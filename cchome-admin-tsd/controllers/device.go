package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdinternal/evsectl"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdmodels"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"gopkg.in/yaml.v2"
)

type DeviceController struct {
	Auth
}

func (c *DeviceController) Prepare() {
	c.Auth.Prepare()

	evseidstr := c.GetString("evseid", "")
	if evseidstr == "" {
		c.Error(http.StatusBadRequest, "evseid is nil")
	}

	evseid, _ := strconv.ParseUint(evseidstr, 10, 64)

	evse, err := models.GetEvseByID(evseid)
	if err != nil {
		c.Error(http.StatusBadRequest, "设备信息获取错误:"+err.Error())
	}

	no, _ := c.GetInt("no", 0)
	if no == 0 {
		no = 1
	}
	c.Data["evse"] = evse
	c.Data["evseid"] = evseidstr
	c.Data["no"] = no

	c.Data["config"].(map[string]interface{})["evse"] = map[string]interface{}{
		"pn":     evse.PN,
		"sn":     evse.SN,
		"evseid": evseidstr,
		"no":     no,
	}
}

func (c *DeviceController) Info() {
	evse := c.Data["evse"].(*models.Evse)
	if !c.Ctx.Input.IsPost() {
		c.TplName = "device/info.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/device"
		c.Data["config"].(map[string]interface{})["actionname"] = "info"
		c.Data["title"] = "devcie info"
		c.Data["name"] = c.getManager().GetModel().Name
		return
	}

	var infos []interface{}
	infos = append(infos, map[string]interface{}{"key": "ID", "val": evse.ID.String()})
	infos = append(infos, map[string]interface{}{"key": "PN", "val": evse.PN})
	infos = append(infos, map[string]interface{}{"key": "SN", "val": evse.SN})
	infos = append(infos, map[string]interface{}{"key": "vendor", "val": "goiot"})
	infos = append(infos, map[string]interface{}{"key": "firmware version", "val": evse.FirmwareVersion})
	infos = append(infos, map[string]interface{}{"key": "protocol version", "val": evse.BTVersion})
	infos = append(infos, map[string]interface{}{"key": "connector number", "val": evse.CNum})
	infos = append(infos, map[string]interface{}{"key": "android_mac", "val": evse.AndroidMac})
	infos = append(infos, map[string]interface{}{"key": "ios_mac", "val": evse.IOSMac})

	c.Total = len(infos)
	c.Rows = infos
}

func (c *DeviceController) Connector() {
	if !c.Ctx.Input.IsPost() {
		c.AddBreadCrumb("", "connector")
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/device/evse"
		c.Data["config"].(map[string]interface{})["actionname"] = "info"
		c.TplName = "device/info.html"
	}
	evse := c.Data["evse"].(*models.Evse)

	var connectors []models.Connector
	if err := gormv2.Find(c.HeaderToContext(), &connectors, "evse_id=?", evse.ID); err != nil {
		c.Error(http.StatusInternalServerError, "get connector error: "+err.Error())
	}

	var rows []interface{}
	for _, connector := range connectors {
		rows = append(rows, map[string]interface{}{
			"id":          connector.ID.String(),
			"cno":         connector.CNO,
			"state":       connector.State,
			"current":     fmt.Sprintf("%.1f", float64(connector.CurrentA)/10),
			"voltage":     fmt.Sprintf("%.1f", float64(connector.VoltageA)/10),
			"power":       fmt.Sprintf("%.2f", float64(connector.Power)/100),
			"electricity": fmt.Sprintf("%.3f", float64(connector.ConsumedElectric)/1000),
			"record_id":   connector.RecordID,
			"fault_code":  connector.FaultCode,
		})
	}

	c.Rows = rows
	c.Total = len(c.Rows)
}

// StartCharger 启动充电
func (c *DeviceController) StartCharger() {
	evse := c.Data["evse"].(*models.Evse)

	if err := evsectl.StartCharger(evse.SN, 0, evse.RatedMaxCurrent); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	c.Msg = "start charger success"
}

// StopCharger 启动充电
func (c *DeviceController) StopCharger() {
	evse := c.Data["evse"].(*models.Evse)

	if err := evsectl.StopCharger(evse.SN); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	c.Msg = "stop charger success"
}

// Reset 重置
func (c *DeviceController) Reboot() {
	evse := c.Data["evse"].(*models.Evse)

	if err := evsectl.Reboot(evse.SN); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	c.Msg = "reboot success"
}

// Reset 重置
func (c *DeviceController) Reset() {
	evse := c.Data["evse"].(*models.Evse)

	if err := evsectl.Reset(evse.SN); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	c.Msg = "reset success"
}

// TriggerTelemetry 触发遥测
func (c *DeviceController) TriggerTelemetry() {
	evse := c.Data["evse"].(*models.Evse)
	if err := evsectl.TriggerTelemetry(evse.SN); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	c.Msg = "trigger telemetry success"
}

// Upgrade 固件升级
func (c *DeviceController) Upgrade() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "device/upgrade.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/device"
		return
	}
	evse := c.Data["evse"].(*models.Evse)

	var form struct {
		UpgradeFTP string `form:"upgrade_ftp"`
	}
	if err := c.ParseForm(&form); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	if form.UpgradeFTP == "" {
		c.Error(http.StatusBadRequest, "upgrate address is nil")
	}

	if err := evsectl.Upgrade(evse.SN, form.UpgradeFTP); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}

	c.Msg = "request success"
}

// GetAppointConfig 获取指定参数
func (c *DeviceController) GetAppointConfig() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "device/getappointconfig.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/device"
		c.Data["config"].(map[string]interface{})["actionname"] = "getappointconfig"
		return
	}
	evse := c.Data["evse"].(*models.Evse)

	var form struct {
		ConfigKey string `form:"config_key"`
	}
	if err := c.ParseForm(&form); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	} else if form.ConfigKey == "" {
		c.Error(http.StatusBadRequest, "param is nil")
	} else if strings.Contains(form.ConfigKey, ": ") {
		c.Error(http.StatusBadRequest, "format error")
	}
	form.ConfigKey = strings.Trim(form.ConfigKey, "\r\n")
	form.ConfigKey = strings.Trim(form.ConfigKey, "\n")
	form.ConfigKey = strings.Trim(form.ConfigKey, " ")

	content, err := evsectl.GetConfig(evse.SN, form.ConfigKey)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get param error: "+err.Error())
	}
	configparams := fmt.Sprintf("%s: %s\r\n", form.ConfigKey, content)

	if c.JsonData == nil {
		c.JsonData = make(map[string]interface{})
	}
	c.JsonData["params"] = configparams
	c.Msg = "get param success"
}

// SetConfig 设置参数
func (c *DeviceController) SetConfig() {
	if !c.Ctx.Input.IsPost() {
		c.TplName = "device/setconfig.html"
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/device"
		c.Data["config"].(map[string]interface{})["actionname"] = "setconfig"
		return
	}
	evse := c.Data["evse"].(*models.Evse)

	var form struct {
		ConfigData string `form:"config_data"`
	}
	if err := c.ParseForm(&form); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	} else if form.ConfigData == "" {
		c.Error(http.StatusBadRequest, "config param is nil")
	}

	configs := make(map[string]string)
	if err := yaml.Unmarshal([]byte(form.ConfigData), &configs); err != nil {
		c.Error(http.StatusBadRequest, "yaml format error:"+err.Error())
	}
	if len(configs) > 1 {
		c.Error(http.StatusBadRequest, "not support muitl config")
	}

	for k, v := range configs {
		if err := evsectl.SetConfig(evse.SN, k, v); err != nil {
			c.Error(http.StatusBadRequest, err.Error())
		}
	}

	c.Msg = "set config error"
}
