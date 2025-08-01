package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/internal/appproto"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/internal/evsectl"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/models"
	"github.com/chenwm-topstar/pbs/commonpb"
	"github.com/chenwm-topstar/utils/gormv2"
	"github.com/chenwm-topstar/utils/lg"
	"github.com/chenwm-topstar/utils/redigo"
	"github.com/chenwm-topstar/utils/slices"
	"github.com/chenwm-topstar/utils/uuid"
	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/now"
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
		privateUser.IsACT = nil
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

// Reset 重置
func (c *AppAfterLoginController) EvseReboot() {
	req := &appproto.RebootReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if err := evsectl.Reboot(req.SN); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	c.Resp.RawData = &appproto.RebootReply{}
}

// Reset 重置
func (c *AppAfterLoginController) EvseReset() {
	req := &appproto.ResetReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	if err := evsectl.Reset(req.SN); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	c.Resp.RawData = &appproto.ResetReply{}
}

func (c *AppAfterLoginController) Statistics() {
	req := &appproto.StatisticsReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}
	reqmon, err := time.ParseInLocation("2006-01", req.Time, time.Local)
	if err != nil {
		c.Error(http.StatusBadRequest, "time format error: "+err.Error())
	}
	nowt := time.Now().Local()

	var b, e time.Time
	if v := nowt.AddDate(0, -5, 0); v.Before(reqmon) { // 获取最新6个月数据
		b, e = v, nowt
	} else {
		b, e = reqmon.AddDate(0, -2, 0), reqmon.AddDate(0, 3, 0) // 获取前两个月， 后三个月数据
	}
	b = now.With(b).BeginningOfMonth()
	e = now.With(e).EndOfMonth()

	db := gormv2.GetDB().Model(&models.EvseRecord{}).Order("created_at desc")

	var sns []string
	if err := gormv2.GetDB().Model(&models.EvseBind{}).Where("uid=?", c.Data["uid"]).Select("sn").Scan(&sns).Error; err != nil {
		c.Error(http.StatusInternalServerError, "load bind sn error: "+err.Error())
	}
	if len(sns) > 0 {
		db = db.Where("sn in (?) or uid = ?", sns, c.Data["uid"])
	} else {
		db = db.Where("uid = ?", c.Data["uid"].(uuid.ID).Uint64())
	}
	db = db.Where("? < created_at and created_at <= ?", b, e)

	var rets []struct {
		Months string `gorm:"column:months;"`
		C      uint32 `gorm:"column:c;"`
		Elec   uint32 `gorm:"column:elec;"`
	}

	if err = db.Debug().Select("DATE_FORMAT(created_at, '%Y-%m') months, count(id) c, sum(total_electricity) elec").
		Group("months").Find(&rets).Error; err != nil {
		c.Error(http.StatusInternalServerError, "statistics error: "+err.Error())
	}
	reply := &appproto.StatisticsReply{}
	for _, v := range rets {
		reply.StatisticsInfos = append(reply.StatisticsInfos, appproto.StatisticsInfo{
			Time:     v.Months,
			KWh:      float64(v.Elec) / 100,
			TotalNum: v.C,
		})
	}
	reply.StatisticsInfos = autoCompleteTime(b, e, reply.StatisticsInfos)
	c.Resp.RawData = reply
}

func autoCompleteTime(b, e time.Time, statistics []appproto.StatisticsInfo) []appproto.StatisticsInfo {
	ret := make([]appproto.StatisticsInfo, 0, 6)
	for i := b; i.Before(e); i = i.AddDate(0, 1, 0) {
		t, falg := i.Format("2006-01"), false
		for _, s := range statistics {
			if s.Time == t {
				falg = true
				ret = append(ret, s)
			}
		}
		if !falg {
			ret = append(ret, appproto.StatisticsInfo{Time: t})
		}
	}
	return ret
}

func (c *AppAfterLoginController) GetProfile() {
	req := &appproto.GetProfileReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}

	user := &models.User{}
	if err := gormv2.Last(c.HeaderToContext(), user, "id=?", c.Data["uid"]); err != nil {
		c.Error(http.StatusInternalServerError, "get user info error: "+err.Error())
	}

	reply := &appproto.GetProfileReply{Price: user.Price}
	for _, v := range user.Funcs {
		switch v {
		case models.UserFuncAlertNotify:
			reply.AlertNotify = true
		case models.UserFuncStopNotify:
			reply.StopChargingNotify = true
		}
	}

	c.Resp.RawData = reply
}

