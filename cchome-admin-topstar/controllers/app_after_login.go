package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/internal/appproto"
	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/internal/evsectl"
	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/models"
	"github.com/chenwm-topstar/chargingc/pbs/evsepb"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/redigo"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
	"github.com/garyburd/redigo/redis"
)

type AppAfterLoginController struct {
	AppController
}

func (c *AppAfterLoginController) Prepare() {
	c.AppController.Prepare()

	token := c.Ctx.Request.Header.Get("token")
	if token == "" {
		c.Error(http.StatusUnauthorized, "token is nil. Please login")
	}
	key := token + ":token:private"
	userID, err := redis.String(redigo.Do("get", key))
	if err != nil {
		switch err {
		case redis.ErrNil:
			c.Error(http.StatusUnauthorized, "token expired. Please login")
		default:
			c.Error(http.StatusInternalServerError, "check token error. Please later try again")
		}
	}
	if userID != c.Req.Uid {
		c.Error(http.StatusUnauthorized, "token error. Please login")
	}
	if ttl, err := redis.Uint64(redigo.Do("ttl", key)); err == nil && ttl < 48*3600 {
		redigo.Do("set", key, c.Req.Uid, "ex", 15*24*3600)
	}

	id, _ := uuid.ParseID(userID)
	c.Data["uid"] = id
}

func (c *AppAfterLoginController) Logout() {
	token := c.Ctx.Request.Header.Get("token")
	if token == "" {
		c.Error(http.StatusUnauthorized, "token is nil. Please login")
	}
	key := token + ":token:private"
	redigo.Do("del", key)

}
func (c *AppAfterLoginController) Logoff() {
	privateUser := &models.User{}
	if err := gormv2.Find(c.HeaderToContext(), privateUser, "id=?", c.Data["uid"]); err != nil {
		c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
	}

	if privateUser.IsExists() {
		privateUser.IsLogoff = true
		if err := gormv2.Save(c.HeaderToContext(), privateUser); err != nil {
			c.Error(http.StatusInternalServerError, "internel error. Please try again later!")
		}
	}
	c.Resp.RawData = &appproto.UserLogoffReply{}
}
func (c *AppAfterLoginController) ChangePasswd() {
	req := &appproto.ChangePasswdReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "parse param error: "+err.Error())
	}

	privateUser := &models.User{}
	if err := gormv2.Find(c.HeaderToContext(), privateUser, "id=?", c.Data["uid"]); err != nil {
		c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
	}
	if privateUser.IsNew() {
		c.Error(http.StatusNotFound, fmt.Sprintf("userid[%+v] not found", c.Data["uid"]))
	}
	if privateUser.Passwd != req.CurrentPasswd {
		c.Error(http.StatusBadRequest, "current passwd error")
	}
	privateUser.Passwd = req.NewPasswd
	if err := gormv2.Save(c.HeaderToContext(), privateUser); err != nil {
		c.Error(http.StatusInternalServerError, "update passwd error: "+err.Error())
	}

	c.Resp.RawData = &appproto.ChangePasswdReply{}
}

func (c *AppAfterLoginController) ChangeUserInfo() {
	req := &appproto.ChangeUserInfoReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "parse param error: "+err.Error())
	}

	privateUser := &models.User{}
	if err := gormv2.Find(c.HeaderToContext(), privateUser, "id=?", c.Data["uid"]); err != nil {
		c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
	}
	if privateUser.IsNew() {
		c.Error(http.StatusNotFound, fmt.Sprintf("userid[%+v] not found", c.Data["uid"]))
	}

	updates := make(map[string]interface{})
	if req.Email != "" && privateUser.Email != req.Email {
		updates["email"] = req.Email
	}

	if req.Name != "" && privateUser.Name != req.Name {
		updates["name"] = req.Name
	}

	if len(updates) > 0 {
		if err := gormv2.Model(c.HeaderToContext(), privateUser).UpdateColumns(updates).Error; err != nil {
			c.Error(http.StatusInternalServerError, "update fail. %s", err.Error())
		}
	}

	c.Resp.RawData = &appproto.ChangeUserInfoReply{}
}

