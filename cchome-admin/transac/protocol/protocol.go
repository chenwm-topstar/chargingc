package protocol

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chenwm-topstar/chargingc/cchome-admin/models"
	"github.com/chenwm-topstar/chargingc/cchome-admin/transac/itransac"
	"github.com/chenwm-topstar/chargingc/pbs/commonpb"
	"github.com/chenwm-topstar/chargingc/utils/abiz/access/codec"
	"github.com/chenwm-topstar/chargingc/utils/abiz/access/driver"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
	"github.com/sirupsen/logrus"
)

type BootReq struct {
	Model             driver.Byte16 // 设备型号 String 16 pn设备型号
	Vendor            driver.Byte16 // 供应商 String 16 vendor供应商代码
	CNum              uint8         // 枪头数量 BIN 1 充电桩枪头数量 6 最小充电电流 BIN 1 充电电流下限值（A） 7 最大充电电流 BIN 1 充电电流上限值（A） 8 软件版本号 BIN 1 例如：1.1 ，11.2 的 10 倍
	MinCurrent        uint8         // 最小充电电流 BIN 1 充电电流下限值（A）
	MaxCurrent        uint8         // 最大充电电流 BIN 1 充电电流上限值（A）
	FirmwareVersion   uint16        // 硬件版本号 BIN 1 例如：1.1 ，11.2 的 10 倍
	BTVersion         uint8         // 蓝牙软件版本 BIN 1 例如：1.1 ，11.2 的 10 倍
	BTMac             driver.Byte20 // 蓝牙MAX地址 String 20 蓝牙max地址
	TotalChargeNum    uint32        // 累计充电次数 BIN 4 每充一次累加 1
	TotalExceptionNum uint32        // 累计故障次数 BIN 4 每故障状态变化一次累加 1
	NetWork           uint8         // 充电桩联网方 式 BIN 1 0：WIFI 1：4G
	Standard          uint8         // 充电桩类型 BIN 1 1：美标 2：欧标
	Phase             uint8         // 充电桩相数 BIN 1 1：单相 2：三相
	Rssi              uint8         // 20 4G的信号强度 BIN 1 0-31，越大信号越好
	SIM               driver.Byte20 // 21 4G的SIM卡号 BIN 20 4G的SIM卡号
}

// bootConf * CMD=105	服务器应答充电桩签到命令
type BootConf struct {
	State    uint8  // 6 登录结果 BIN 1 0x00：登录成功 0x01：登录失败
	Template uint32 // 7 对时时钟 BIN 4 时间戳
	TimeZone uint8  // 8 时区 BIN 1 0-11：西时区 12：零时区 13-24：东时区
}

// ToPlatformPayload 转换成平台的Payload
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
	evse.AndroidMac = bn.BTMac.String()
	evse.Vendor = bn.Vendor.String()
	evse.CNum = bn.CNum
	evse.State = commonpb.EvseState_ES_ONLINE
	evse.FirmwareVersion = fmt.Sprintf("%d", bn.FirmwareVersion)
	evse.BTVersion = fmt.Sprintf("%d", bn.BTVersion)
	evse.Standard = getEvseStandard(bn.Standard)
	evse.RatedMinCurrent = int32(bn.MinCurrent)
	evse.RatedMaxCurrent = int32(bn.MaxCurrent)
	evse.NetWork = models.KindNetWork(bn.NetWork)
	evse.Rssi = bn.Rssi
	evse.SIM = bn.SIM.String()

	saves = append(saves, evse)
	if err := gormv2.Saves(context.Background(), saves...); err != nil {
		ctx.Kick = true
		return nil, fmt.Errorf("保存设备信息错误:" + err.Error())
	}
	ctx.Mark = evse.SN
	ctx.Data["retcmd"] = CmdBootConf

	// 加载时区
	zone, _ := models.GetEvseZone(evse.ID)

	_apdu := &APDU{
		Seq: apdu.Seq,
		Cmd: CmdBootConf,
		Payload: &BootConf{
			State:    0,
			Template: uint32(time.Now().Unix()),
			TimeZone: uint8(zone),
		},
	}
	copy(_apdu.SN[:], apdu.SN[:])

	retapdu, err = _apdu.Marshal()

	var evseAutoUpgrade models.EvseAutoUpgrade
	if e := gormv2.Last(context.Background(), &evseAutoUpgrade, "sn=? and upgrade_firmware_version > ?", evse.SN, evse.FirmwareVersion); e != nil {
		return nil, fmt.Errorf("检测强制升级错误:" + e.Error())
	}
	if evseAutoUpgrade.IsExists() {
		models.OTACH <- evseAutoUpgrade
	}

	return
}