func (c *AppAfterLoginController) SetProfile() {
	req := &appproto.SetProfileReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}

	user := &models.User{}
	if err := gormv2.Last(c.HeaderToContext(), user, "id=?", c.Data["uid"]); err != nil {
		c.Error(http.StatusInternalServerError, "get user info error: "+err.Error())
	}

	user.Price = req.Price

	funcs := []string{}
	if !slices.ContainString(user.Funcs, models.UserFuncAlertNotify) {
		funcs = append(funcs, models.UserFuncAlertNotify)
	}
	if !slices.ContainString(user.Funcs, models.UserFuncStopNotify) {
		funcs = append(funcs, models.UserFuncStopNotify)
	}
	user.Funcs = funcs

	if err := gormv2.Save(context.Background(), user); err != nil {
		c.Error(http.StatusInternalServerError, "save profile error: "+err.Error())
	}

	c.Resp.RawData = &appproto.SetProfileReply{}
}

func (c *AppAfterLoginController) EvseAuthCode() {
	if ok := c.app.HasShare(); !ok {
		c.Error(http.StatusForbidden, "not support")
	}

	req := &appproto.EvseAuthCodeReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}

	evseBind := &models.EvseBind{}
	if err := gormv2.Last(c.HeaderToContext(), evseBind, "uid=? and sn=?", c.Data["uid"], req.SN); err != nil {
		c.Error(http.StatusInternalServerError, "check bind info error: "+err.Error())
	}
	if evseBind.IsMaster == nil && !(*evseBind.IsMaster) {
		c.Error(http.StatusBadRequest, "Not an administrator")
	}
	reply := &appproto.EvseAuthCodeReply{TotalTime: 120}

	code, ttl, e := models.GetAuthCode(evseBind.EvseID.Uint64())
	if e == nil {
		reply.AuthCode, reply.TTL = code, uint32(ttl)
	} else {
		code = fmt.Sprintf("%04d", uuid.GetID()%10000)
		if err := models.SetAuthCode(evseBind.EvseID.Uint64(), code, 123); err != nil {
			c.Error(http.StatusBadRequest, "gen code error: "+err.Error())
		}
		reply.AuthCode, reply.TTL = code, 120
	}

	c.Resp.RawData = reply
}

func (c *AppAfterLoginController) BindMembers() {
	req := &appproto.BindMembersReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}

	list := []models.EvseBind{}
	if err := gormv2.Find(c.HeaderToContext(), &list, "appid=? and sn=?", c.app.ID, req.SN); err != nil {
		c.Error(http.StatusInternalServerError, "check bind info error: "+err.Error())
	}

	var master appproto.BindMember
	var slave []appproto.BindMember
	for _, l := range list {
		user, err := models.GetUserByID(l.UID.Uint64())
		if err != nil {
			c.Error(http.StatusInternalServerError, "load error: "+err.Error())
		}
		tmp := appproto.BindMember{
			Id:       l.UID.String(),
			Nickname: user.Name,
			IsMaster: func() bool {
				if l.IsMaster != nil && *l.IsMaster {
					return true
				}
				return false
			}(),
			EnableCharing: false,
		}
		if tmp.IsMaster {
			master = tmp
		} else {
			slave = append(slave, tmp)
		}
	}

	reply := &appproto.BindMembersReply{}
	if c.Data["uid"].(uuid.ID).String() == master.Id {
		reply.BindMembers = slave
	} else {
		reply.BindMembers = append(reply.BindMembers, master)
	}

	c.Resp.RawData = reply
}

