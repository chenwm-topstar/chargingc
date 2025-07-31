package protocol

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/models"
	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/transac/itransac"
	"github.com/chenwm-topstar/chargingc/pbs/evsepb"
	"github.com/chenwm-topstar/chargingc/utils/access/codec"
	"github.com/chenwm-topstar/chargingc/utils/access/driver"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
	"github.com/sirupsen/logrus"
)

type BootReq struct {
	Model             driver.Byte16
	Vendor            driver.Byte16
	CNum              uint8
	MinCurrent        uint8
	MaxCurrent        uint8
	FirmwareVersion   uint16
	BTVersion         uint8
	BTMac             driver.Byte20
	TotalChargeNum    uint32
	TotalExceptionNum uint32
	BTStatus          uint8
	Standard          uint8
	Phase             uint8
}

type BootConf struct {
	State    uint8
	Template uint32
	TimeZone uint8
}

func (bn *BootReq) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	apdu := ctx.Data["apdu"].(*APDU)
	evse := ctx.Data["evse"].(*models.Evse)

	var saves []interface{}
	if evse.IsNew() {
		evse.ID = uuid.GetID()
		evse.SN = apdu.SN.String()
		evse.Alias = evse.SN

		connector := &models.Connector{
			ID:     uuid.GetID(),
			EvseID: evse.ID,
			CNO:    1,
		}
		saves = append(saves, connector)
	}
	evse.PN = bn.Model.String()
	evse.Mac = bn.BTMac.String()
	evse.Vendor = bn.Vendor.String()
	evse.CNum = bn.CNum
	evse.State = evsepb.EvseState_ES_ONLINE
	evse.FirmwareVersion = fmt.Sprintf("%d", bn.FirmwareVersion)
	evse.BTVersion = fmt.Sprintf("%d", bn.BTVersion)
	evse.Standard = getEvseStandard(bn.Standard)
	evse.RatedMinCurrent = int32(bn.MinCurrent)
	evse.RatedMaxCurrent = int32(bn.MaxCurrent)

	saves = append(saves, evse)
	if err := gormv2.Saves(context.Background(), saves...); err != nil {
		ctx.Kick = true
		return nil, fmt.Errorf("保存设备信息错误:" + err.Error())
	}
	ctx.Mark = evse.SN
	ctx.Data["retcmd"] = CmdBootConf

	_apdu := &APDU{
		Seq: apdu.Seq,
		Cmd: CmdBootConf,
		Payload: &BootConf{
			State:    0,
			Template: uint32(time.Now().Unix()),
			TimeZone: 0,
		},
	}
	copy(_apdu.SN[:], apdu.SN[:])

	retapdu, err = _apdu.Marshal()

	return
}

type HeartbeatReq struct {
}

type HeartbeatConf struct {
}

type GetConfigReq struct {
	ConfNameLen uint16
	ConfName    driver.Bytes
}

func (rc *GetConfigReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-getConfig", ctx.Mark)
	return nil
}

type GetConfigConf struct {
	ConfNameLen uint16
	ConfValLen  uint16
	ConfName    driver.Bytes `len_inx:"-2"`
	ConfVal     driver.Bytes `len_inx:"-2"`
}

func (rc *GetConfigConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-getConfig", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type SetConfigReq struct {
	ConfNameLen uint16
	ConfValLen  uint16
	ConfName    driver.Bytes
	ConfVal     driver.Bytes
}

func (rc *SetConfigReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-setConfig", ctx.Mark)
	return nil
}

type SetConfigConf struct {
	Status uint8
}

