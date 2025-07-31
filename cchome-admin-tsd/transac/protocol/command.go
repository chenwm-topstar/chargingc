package protocol

import (
	"fmt"
)

type Cmd uint8

const (
	CmdBootReq              Cmd = 0x02 // 登录请求
	CmdBootConf             Cmd = 0x01 // 登录应答
	CmdRemoteCtrlReq        Cmd = 0x03 // 充电授权下发 APP->充电桩 APP 启动/停止充电授权
	CmdRemoteCtrlConf       Cmd = 0x04 // 充电授权应答 充电桩->APP 充电桩应答
	CmdTimingReq            Cmd = 0x05 // 预约设置下发 APP->充电桩 APP 配置信息下发
	CmdTimingConf           Cmd = 0x06 // 预约设置应答 充电桩->APP 充电桩应答
	CmdTriggerTelemeteryReq Cmd = 0x07 // 实时状态数据请求 APP->充电桩 APP 请求状态读取
	CmdTelemetryReq         Cmd = 0x08 // 实时状态数据应答 充电桩->APP 充电桩应答
	CmdGetRecordReq         Cmd = 0x10 // 充电/故障记录读取 APP->充电桩 APP 主动请求读取记录
	CmdGetRecordConf        Cmd = 0x11 // 充电/故障记录应答 充电桩->APP 充电桩应答请求
	CmdGetLogReq            Cmd = 0x12 // 日志记录上传请求 充电桩->APP 充电桩响应上传记录
	CmdGetLogConf           Cmd = 0x13 // 日志记录上传应答 APP->充电桩 APP 应答上传记录
	CmdLogFinishNoitfyReq   Cmd = 0x14 // 日志记录下载完成应答 充电桩->APP 充电桩上传完记录后请求完成
	CmdLogFinishNoitfyConf  Cmd = 0x15 // 日志记录下载完成请求 APP->充电桩 APP 响应请求完成
	CmdOTAReq               Cmd = 0x20 // 升级请求 APP->充电桩 APP 主动请求系统升级
	CmdOTAConf              Cmd = 0x21 // 升级应答 充电桩->APP 充电桩响应升级请求
	CmdHeartbeatConf        Cmd = 0x32 // 心跳应答
	CmdHeartbeatReq         Cmd = 0x33 // 心跳请求
	CmdSetConfigReq         Cmd = 0x54 // 参数配置请求 APP->充电桩 APP请求设置
	CmdSetConfigConf        Cmd = 0x55 // 参数配置请求应答 充电桩->APP 充电桩响应设置
	CmdGetConfigReq         Cmd = 0x58 // 参数查询请求 APP->充电桩 APP请求查询
	CmdGetConfigConf        Cmd = 0x59 // 参数查询请求应答 充电桩->APP 充电桩响应查询
	CmdSetReserverReq       Cmd = 0x62 // 预约计划下发 APP->充电桩 APP下发预约计划
	CmdSetReserverConf      Cmd = 0x63 // 预约计划下发应答 充电桩->APP 充电桩响应下发结果
	CmdGetReserverReq       Cmd = 0x64 // 预约计划查询 APP->充电桩 APP查询预约计划
	CmdGetReserverConf      Cmd = 0x65 // 预约计划查询应答 充电桩->APP 充电桩响应预约计划
	CmdTransactionReq       Cmd = 0x90 // 充电记录上送请求
	CmdTransactionConf      Cmd = 0x91 // 充电记录上送应答
	CmdSetWhitelistReq      Cmd = 0x50 // 白名单管理请求 APP->充电桩 APP请求设置
	CmdSetWhitelistConf     Cmd = 0x51 // 白名单管理请求应答 充电桩->APP 充电桩响应设置
	CmdGetWhitelistReq      Cmd = 0x52 // 白名单查询请求 APP->充电桩 APP请求查询
	CmdGetWhitelistConf     Cmd = 0x53 // 白名单查询请求应答 充电桩->APP 充电桩响应查询
	CmdSetWorkModeReq       Cmd = 0x40 // 即插即充设置请求 APP->充电桩 APP请求设置
	CmdSetWorkModeConf      Cmd = 0x41 // 即插即充设置请求应答 充电桩->APP 充电桩响应设置

	// cmdHardshakeReq cmd = 0x22 // 负载均衡总功率设置 APP->充电桩 APP 请求设置
	// cmdHardshakeReq cmd = 0x23 // 负载均衡总功率设置应 答 充电桩->APP 充电桩响应设置
	// cmdHardshakeReq cmd = 0x30 // 配网请求 APP->充电桩 APP发送WIFI信息给桩配网
	// cmdHardshakeReq cmd = 0x31 // 配网应答 充电桩->APP 充电桩响应配网结果
	// cmdHardshakeReq cmd = 0x42 // 即插即充查询请求 APP->充电桩 APP请求查询
	// cmdHardshakeReq cmd = 0x43 // 即插即充查询请求应答 充电桩->APP 充电桩响应查询
	// cmdHardshakeReq cmd = 0x44 // 密码设置请求 APP->充电桩 APP请求设置
	// cmdHardshakeReq cmd = 0x45 // 密码设置请求应答 充电桩->APP 充电桩响应设置
	// cmdHardshakeReq cmd = 0x46 // 密码查询请求 APP->充电桩 APP请求查询
	// cmdHardshakeReq cmd = 0x47 // 密码查询请求应答 充电桩->APP 充电桩响应查询
	// cmdHardshakeReq cmd = 0x56 // 网络状态查询请求 APP->充电桩 APP请求查询
	// cmdHardshakeReq cmd = 0x57 // 网络状态查询请求应答 充电桩->APP 充电桩响应查询
	// cmdHardshakeReq cmd = 0x60 // WIFI列表查询请求 APP->充电桩 APP请求查询
	// cmdHardshakeReq cmd = 0x61 // WIFI列表查询请求应答 充电桩->APP 充电桩响应查询
)