func (c *AppAfterLoginController) BindEvse() {
	req := &appproto.UserBindEvseReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	var saves []interface{}

	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	if evse.IsNew() {
		evse.ID = uuid.GetID()
		evse.SN = req.SN
		evse.PN = ""
		evse.Mac = req.Mac
		connector := &models.Connector{
			ID:     uuid.GetID(),
			EvseID: evse.ID,
			CNO:    1,
		}
		saves = append(saves, connector, evse)
	}

	evseBind := &models.EvseBind{}
	if err := gormv2.Last(c.HeaderToContext(), evseBind, "uid=? and sn=?", c.Data["uid"], req.SN); err != nil {
		c.Error(http.StatusInternalServerError, "check bind info error: "+err.Error())
	}
	if evseBind.IsNew() {
		evseBind.UID = c.Data["uid"].(uuid.ID)
		evseBind.SN = req.SN
		evseBind.EvseID = evse.ID

		saves = append(saves, evseBind)
	}
	if len(saves) > 0 {
		if err := gormv2.Saves(c.HeaderToContext(), saves...); err != nil {
			c.Error(http.StatusInternalServerError, "bind evse error: "+err.Error())
		}
	}

	c.Resp.RawData = &appproto.UserBindEvseReply{
		EvseStaticData: appproto.EvseStaticData{
			SN:              evse.SN,
			PileModel:       evse.PN,
			RatedPower:      int(evse.RatedPower),
			RatedMinCurrent: int(evse.RatedMinCurrent),
			RatedMaxCurrent: int(evse.RatedMaxCurrent),
			RatedVoltage:    int(evse.RatedVoltage),
			Mac:             evse.Mac,
			FirmwareVersion: parseVersion(evse.FirmwareVersion),
			BTVersion:       parseVersion(evse.BTVersion),
		},
	}
}
func parseVersion(v string) uint16 {
	ret, _ := strconv.ParseUint(v, 10, 64)
	return uint16(ret)
}
func (c *AppAfterLoginController) UnbindEvse() {
	req := &appproto.UserUnbindEvseReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}

	if err := gormv2.GetDB().Unscoped().Delete(&models.EvseBind{}, "uid=? and sn=?", c.Data["uid"], req.SN).Error; err != nil {
		c.Error(http.StatusInternalServerError, "delete save: "+err.Error())
	}

	c.Resp.RawData = &appproto.UserUnbindEvseReply{}
}

func (c *AppAfterLoginController) ChangeEvseInfo() {
	req := &appproto.ChangeEvseInfoReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}

	if req.SN == "" {
		c.Error(http.StatusBadRequest, "req sn is nil")
	}

	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	if req.Alias != "" {
		if err := gormv2.Model(c.HeaderToContext(), evse).UpdateColumns(map[string]interface{}{"alias": req.Alias}).Error; err != nil {
			c.Error(http.StatusInternalServerError, "update fail. %s", err.Error())
		}
	}

	c.Resp.RawData = &appproto.ChangeEvseInfoReply{}
}

func (c *AppAfterLoginController) EvseList() {
	var pebs []*models.EvseBind
	if err := gormv2.Find(c.HeaderToContext(), &pebs, "uid=?", c.Data["uid"]); err != nil {
		c.Error(http.StatusInternalServerError, "find binds error: "+err.Error())
	}

	reply := &appproto.UserEvsesReply{}
	for _, peb := range pebs {
		if peb.EvseID.Uint64() > 0 {
			evse, err := models.GetEvseByID(peb.EvseID.Uint64())
			if err != nil {
				c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
			}
			reply.EvseInfos = append(reply.EvseInfos, appproto.BindEvseInfo{
				EvseStaticData: appproto.EvseStaticData{
					SN:              evse.SN,
					PileModel:       evse.PN,
					RatedPower:      int(evse.RatedPower),
					RatedMinCurrent: int(evse.RatedMinCurrent),
					RatedMaxCurrent: int(evse.RatedMaxCurrent),
					RatedVoltage:    int(evse.RatedVoltage),
					Mac:             evse.Mac,
					FirmwareVersion: parseVersion(evse.FirmwareVersion),
					BTVersion:       parseVersion(evse.BTVersion),
					Alias: func() string {
						if evse.Alias != "" {
							return evse.Alias
						}
						return evse.SN
					}(),
				},
				Status: int(evse.State),
			})
		}
	}

	c.Resp.RawData = reply
}