func (rc *SetConfigConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	ctx.Log.Data["retcode"] = rc.Status
	sess := itransac.LoadSession(fmt.Sprintf("%s-setConfig", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

func (h *HeartbeatReq) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	apdu := ctx.Data["apdu"].(*APDU)

	ctx.Data["retcmd"] = CmdHeartbeatConf

	_apdu := &APDU{
		Seq:     apdu.Seq,
		Cmd:     CmdHeartbeatConf,
		Payload: &HeartbeatConf{},
	}
	copy(_apdu.SN[:], apdu.SN[:])

	return _apdu.Marshal()
}

type GetReserverInfoReq struct {
	UserType uint8
	UserID   uint32
}

func (rc *GetReserverInfoReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-GetReserverInfo", ctx.Mark)
	return nil
}

type ReserverInfo struct {
	Repeat     uint8
	StartTime  uint32
	ChargeTime uint16
}

type GetReserverInfoConf struct {
	UserType      uint8
	UserID        uint32
	ReserverInfos []ReserverInfo
}

func (rc *GetReserverInfoConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-GetReserverInfo", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type GetWhitelistReq struct {
	UserType uint8
	UserID   uint32
}

func (rc *GetWhitelistReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-GetWhitelist", ctx.Mark)
	return nil
}

type GetWhitelistConf struct {
	TNum  uint8
	Cards []driver.Byte16
}

func (rc *GetWhitelistConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-GetWhitelist", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type SetWhitelistReq struct {
	Func uint8
	Card driver.Byte16
}

func (rc *SetWhitelistReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-SetWhitelist", ctx.Mark)
	return nil
}

type SetWhitelistConf struct {
	Status uint8
}

func (rc *SetWhitelistConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-SetWhitelist", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type SetReserverInfoReq struct {
	UserType      uint8
	UserID        uint32
	ReserverInfos []ReserverInfo
}

func (rc *SetReserverInfoReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-SetReserverInfo", ctx.Mark)
	return nil
}

type SetReserverInfoConf struct {
	Status uint8
}

func (rc *SetReserverInfoConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-SetReserverInfo", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type SetWorkModeReq struct {
	UserType uint8
	UserID   uint32
	WorkMode uint8
}

func (rc *SetWorkModeReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-SetWorkMode", ctx.Mark)
	return nil
}

type SetWorkModeConf struct {
	Status uint8
}

func (rc *SetWorkModeConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-SetWorkMode", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type RemoteCtrlReq struct {
	UserType        uint8
	UserID          uint32
	ChargingCurrent uint8
	Command         uint8
}

func (rc *RemoteCtrlReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-remotecontrol", ctx.Mark)
	return nil
}

type RemoteCtrlConf struct {
	Status    uint8
	StartTime uint32
	Meter     uint32
}

func (rc *RemoteCtrlConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	ctx.Log.Data["retcode"] = rc.Status
	sess := itransac.LoadSession(fmt.Sprintf("%s-remotecontrol", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type TriggerTelemetryReq struct {
	UserType uint8
	UserID   uint32
}

func (rc *TriggerTelemetryReq) ToDevicePayload(ctx *itransac.Ctx) error {

	return nil
}

type TelemetryReq struct {
	Status           uint8
	FaultCode        uint8
	SetOutputCurrent uint16
	WorkMode         uint8
	Voltage          uint16
	Current          uint16
	Power            uint16
	ConsumedElectric uint32
	Meter            uint32
	ChargingTime     uint16
	AuthMode         uint8
	RecordID         driver.Byte32
}

func (rc *TelemetryReq) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	evse := ctx.Data["evse"].(*models.Evse)

	connect := &models.Connector{}
	if err = gormv2.MustFind(context.Background(), connect, "evse_id=? and cno=1", evse.ID); err != nil {
		return
	}
	if evse.WorkMode != rc.WorkMode {
		evse.WorkMode = rc.WorkMode
		if e := gormv2.GetDB().Model(evse).Where("id=?", evse.ID).Update("work_mode", rc.WorkMode).Error; e != nil {
			ctx.Log.Error("update work mode error: " + e.Error())
		}
	}

	connect.CurrentLimit = int16(rc.SetOutputCurrent / 10)
	connect.State = getConnectState(rc.Status)
	connect.RecordID = rc.RecordID.String()
	connect.Power = uint32(rc.Power)
	connect.CurrentA = uint32(rc.Current)
	connect.VoltageA = uint32(rc.Voltage)
	connect.ConsumedElectric = uint32(rc.ConsumedElectric)
	connect.ChargingTime = rc.ChargingTime
	connect.FaultCode = uint16(rc.FaultCode)

	if err = gormv2.Saves(context.Background(), connect); err != nil {
		return
	}

	return
}

type TransctionReq struct {
	UserID           uint32
	AuthMode         uint8
	RecordID         driver.Byte32
	StartTime        uint32
	ChargeTime       uint32
	TotalElectricity uint32
	Meter            uint32
	StopReason       uint8
	FaultCode        uint8
}

func (p *TransctionReq) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	apdu := ctx.Data["apdu"].(*APDU)
	evse := ctx.Data["evse"].(*models.Evse)

	er := &models.EvseRecord{}
	if err = gormv2.Last(context.Background(), er, "evse_id=? and record_id=?", evse.ID, p.RecordID.String()); err != nil {
		return
	}
	if er.IsNew() {
		er.ID = uuid.GetID()
		er.UID = uuid.ID(p.UserID)
		er.EvseID = evse.ID
		er.SN = evse.SN
		er.RecordID = p.RecordID.String()
		er.AuthID = fmt.Sprintf("%d", p.UserID)
		er.AuthMode = p.AuthMode
		er.StartTime = p.StartTime
		er.ChargeTime = p.ChargeTime
		er.TotalElectricity = p.TotalElectricity
		er.StopReason = p.StopReason
		er.FaultCode = p.FaultCode
		if p.AuthMode == 0 {
			eb := &models.EvseBind{}
			if err = gormv2.Last(context.Background(), eb, "evse_id=?", evse.ID); err != nil {
				return
			}
			if eb.IsExists() {
				er.UID = eb.UID
			}
		}
		if err = gormv2.Save(context.Background(), er); err != nil {
			return
		}
	}
	ctx.Data["retcmd"] = CmdTransactionConf

	tc := &TransctionConf{State: 0}
	copy(tc.RecordID[:], p.RecordID[:])

	_apdu := &APDU{
		Seq:     apdu.Seq,
		Cmd:     CmdTransactionConf,
		Payload: tc,
	}
	copy(_apdu.SN[:], apdu.SN[:])

	retapdu, err = _apdu.Marshal()

	return
}

type UpdateFirmwareReq struct {
	FTPAddress driver.Byte192
}
type UpdateFirmwareConf struct {
	NowVersion uint8
	Status     uint8
}

func (rc *UpdateFirmwareReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-UpdateFirmware", ctx.Mark)
	return nil
}

func (rc *UpdateFirmwareConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	ctx.Log.Data["retcode"] = rc.Status
	sess := itransac.LoadSession(fmt.Sprintf("%s-UpdateFirmware", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type TransctionConf struct {
	RecordID driver.Byte32
	State    uint8
}

type APDU struct {
	Head    uint8
	Length  uint8
	Seq     uint8
	Cmd     Cmd
	SN      driver.Byte16
	Payload interface{}
	Check   uint8
}

func NewAPDU() *APDU {
	return &APDU{}
}

func (apdu *APDU) ToAPDU(ctx *itransac.Ctx) (ret []byte, err error) {
	b := ctx.Raw.([]byte)
	ctx.Log = logrus.WithFields(logrus.Fields{
		"method": "toapdu",
		"buf":    fmt.Sprintf("[%d][%x]", len(b), b),
		"sn":     ctx.Mark,
	})

	if err = apdu.Unmarshal(b); err != nil {
		goto _ret_toapdu
	}
	ctx.Data["apdu"] = apdu

	if apdu.Cmd != CmdHeartbeatReq {
		evse, err := models.GetEvseBySN(apdu.SN.String())
		if err != nil {
			goto _ret_toapdu
		}
		ctx.Data["evse"] = evse
	}

	ret, err = apdu.Payload.(IUpPayload).ToPlatformPayload(ctx)

_ret_toapdu:
	if err != nil {
		ctx.Log.Errorf("to %#x error:[%s]", apdu.Cmd, err.Error())
	} else {
		ctx.Log.Infof("to %#x payload:[%+v] ", apdu.Cmd, apdu.Payload)
		if ret != nil {
			ctx.Log.Data["buf"] = fmt.Sprintf("%#x", ret)
			ctx.Log.Infof("ret %#x", ctx.Data["retcmd"])
		}
	}
	return
}

func (apdu *APDU) FromAPDU(ctx *itransac.Ctx) (buf []byte, err error) {
	ctx.Log = logrus.WithFields(logrus.Fields{
		"method": "fromapdu",
		"sn":     ctx.Mark,
	})

	apdu.Payload = ctx.Raw
	apdu.Payload.(IDownPayload).ToDevicePayload(ctx)
	buf, err = apdu.Marshal()
	if err != nil {
		ctx.Log.Errorf("from error:[%s]", err.Error())
	} else {
		ctx.Log.Infof("from %#x payload:[%+v]", apdu.Cmd, apdu.Payload)
	}
	return
}

func (a *APDU) Unmarshal(b []byte) (err error) {
	l := len(b)
	if !checkSum(b) {
		return errors.New("数据报文校验无法通过")
	}
	if err = codec.Unmarshal(b[:20], a); err != nil {
		return err
	}
	if a.Payload, err = getPayloadByCMD(a.Cmd); err != nil {
		return err
	}
	if l > 21 {
		if err = codec.Unmarshal(b[20:l-1], a.Payload); err != nil {
			return err
		}
	}

	a.Check = uint8(b[l-1])
	return nil
}

func (a *APDU) Marshal() (buf []byte, err error) {
	buf, err = codec.Marshal(a)
	if err != nil {
		return nil, err
	}
	buf[0] = 0x68
	buf[1] = uint8(len(buf) - 2)
	return addSum(buf), nil
}