func (c Cmd) Desc() string {
	return fmt.Sprintf("%x", c)
}

func getPayloadByCMD(c Cmd) (interface{}, error) {
	switch c {
	case CmdBootReq:
		return &BootReq{}, nil
	case CmdRemoteCtrlReq:
	case CmdRemoteCtrlConf:
		return &RemoteCtrlConf{}, nil
	case CmdTimingReq:
	case CmdTimingConf:
	case CmdTriggerTelemeteryReq:
	case CmdTelemetryReq:
		return &TelemetryReq{}, nil
	case CmdGetRecordReq:
	case CmdGetRecordConf:
	case CmdGetLogReq:
	case CmdGetLogConf:
	case CmdLogFinishNoitfyReq:
	case CmdLogFinishNoitfyConf:
	case CmdOTAReq:
	case CmdOTAConf:
		return &UpdateFirmwareConf{}, nil
	case CmdHeartbeatReq:
		return &HeartbeatReq{}, nil
	// case CmdHeartbeatConf:
	case CmdSetConfigReq:
	case CmdSetConfigConf:
		return &SetConfigConf{}, nil
	case CmdGetConfigReq:
	case CmdGetConfigConf:
		return &GetConfigConf{}, nil
	case CmdSetReserverReq:
	case CmdSetReserverConf:
		return &SetReserverInfoConf{}, nil
	case CmdGetReserverReq:
	case CmdGetReserverConf:
		return &GetReserverInfoConf{}, nil
	case CmdTransactionReq:
		return &TransctionReq{}, nil
	case CmdTransactionConf:
	case CmdSetWhitelistReq:
	case CmdSetWhitelistConf:
		return &SetWhitelistConf{}, nil
	case CmdGetWhitelistReq:
	case CmdGetWhitelistConf:
		return &GetWhitelistConf{}, nil
	case CmdSetWorkModeConf:
		return &SetWorkModeConf{}, nil
	}
	return nil, fmt.Errorf("cmd %#x not support", c)
}