func (c *AppAfterLoginController) BindEvse() {
	req := &appproto.UserBindEvseReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "decode error: "+err.Error())
	}

	unlockf, err := redigo.Lock(fmt.Sprintf("%s:cchome", req.SN), 10)
	if err != nil {
		c.Error(http.StatusBadRequest, "lock error: "+err.Error())
	}
	defer unlockf()

	var saves []interface{}

	evse, err := models.GetEvseBySN(req.SN)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get evse info error: "+err.Error())
	}

	isMaster := false
	if evse.IsNew() {
		isMaster = true

		evse.ID = uuid.GetID()
		evse.SN = req.SN
		evse.PN = ""
		evse.AndroidMac = req.AndroidMac
		evse.IOSMac = req.IOSMac
		connector := &models.Connector{
			ID:     uuid.GetID(),
			EvseID: evse.ID,
			CNO:    1,
		}
		saves = append(saves, connector, evse)
	} else {
		flag := false
		if req.AndroidMac != "" && req.AndroidMac != evse.AndroidMac {
			evse.AndroidMac, flag = req.AndroidMac, true
		}
		if req.IOSMac != "" && req.IOSMac != evse.IOSMac {
			evse.IOSMac, flag = req.IOSMac, true
		}
		if flag {
			saves = append(saves, evse)
		}
	}

	evseBind := &models.EvseBind{}
	if err := gormv2.Last(c.HeaderToContext(), evseBind, "appid=? and uid=? and evse_id=?", c.app.ID, c.Data["uid"], evse.ID); err != nil {
		c.Error(http.StatusInternalServerError, "check bind info error: "+err.Error())
	}
	if evseBind.IsNew() {
		if c.app.HasShare() {
			if evse.IsExists() { // 查看是否是第一个绑定的
				cnt, err := gormv2.Count(c.HeaderToContext(), evseBind, "appid=? and evse_id=?", c.app.ID, evse.ID)
				if err != nil {
					c.Error(http.StatusInternalServerError, "check bind error: "+err.Error())
				}
				if cnt == 0 {
					isMaster = true
				}
			}

			if !isMaster { // 校验设备code
				authCode, _, err := models.GetAuthCode(evse.ID.Uint64())
				if err != nil {
					c.Error(http.StatusInternalServerError, "get auth error: "+err.Error())
				}
				if ok := c.app.HasShare(); ok {
					if authCode != req.Auth {
						c.Error(http.StatusBadRequest, "auth code error")
					}
				}

			}
		}

		evseBind.UID = c.Data["uid"].(uuid.ID)
		evseBind.SN = req.SN
		evseBind.APPID = c.app.ID
		evseBind.EvseID = evse.ID
		if isMaster {
			tmp := true
			evseBind.IsMaster = &tmp
		}

		saves = append(saves, evseBind)
	}
	if evseBind.IsMaster != nil && *evseBind.IsMaster && evse.IsExists() && evse.NetWork == models.NetWork4G {
		zone, err := models.GetUserZone(evseBind.UID)
		if err != nil {
			c.Error(http.StatusInternalServerError, "load user error: "+err.Error())
		}
		if e := evsectl.SetTimezone(evse.SN, zone); e != nil {
			// c.Error(http.StatusInternalServerError,)
			lg.Warnf("set zone error: " + e.Error())
		}
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
			AndroidMac:      evse.AndroidMac,
			IOSMac:          evse.IOSMac,
			FirmwareVersion: parseVersion(evse.FirmwareVersion),
			BTVersion:       parseVersion(evse.BTVersion),
			NetWork:         int(evse.NetWork),
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
	evseBind := &models.EvseBind{}
	if err := gormv2.Find(c.HeaderToContext(), evseBind, "uid=? and sn=?", c.Data["uid"], req.SN); err != nil {
		c.Error(http.StatusInternalServerError, "check bind error: "+err.Error())
	}

	if evseBind.IsExists() {
		if evseBind.IsMaster != nil && *evseBind.IsMaster && req.MemID == "" {
			if err := gormv2.GetDB().Unscoped().Delete(&models.EvseBind{}, "sn=?", req.SN).Error; err != nil {
				c.Error(http.StatusInternalServerError, "unbind all error: "+err.Error())
			}
		} else {
			uid := req.MemID
			if uid == "" {
				uid = c.Data["uid"].(uuid.ID).String()
			}
			if err := gormv2.GetDB().Unscoped().Delete(&models.EvseBind{}, "uid=? and sn=?", uid, req.SN).Error; err != nil {
				c.Error(http.StatusInternalServerError, "unbind error: "+err.Error())
			}
		}
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
					AndroidMac:      evse.AndroidMac,
					IOSMac:          evse.IOSMac,
					Mac:             evse.AndroidMac,
					FirmwareVersion: parseVersion(evse.FirmwareVersion),
					BTVersion:       parseVersion(evse.BTVersion),
					NetWork:         int(evse.NetWork),
					Alias: func() string {
						if evse.Alias != "" {
							return evse.Alias
						}
						return evse.SN
					}(),
				},
				State:  int(evse.State),
				Status: int(evse.State),
				IsMaster: func() bool {
					if peb.IsMaster != nil {
						return *peb.IsMaster
					}
					return false
				}(),
			})
		}
	}

	c.Resp.RawData = reply
}

