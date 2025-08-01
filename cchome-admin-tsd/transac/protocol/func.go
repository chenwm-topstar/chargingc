package protocol

import "github.com/chenwm-topstar/pbs/commonpb"

func getConnectState(s uint8) commonpb.ConnectorState {
	if s&1 == 1 {
		return commonpb.ConnectorState_CS_Faulted // 故障状态
	}
	if s>>1&1 == 0 { // 空闲状态
		return commonpb.ConnectorState_CS_Available
	}
	switch s >> 2 & 3 {
	case 0:
		return commonpb.ConnectorState_CS_Preparing // 已插枪
	case 1:
		return commonpb.ConnectorState_CS_SuspendedEV // 充电已开启，电动汽车还未充电
	case 2:
		return commonpb.ConnectorState_CS_Charging // 充电中
	case 3:
		return commonpb.ConnectorState_CS_Finishing // 充电完成
	}
	if s>>6&1 == 1 {
		return commonpb.ConnectorState_CS_Waiting // 等待
	}

	return commonpb.ConnectorState_CS_Unavailable // 不可用
}

func getEvseStandard(s uint8) commonpb.EvseStandard {
	switch s {
	case 1:
		return commonpb.EvseStandard_ES_AMERICAN
	case 2:
		return commonpb.EvseStandard_ES_EUROPEAN
	}
	return commonpb.EvseStandard_ES_UNKNOWN
}

func getEvsePhase(s uint8) commonpb.EvsePhase {
	switch s {
	case 1:
		return commonpb.EvsePhase_EP_ONE
	case 2:
		return commonpb.EvsePhase_EP_THREE
	}
	return commonpb.EvsePhase_EP_UNKNOWN
}

// checkSum 校验数据
func checkSum(buf []byte) bool {
	l, sum := len(buf), uint8(0)
	for i := 2; i < l-1; i++ {
		sum += uint8(buf[i])
	}

	return sum == uint8(buf[l-1])
}

// addEvsePacketSum
func addSum(buf []byte) []byte {
	l, sum := len(buf), uint8(0)
	for i := 2; i < l-1; i++ {
		sum += uint8(buf[i])
	}
	buf[l-1] = sum
	return buf
}