func (c *AppAfterLoginController) GetEvseInfo() {
	req := &appproto.EvseInfoReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if req.SN == "" {
		c.Error(http.StatusBadRequest, "req sn is nil")
	}

	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	connector, err := models.GetConnector(evse.ID)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get connector info error: "+err.Error())
	}
	nowt := time.Now().Unix()

	reply := &appproto.EvseInfoReply{
		EvseDynamicData: appproto.EvseDynamicData{
			SN:                   evse.SN,
			OrderID:              connector.RecordID,
			ChargingVoltage:      int(connector.VoltageA),
			ChargingCurrent:      int(connector.CurrentA),
			ChargingPower:        int(connector.Power),
			ChargedElectricity:   int(connector.ConsumedElectric),
			StartChargingTime:    nowt - int64(connector.ChargingTime*60),
			ChargingTime:         int64(connector.ChargingTime),
			Status:               evse.State.String(),
			ConnectingStatus:     int(connector.State),
			ConnectingStatusDesc: connector.State.String(),
			OrderStatus:          0,
			ReservedStartTime:    0,
			ReservedStopTime:     0,
			StartType:            0,
			Phone:                "",
			FaultCode:            connector.FaultCode,
		},
		Alias: func() string {
			if evse.Alias != "" {
				return evse.Alias
			}
			return evse.SN
		}(),
		RatedMinCurrent: int(evse.RatedMinCurrent),
		RatedMaxCurrent: int(evse.RatedMaxCurrent),
		HasCharingPrem:  false,
		SettingCurrent: func() int {
			if connector.CurrentLimit <= 0 {
				return int(evse.RatedMaxCurrent)
			}
			return int(connector.CurrentLimit)
		}(),
	}
	c.Resp.RawData = reply
}
func (c *AppAfterLoginController) StartCharger() {
	req := &appproto.EvseStartReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}

	if req.SN == "" {
		c.Error(http.StatusBadRequest, "sn is nil")
	}

	connector := &models.Connector{}
	if err := gormv2.Last(c.HeaderToContext(), connector, "cno=1 and evse_id in (select id from evses where sn=?)", req.SN); err != nil {
		c.Error(http.StatusInternalServerError, "get connector error: "+err.Error())
	}
	if connector.CurrentLimit != int16(req.ChargingCurrent) {
		if err := gormv2.GetDB().Model(connector).Where("id=?", connector.ID).Update("current_limit", req.ChargingCurrent).Error; err != nil {
			c.Error(http.StatusInternalServerError, "save connector error: "+err.Error())
		}
	}
	switch connector.State {
	case evsepb.ConnectorState_CS_Unavailable,
		evsepb.ConnectorState_CS_Charging,
		evsepb.ConnectorState_CS_SuspendedEVSE,
		evsepb.ConnectorState_CS_SuspendedEV,
		evsepb.ConnectorState_CS_Reserved,
		evsepb.ConnectorState_CS_Faulted,
		evsepb.ConnectorState_CS_Waiting,
		evsepb.ConnectorState_CS_Occupied:
		c.Error(http.StatusNotAcceptable, "connector state not supported charging")
	}

	if err := evsectl.StartCharger(req.SN, uint32(c.Data["uid"].(uuid.ID).Uint64()), int32(req.ChargingCurrent)); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	c.Resp.RawData = &appproto.EvseStartReply{}
}

func (c *AppAfterLoginController) StopCharger() {
	req := &appproto.EvseStopReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if err := evsectl.StopCharger(req.SN); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	c.Resp.RawData = &appproto.EvseStopReply{}
}
func (c *AppAfterLoginController) Reset() {
	req := &appproto.ResetReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if err := evsectl.Reset(req.SN); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	c.Resp.RawData = &appproto.ResetReply{}
}