func syncZone(uid uuid.ID, sn string) {
	zone, err := models.GetUserZone(uid)
	if err != nil {
		lg.Warnf("set user zone error: " + err.Error())
	}
	if zone == 0 {
		return
	}

	key := fmt.Errorf("%s:sync:cchome-admin", sn)
	if ttl, err := redis.Uint64(redigo.Do("ttl", key)); err == nil && ttl < 10 {
		if e := evsectl.SetTimezone(sn, zone); e != nil {
			lg.Warnf("set zone error: " + e.Error())
		}
		redigo.Do("set", key, uid.String(), "ex", 24*3600) // 24小时对时一次
	}
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

	evseBind := &models.EvseBind{}
	if err := gormv2.Find(c.HeaderToContext(), evseBind, "sn=? and is_master=?", req.SN, true); err != nil {
		c.Error(http.StatusInternalServerError, "check bind error: "+err.Error())
	}
	supportBindMode := 0
	if evseBind.IsExists() && evseBind.UID != c.Data["uid"].(uuid.ID) {
		supportBindMode = 1
	}
	go syncZone(c.Data["uid"].(uuid.ID), req.SN)

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
			State:                int(evse.State),
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
		SupportBindMode: supportBindMode,
		NetWork:         int(evse.NetWork),
		WorkMode:        int(evse.WorkMode),
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
	case commonpb.ConnectorState_CS_Unavailable,
		commonpb.ConnectorState_CS_Charging,
		commonpb.ConnectorState_CS_SuspendedEVSE,
		commonpb.ConnectorState_CS_SuspendedEV,
		commonpb.ConnectorState_CS_Reserved,
		commonpb.ConnectorState_CS_Faulted,
		commonpb.ConnectorState_CS_Waiting,
		commonpb.ConnectorState_CS_Occupied:
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

	privateUser := &models.User{}
	if err := gormv2.Find(c.HeaderToContext(), privateUser, "id=?", c.Data["uid"]); err != nil {
		c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
	}

	db := gormv2.GetDB().Model(&models.EvseRecord{}).Order("start_time desc")
	if req.BeginTime > 0 && req.EndTime >= req.BeginTime {
		db = db.Where("start_time>=? and start_time<=?", req.BeginTime, req.EndTime)
	}
	// db = db.Where("total_electricity > 0 and charge_time > 0")
	if req.SN != "" {
		db = db.Where("sn=?", req.SN)
	} else {
		var sns []string
		if err := gormv2.GetDB().Model(&models.EvseBind{}).Where("uid=?", c.Data["uid"]).Select("sn").Scan(&sns).Error; err != nil {
			c.Error(http.StatusInternalServerError, "load bind sn error: "+err.Error())
		}
		if len(sns) > 0 {
			db = db.Where("sn in (?) or uid = ?", sns, c.Data["uid"])
		} else {
			db = db.Where("uid = ?", c.Data["uid"].(uuid.ID).Uint64())
		}
	}
	db = db.Where("created_at > ?", privateUser.CreatedAt)

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
			StopChargingTime:  int64(record.StartTime + record.ChargeTime),
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
	language := c.Ctx.Request.Header.Get("language")

	count, err := gormv2.Count(c.HeaderToContext(), &models.QAA{}, "language=?", language)
	if err != nil {
		c.Error(http.StatusBadRequest, "count user error: "+err.Error())
	}

	var list []models.QAA
	if err := gormv2.GetDB().Where("language=?", language).Order("created_at desc").Offset(req.Page * req.Size).Limit(req.Size).Find(&list).Error; err != nil {
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

func (c *AppAfterLoginController) GetApikeyInfo() {
	c.Resp.RawData = &appproto.GetApikeyInfoReply{
		FID:    c.app.Config.Fid,
		Apikey: c.app.Config.Apikey,
	}
}
