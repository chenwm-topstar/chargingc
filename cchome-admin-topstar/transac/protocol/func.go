package protocol

import "github.com/chenwm-topstar/chargingc/pbs/evsepb"

func getConnectState(s uint8) evsepb.ConnectorState {
	if s&1 == 1 {
		return evsepb.ConnectorState_CS_Faulted
	}
	if s>>1&1 == 0 {
		return evsepb.ConnectorState_CS_Available
	}
	switch s >> 2 & 3 {
	case 0:
		return evsepb.ConnectorState_CS_Preparing
	case 1:
		return evsepb.ConnectorState_CS_SuspendedEV
	case 2:
		return evsepb.ConnectorState_CS_Charging
	case 3:
		return evsepb.ConnectorState_CS_Finishing
	}
	if s>>6&1 == 1 {
		return evsepb.ConnectorState_CS_Waiting
	}

	return evsepb.ConnectorState_CS_Unavailable
}

func getEvseStandard(s uint8) evsepb.EvseStandard {
	switch s {
	case 1:
		return evsepb.EvseStandard_ES_AMERICAN
	case 2:
		return evsepb.EvseStandard_ES_EUROPEAN
	}
	return evsepb.EvseStandard_ES_UNKNOWN
}

func getEvsePhase(s uint8) evsepb.EvsePhase {
	switch s {
	case 1:
		return evsepb.EvsePhase_EP_ONE
	case 2:
		return evsepb.EvsePhase_EP_THREE
	}
	return evsepb.EvsePhase_EP_UNKNOWN
}

func checkSum(buf []byte) bool {
	l, sum := len(buf), uint8(0)
	for i := 2; i < l-1; i++ {
		sum += uint8(buf[i])
	}

	return sum == uint8(buf[l-1])
}

func addSum(buf []byte) []byte {
	l, sum := len(buf), uint8(0)
	for i := 2; i < l-1; i++ {
		sum += uint8(buf[i])
	}
	buf[l-1] = sum
	return buf
}