func (c *AppAfterLoginController) Orders() {
	req := &appproto.OrdersReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}

	db := gormv2.GetDB().Model(&models.EvseRecord{}).Order("start_time desc")
	if req.BeginTime > 0 && req.EndTime >= req.BeginTime {
		db = db.Where("start_time>=? and start_time<=?", req.BeginTime, req.EndTime)
	}
	db.Where("total_electricity > 0 and charge_time > 0")
	if req.SN != "" {
		db = db.Where("sn=?", req.SN)
	} else {
		var sns []string
		if err := gormv2.GetDB().Model(&models.EvseBind{}).Where("uid=?", c.Data["uid"]).Select("sn").Scan(&sns).Error; err != nil {
			c.Error(http.StatusInternalServerError, "load bind sn error: "+err.Error())
		}
		if len(sns) > 0 {
			db = db.Where("sn in (?)", sns)
		} else {
			db = db.Where("uid = ?", c.Data["uid"].(uuid.ID).Uint64())
		}
	}

	count := int64(0)
	if err := db.Count(&count).Error; err != nil {
		c.Error(http.StatusBadRequest, "count record error: "+err.Error())
	}
	if req.Size > 0 {
		db = db.Offset(req.Page * req.Size).Limit(req.Size)
	}

	var records []models.EvseRecord
	if err := db.Find(&records).Error; err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}

	reply := &appproto.OrdersReply{Total: int(count)}

	for _, record := range records {
		reply.Orders = append(reply.Orders, appproto.Order{
			ID:                record.RecordID,
			Sn:                record.SN,
			StartChargingTime: int64(record.StartTime),
			StopChargingTime:  int64(record.StartTime + record.ChargeTime*60),
			Elec:              int(record.TotalElectricity),
			Reason:            fmt.Sprintf("%d", record.StopReason),
			StartType:         0,
			Phone:             "",
		})
	}

	c.Resp.RawData = reply
}

func (c *AppAfterLoginController) SetWhitelistCard() {
	req := &appproto.SetWhitelistCardReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if req.SN == "" {
		c.Error(http.StatusBadRequest, "sn is nil")
	}
	if req.Card == "" {
		c.Error(http.StatusBadRequest, "req card is nil")
	}
	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	err = evsectl.SetWhitelistCard(evse.SN, uint32(c.Data["uid"].(uuid.ID).Uint64()), req.IsDel, req.Card)
	if err != nil {
		c.Error(http.StatusServiceUnavailable, "get reserver info error: "+err.Error())
	}

	c.Resp.RawData = &appproto.SetWhitelistCardReply{}
}

func (c *AppAfterLoginController) GetWhitelistCard() {
	req := &appproto.GetWhitelistCardReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if req.SN == "" {
		c.Error(http.StatusBadRequest, "sn is nil")
	}
	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	cards, err := evsectl.GetWhitelistCard(evse.SN, uint32(c.Data["uid"].(uuid.ID).Uint64()))
	if err != nil {
		c.Error(http.StatusServiceUnavailable, "get reserver info error: "+err.Error())
	}

	c.Resp.RawData = &appproto.GetWhitelistCardReply{
		Cards: cards,
	}
}

func (c *AppAfterLoginController) GetReserverInfo() {
	req := &appproto.GetReserverInfoReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if req.SN == "" {
		c.Error(http.StatusBadRequest, "sn is nil")
	}
	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	ris, err := evsectl.GetReserverInfo(evse.SN, uint32(c.Data["uid"].(uuid.ID).Uint64()))
	if err != nil {
		c.Error(http.StatusServiceUnavailable, "get reserver info error: "+err.Error())
	}

	c.Resp.RawData = &appproto.GetReserverInfoReply{
		ReserverInfos: ris,
	}
}

