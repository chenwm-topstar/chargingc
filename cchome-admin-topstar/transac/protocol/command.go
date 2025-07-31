package protocol

import (
	"fmt"
)

type Cmd uint8

const (
	CmdBootReq              Cmd = 0x02
	CmdBootConf             Cmd = 0x01
	CmdRemoteCtrlReq        Cmd = 0x03
	CmdRemoteCtrlConf       Cmd = 0x04
	CmdTimingReq            Cmd = 0x05
	CmdTimingConf           Cmd = 0x06
	CmdTriggerTelemeteryReq Cmd = 0x07
	CmdTelemetryReq         Cmd = 0x08
	CmdGetRecordReq         Cmd = 0x10
	CmdGetRecordConf        Cmd = 0x11
	CmdGetLogReq            Cmd = 0x12
	CmdGetLogConf           Cmd = 0x13
	CmdLogFinishNoitfyReq   Cmd = 0x14
	CmdLogFinishNoitfyConf  Cmd = 0x15
	CmdOTAReq               Cmd = 0x20
	CmdOTAConf              Cmd = 0x21
	CmdHeartbeatConf        Cmd = 0x32
	CmdHeartbeatReq         Cmd = 0x33
	CmdSetConfigReq         Cmd = 0x54
	CmdSetConfigConf        Cmd = 0x55
	CmdGetConfigReq         Cmd = 0x58
	CmdGetConfigConf        Cmd = 0x59
	CmdSetReserverReq       Cmd = 0x62
	CmdSetReserverConf      Cmd = 0x63
	CmdGetReserverReq       Cmd = 0x64
	CmdGetReserverConf      Cmd = 0x65
	CmdTransactionReq       Cmd = 0x90
	CmdTransactionConf      Cmd = 0x91
	CmdSetWhitelistReq      Cmd = 0x50
	CmdSetWhitelistConf     Cmd = 0x51
	CmdGetWhitelistReq      Cmd = 0x52
	CmdGetWhitelistConf     Cmd = 0x53
	CmdSetWorkModeReq       Cmd = 0x40
	CmdSetWorkModeConf      Cmd = 0x41
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