// HeartbeatReq 心跳请求
type HeartbeatReq struct {
}

// HeartbeatResp 心跳回复
type HeartbeatConf struct {
}

// GetConfigReq
type GetConfigReq struct {
	ConfNameLen uint16       // 6 配置参数名字 长度 BIN 2 需要配置的参数名字长度，比如配 置“sn”，长度就为2
	ConfName    driver.Bytes // 8 配置参数名字 String 序号6字段的值 配置参数的名字
}

func (rc *GetConfigReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-getConfig", ctx.Mark)
	return nil
}

// GetConfigConf
type GetConfigConf struct {
	ConfNameLen uint16       // 6 配置参数名字 长度 BIN 2 需要配置的参数名字长度，比如配 置“sn”，长度就为2
	ConfValLen  uint16       // 7 配置参数值长 度 BIN 2 对应参数值的长度
	ConfName    driver.Bytes `len_inx:"-2"` // 8 配置参数名字 String 序号6字段的值 配置参数的名字
	ConfVal     driver.Bytes `len_inx:"-2"` // 9 配置参数值 String 序号7字段的值 配置参数的值
}

// ToPlatformPayload 转换成平台的Payload
func (rc *GetConfigConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-getConfig", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

// SetConfigReq 心跳请求
type SetConfigReq struct {
	ConfNameLen uint16       // 6 配置参数名字 长度 BIN 2 需要配置的参数名字长度，比如配 置“sn”，长度就为2
	ConfValLen  uint16       // 7 配置参数值长 度 BIN 2 对应参数值的长度
	ConfName    driver.Bytes // 8 配置参数名字 String 序号6字段的值 配置参数的名字
	ConfVal     driver.Bytes // 9 配置参数值 String 序号7字段的值 配置参数的值
}

func (rc *SetConfigReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-setConfig", ctx.Mark)
	return nil
}

// SetConfigConf
type SetConfigConf struct {
	Status uint8 // 6 配置结果 BIN 1 0x00：成功 0x01：失败
}

// ToPlatformPayload 转换成平台的Payload
func (rc *SetConfigConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	ctx.Log.Data["retcode"] = rc.Status
	sess := itransac.LoadSession(fmt.Sprintf("%s-setConfig", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

// ToPlatformPayload 转换成平台的Payload
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
	UserType uint8  //6	User	BIN	1	0x00：桩主其他：分享者
	UserID   uint32 //7	UserId	BIN	4	用户 ID：1000008
}

func (rc *GetReserverInfoReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-GetReserverInfo", ctx.Mark)
	return nil
}

type ReserverInfo struct {
	Repeat     uint8  // 8 预约标志 BIN 1 Bit7代表单次，bit0到bit6代表周 一到周日
	StartTime  uint32 // 9 开始时间 BIN 4 重复的表示每天的第几秒；单次表 示时间戳
	ChargeTime uint16 // 10 充电时间 BIN 2 充电时长，分钟
}

type GetReserverInfoConf struct {
	UserType      uint8          // User	BIN	1	0x00：桩主其他：分享者
	UserID        uint32         // UserId	BIN	4	用户 ID：1000008
	ReserverInfos []ReserverInfo //
}

// ToPlatformPayload 转换成平台的Payload
func (rc *GetReserverInfoConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-GetReserverInfo", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type GetWhitelistReq struct {
	UserType uint8  // User	BIN	1	0x00：桩主其他：分享者
	UserID   uint32 // UserId	BIN	4	用户 ID：1000008
}

func (rc *GetWhitelistReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-GetWhitelist", ctx.Mark)
	return nil
}

type GetWhitelistConf struct {
	TNum  uint8           //  当前充电桩已经配置的白名单卡 号数量
	Cards []driver.Byte16 //
}

// ToPlatformPayload 转换成平台的Payload
func (rc *GetWhitelistConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-GetWhitelist", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type SetWhitelistReq struct {
	Func uint8         // 功能码 BIN 1 0：增加卡号 1：删除卡号
	Card driver.Byte16 // 卡号 String 16 卡号
}

func (rc *SetWhitelistReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-SetWhitelist", ctx.Mark)
	return nil
}

type SetWhitelistConf struct {
	Status uint8 // 6 状态 BIN 1 0：成功 1：失败
}

// ToPlatformPayload 转换成平台的Payload
func (rc *SetWhitelistConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-SetWhitelist", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type SetReserverInfoReq struct {
	UserType      uint8          // User	BIN	1	0x00：桩主其他：分享者
	UserID        uint32         // UserId	BIN	4	用户 ID：1000008
	ReserverInfos []ReserverInfo //
}

func (rc *SetReserverInfoReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-SetReserverInfo", ctx.Mark)
	return nil
}

type SetReserverInfoConf struct {
	Status uint8 // 启动应答	BIN	1	0：成功； 1：失败
}

// ToPlatformPayload 转换成平台的Payload
func (rc *SetReserverInfoConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-SetReserverInfo", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type SetWorkModeReq struct {
	UserType uint8  // User	BIN	1	0x00：桩主其他：分享者
	UserID   uint32 // UserId	BIN	4	用户 ID：1000008
	WorkMode uint8  // 8 即插即充模式 BIN 1 0：取消即插即充（APP模式） 1：使能即插即充
}

func (rc *SetWorkModeReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-SetWorkMode", ctx.Mark)
	return nil
}

type SetWorkModeConf struct {
	Status uint8 // 6 状态 BIN 1 0：成功 1：错误
}

// ToPlatformPayload 转换成平台的Payload
func (rc *SetWorkModeConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	sess := itransac.LoadSession(fmt.Sprintf("%s-SetWorkMode", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type RemoteCtrlReq struct {
	UserType        uint8  //6	User	BIN	1	0x00：桩主其他：分享者
	UserID          uint32 //7	UserId	BIN	4	用户 ID：1000008
	ChargingCurrent uint8  //8	充电电流	BIN	1	充电电流大小（A）
	Command         uint8  //9	启/停命令	BIN	1	1：立即启动 2：立即停止 其他无效；
}

// ToDevicePayload 转换成单车的Payload
func (rc *RemoteCtrlReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-remotecontrol", ctx.Mark)
	return nil
}

type RemoteCtrlConf struct {
	Status    uint8  // 启动应答	BIN	1	0：操作失败； 1：启动成功； 2：停止成功； 其他无效；
	StartTime uint32 // 启动时间	BIN	4	时间戳
	Meter     uint32 // 累加充电电量	BIN	4	0.01kW·h
}

// ToPlatformPayload 转换成平台的Payload
func (rc *RemoteCtrlConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	ctx.Log.Data["retcode"] = rc.Status
	sess := itransac.LoadSession(fmt.Sprintf("%s-remotecontrol", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type TriggerTelemetryReq struct {
	UserType uint8  // 6 User BIN 1 0x00：桩主 0x01-0x07：分享者
	UserID   uint32 // 7 UserId BIN 4 用户 ID：1000008
}

// ToDevicePayload 转换成单车的Payload
func (rc *TriggerTelemetryReq) ToDevicePayload(ctx *itransac.Ctx) error {

	return nil
}

type TelemetryReq struct {
	Status           uint8         // 充电桩状态	BIN	1	[bit0]充电桩故障状态 0：正常1：故障 [bit1]充电枪连接状态 0：未连接1：已连接 [bit2-3]充电桩充电状态 0：空闲 1：等待车辆启动中 2：充电中 3：充电停止 [bit4]充电桩握手状态 0：未握手1：已握手 [bit5]充电桩解锁状态 0：未解锁1：已解锁 [bit6]预约状态：0：未预约，1： 已预约 [bit7]保留，置 0
	FaultCode        uint8         // 故障码	BIN	1	0：无故障 其他：对应故障代码
	SetOutputCurrent uint16        // 当前设置输出电流
	WorkMode         uint8         // 启动模式， 1，即插即充， 其他: 需授权启动充电模式
	Voltage          uint16        // 充电电压	BIN	2	分辨率：0.1v
	Current          uint16        // 充电电流	BIN	2	分辨率：0.1A
	Power            uint16        // 充电功率	BIN	2	分辨率：0.01kW
	ConsumedElectric uint32        // 本次充电电量	BIN	4	分辨率：0.001kW·h
	Meter            uint32        // 累计充电电量	BIN	4	分辨率：0.01kW·h
	ChargingTime     uint16        // 本次充电时长	BIN	2	分
	AuthMode         uint8         // 授权模式 BIN 1  授权模式
	RecordID         driver.Byte32 // 充电流水号	String	32	桩端本地生成的流水号
	VoltageB         uint16        // B相充电电压 BIN 2 分辨率：0.1v
	CurrentB         uint16        // B相充电电流 BIN 2 分辨率：0.1A
	VoltageC         uint16        // C相充电电压 BIN 2 分辨率：0.1v
	CurrentC         uint16        // C相充电电流 BIN 2 分辨率：0.1A
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
	connect.CurrentB = uint32(rc.CurrentB)
	connect.CurrentC = uint32(rc.CurrentC)
	connect.VoltageA = uint32(rc.Voltage)
	connect.VoltageB = uint32(rc.VoltageB)
	connect.VoltageC = uint32(rc.VoltageC)
	connect.ConsumedElectric = uint32(rc.ConsumedElectric)
	connect.ChargingTime = rc.ChargingTime
	connect.FaultCode = uint16(rc.FaultCode)

	if err = gormv2.Saves(context.Background(), connect); err != nil {
		return
	}

	return
}

type TransctionReq struct {
	UserID           uint32        // UserId	BIN	4	用户 ID：1000008
	AuthMode         uint8         // 启动授权模式
	RecordID         driver.Byte32 // 充电流水号	String	32	桩端本地生成的流水号
	StartTime        uint32        // 充电开始时间	BIN	4	时间戳
	ChargeTime       uint32        // 充电时长	BIN	4	单位：分钟
	TotalElectricity uint32        // 本次充电电量	BIN	4	分辨率：0.001kW·h
	Meter            uint32        // 累计充电电量	BIN	4	分辨率：0.01kW·h
	StopReason       uint8         // 充电停止原因	BIN	1	0：自动停充1：APP 停充 2：故障停充
	FaultCode        uint8         // 故障码	BIN	1	0：无故障其他：对应故障代码

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
	Status     uint8 // 启动应答	BIN	1	0：成功；  其他失败；
}

// ToDevicePayload 转换成单车的Payload
func (rc *UpdateFirmwareReq) ToDevicePayload(ctx *itransac.Ctx) error {
	ctx.WaitRet = true
	ctx.WaitKey = fmt.Sprintf("%s-UpdateFirmware", ctx.Mark)
	return nil
}

// ToPlatformPayload 转换成平台的Payload
func (rc *UpdateFirmwareConf) ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error) {
	ctx.Log.Data["retcode"] = rc.Status
	sess := itransac.LoadSession(fmt.Sprintf("%s-UpdateFirmware", ctx.Mark))
	if sess != nil {
		sess.CH <- rc
	}
	return nil, nil
}

type TransctionConf struct {
	RecordID driver.Byte32 //6	充电流水号	String	32	桩端本地生成的流水号
	State    uint8         //7	上传结果	BIN	1	0x00：成功 0x01：失败
}

// APDU 电单车 数据包结构和定义
type APDU struct {
	Head    uint8         // 1	起始标志	BCD 码	1	固定码：0x68
	Length  uint8         // 2	数据长度	BIN	1
	Seq     uint8         // 3	序列号域	BIN	1	0-255
	Cmd     Cmd           // 4	帧类型标志	BCD 码	1	固定码：0x01
	SN      driver.Byte16 // 5	桩 SN	String	16	AC40055（不足补0）
	Payload interface{}   //
	Check   uint8         // 12	帧检验域	BIN	1	累加和校验
}

// NewAPDU 创建一个单车通讯数据包结构
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

// FromAPDU
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

// Unmarshal 从字节流中解析出apdu
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

// Marshal 将apdu打包成字节流
func (a *APDU) Marshal() (buf []byte, err error) {
	buf, err = codec.Marshal(a)
	if err != nil {
		return nil, err
	}
	buf[0] = 0x68
	buf[1] = uint8(len(buf) - 2)
	return addSum(buf), nil
}