func (c *AppAfterLoginController) GetWorkMode() {
	req := &appproto.GetWorkModeReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if req.SN == "" {
		c.Error(http.StatusBadRequest, "sn is nil")
	}
	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	c.Resp.RawData = &appproto.GetWorkModeReply{
		WorkMode: evse.WorkMode,
	}
}
func (c *AppAfterLoginController) SetWorkMode() {
	req := &appproto.SetWorkModeReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if req.SN == "" {
		c.Error(http.StatusBadRequest, "sn is nil")
	}
	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	if evse.IsNew() {
		c.Error(http.StatusNotFound, "evse not fund")
	}
	if evse.WorkMode != req.WorkMode {
		if err := evsectl.SetWorkMode(req.SN, uint32(c.Data["uid"].(uuid.ID).Uint64()), req.WorkMode); err != nil {
			c.Error(http.StatusServiceUnavailable, err.Error())
		}
		if e := gormv2.GetDB().Model(evse).Where("id=?", evse.ID).Update("work_mode", req.WorkMode).Error; e != nil {
			c.GetLogger().Error("update work mode error: " + e.Error())
		}
	}

	c.Resp.RawData = &appproto.SetWorkModeReply{}
}

func (c *AppAfterLoginController) SetReserverInfo() {
	req := &appproto.SetReserverInfoReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if req.SN == "" {
		c.Error(http.StatusBadRequest, "sn is nil")
	}
	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	err = evsectl.SetReserverInfo(evse.SN, uint32(c.Data["uid"].(uuid.ID).Uint64()), req.ReserverInfos)
	if err != nil {
		c.Error(http.StatusServiceUnavailable, "get reserver info error: "+err.Error())
	}

	c.Resp.RawData = &appproto.SetReserverInfoReply{}
}
func (c *AppAfterLoginController) SetEvseCurrent() {
	req := &appproto.SetCurrentReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	if req.ChargingCurrent < int(evse.RatedMinCurrent) || req.ChargingCurrent > int(evse.RatedMaxCurrent) {
		c.Error(http.StatusInternalServerError, "set evse current error: "+err.Error())
	}
	if err := evsectl.SetCurrent(req.SN, req.ChargingCurrent); err != nil {
		c.Error(http.StatusBadRequest, "set current error: "+err.Error())
	}
	models.SetConnectorCurrentLimit(evse.ID, req.ChargingCurrent)

	c.Resp.RawData = &appproto.SetCurrentReply{}
}
func (c *AppAfterLoginController) SyncBTOrder() {
	req := &appproto.SyncBTOrderReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	if evse.IsNew() {
		c.Error(http.StatusBadRequest, "evse not found")
	}
	var saves []interface{}
	for _, v := range req.BTOrders {
		if v.RecordID == "" {
			c.GetLogger().Warningf("btOrder:[%+v] param error", v)
			continue
		}
		count, err := gormv2.Count(c.HeaderToContext(), &models.EvseRecord{}, "evse_id=? and record_id=?", evse.ID, v.RecordID)
		if err != nil {
			c.Error(http.StatusInternalServerError, "check order error: "+err.Error())
		}
		if count <= 0 {
			record := &models.EvseRecord{
				UID:              c.Data["uid"].(uuid.ID),
				EvseID:           evse.ID,
				SN:               evse.SN,
				RecordID:         v.RecordID,
				AuthMode:         v.AuthMode,
				StartTime:        v.StartTime,
				ChargeTime:       v.ChargeTime,
				TotalElectricity: uint32(v.TotalElectricity * 1000),
				StopReason:       v.StopReason,
				FaultCode:        v.FaultCode,
			}
			saves = append(saves, record)
		}

	}

	if err = gormv2.Saves(c.HeaderToContext(), saves...); err != nil {
		c.Error(http.StatusInternalServerError, "sync order error: "+err.Error())
	}
}

func (c *AppAfterLoginController) LatestFirmwareVersion() {
	req := &appproto.LatestFirmwareVersionReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if req.SN != "" {
		evse, err := models.GetEvseBySN(req.SN)
		if err != nil {
			c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
		}
		if evse.IsNew() {
			c.Error(http.StatusNotFound, "evse not fund")
		}
		lv := &models.LatestFirmwareVersion{}
		if err = gormv2.Find(context.Background(), lv, "pn=? and vendor=?", evse.PN, evse.Vendor); err != nil {
			c.Error(http.StatusNotFound, "check LatestFirmwareVersion error: "+err.Error())
		}
		c.Resp.RawData = &appproto.LatestFirmwareVersionReply{
			LatestFirmwareVersion: int16(lv.LastVersion),
			LatestFirmwareDesc:    lv.UpgradeDesc,
		}
		return
	}

	c.Resp.RawData = &appproto.LatestFirmwareVersionReply{
		LatestFirmwareVersion: int16(models.LatestFirmwareVersionConfig.LastVersion),
		LatestFirmwareDesc:    models.LatestFirmwareVersionConfig.UpgradeDesc,
	}
}

func (c *AppAfterLoginController) OTAUpgrade() {
	req := &appproto.OTAUpgradeReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if req.SN == "" {
		c.Error(http.StatusBadRequest, "req sn is nil")
	}

	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}
	if evse.IsNew() {
		c.Error(http.StatusNotFound, "evse not found")
	}
	lv := &models.LatestFirmwareVersion{}
	if err = gormv2.Find(context.Background(), lv, "pn=? and vendor=?", evse.PN, evse.Vendor); err != nil {
		c.Error(http.StatusNotFound, "check LatestFirmwareVersion error: "+err.Error())
	}
	c.Resp.RawData = &appproto.LatestFirmwareVersionReply{
		LatestFirmwareVersion: int16(lv.LastVersion),
		LatestFirmwareDesc:    lv.UpgradeDesc,
	}
	if int(parseVersion(evse.FirmwareVersion)) < lv.LastVersion {
		if err := evsectl.Upgrade(req.SN, lv.UpgradeAddress); err != nil {
			c.Error(http.StatusInternalServerError, err.Error())
		}
	}

	c.Resp.RawData = &appproto.OTAUpgradeReply{}
}

func (c *AppAfterLoginController) About() {
	d := &models.Dict{}
	if err := gormv2.GetByID(c.HeaderToContext(), d, uint64(models.KindDictTypeAbout)); err != nil {
		c.Error(http.StatusBadRequest, "get about error:"+err.Error())
	}
	about := d.Val
	if d.IsExists() && d.Val != "" {
		ac := &models.AboutConfig{}
		if err := json.Unmarshal([]byte(d.Val), ac); err == nil {
			about = ac.Content
		}
	}
	c.Resp.RawData = &appproto.AboutReply{
		Content: about,
	}
}
func (c *AppAfterLoginController) QuestionAndAnswer() {
	req := &appproto.QuestionAndAnswerReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if req.Size == 0 {
		req.Size = 5
	}

	count, err := gormv2.Count(c.HeaderToContext(), &models.QAA{}, "1=1")
	if err != nil {
		c.Error(http.StatusBadRequest, "count user error: "+err.Error())
	}

	var list []models.QAA
	if err := gormv2.GetDB().Order("created_at desc").Offset(req.Page * req.Size).Limit(req.Size).Find(&list).Error; err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	reply := &appproto.QuestionAndAnswerReply{
		Total: int(count),
		QAA:   []appproto.QuestionAndAnswer{},
	}

	for _, l := range list {
		reply.QAA = append(reply.QAA, appproto.QuestionAndAnswer{
			Q: l.Q,
			A: l.A,
		})
	}
	c.Resp.RawData = reply
}
func (c *AppAfterLoginController) Feedback() {
	req := &appproto.FeedbackReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if req.Content == "" || req.Email == "" {
		c.Error(http.StatusBadRequest, "email or content is nil")
	}

	feedback := &models.Feedback{
		UID:       c.Data["uid"].(uuid.ID),
		Content:   req.Content,
		IsProcess: false,
		Remark:    "",
		Email:     req.Email,
	}
	if err := gormv2.Save(c.HeaderToContext(), feedback); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	c.Resp.RawData = &appproto.FeedbackReply{}
}
