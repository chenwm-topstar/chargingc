package evsectl

import (
	"fmt"

	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/internal/appproto"
	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/transac"
	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/transac/protocol"
	"golang.org/x/sync/singleflight"
)

var sg singleflight.Group

func StartCharger(sn string, userID uint32, chargingCurrent int32) error {
	key := sn + ":startcharger"
	_, err, _ := sg.Do(key, func() (interface{}, error) {
		rc := &protocol.RemoteCtrlReq{
			UserID:          userID,
			ChargingCurrent: uint8(chargingCurrent),
			Command:         1,
		}
		ret, err := transac.Send(sn, protocol.CmdRemoteCtrlReq, rc)
		if err != nil {
			return nil, err
		}
		if v, ok := ret.(*protocol.RemoteCtrlConf); !ok || v.Status != 1 {
			return nil, fmt.Errorf("start charger error. error[%+v][%+v]", ok, v.Status)
		}
		return nil, nil
	})

	return err
}

func StopCharger(sn string) error {
	rc := &protocol.RemoteCtrlReq{
		Command: 2,
	}
	ret, err := transac.Send(sn, protocol.CmdRemoteCtrlReq, rc)
	if err != nil {
		return err
	}
	if v, ok := ret.(*protocol.RemoteCtrlConf); !ok || v.Status != 2 {
		return fmt.Errorf("stop charger error. error[%+v][%+v]", ok, v.Status)
	}
	return nil
}

func Reset(sn string) error {
	return SetConfig(sn, "recovery", "")
}

func SetCurrent(sn string, current int) error {
	return SetConfig(sn, "Power", fmt.Sprintf("%d", current))
}

func TriggerTelemetry(sn string) error {
	req := &protocol.TriggerTelemetryReq{}
	_, err := transac.Send(sn, protocol.CmdTriggerTelemeteryReq, req)
	if err != nil {
		return err
	}
	return nil
}

func Upgrade(sn, FTPAddr string) error {
	req := &protocol.UpdateFirmwareReq{}
	copy(req.FTPAddress[:], []byte(FTPAddr))
	ret, err := transac.Send(sn, protocol.CmdOTAReq, req)
	if err != nil {
		return err
	}
	if v, ok := ret.(*protocol.UpdateFirmwareConf); !ok || v.Status != 0 {
		return fmt.Errorf("upgrade error. error[%+v][%+v]", ok, v.Status)
	}
	return nil
}

func SetConfig(sn string, name, val string) error {
	rc := &protocol.SetConfigReq{
		ConfNameLen: uint16(len(name)),
		ConfValLen:  uint16(len(val)),
	}
	rc.ConfName = append(rc.ConfName, []byte(name)...)
	rc.ConfVal = append(rc.ConfVal, []byte(val)...)

	ret, err := transac.Send(sn, protocol.CmdSetConfigReq, rc)
	if err != nil {
		return err
	}
	v, ok := ret.(*protocol.SetConfigConf)
	if !ok || v.Status != 0 {
		return fmt.Errorf("set config %s error. error[%+v][%+v]", name, ok, v.Status)
	}
	return nil
}

func GetConfig(sn string, name string) (string, error) {
	rc := &protocol.GetConfigReq{
		ConfNameLen: uint16(len(name)),
	}
	rc.ConfName = append(rc.ConfName, []byte(name)...)

	ret, err := transac.Send(sn, protocol.CmdGetConfigReq, rc)
	if err != nil {
		return "", err
	}
	v, ok := ret.(*protocol.GetConfigConf)
	if !ok {
		return "", fmt.Errorf("set charger current error. error[%+v]", ok)
	}
	return v.ConfVal.String(), nil
}

func GetReserverInfo(sn string, userID uint32) (ris []appproto.ReserverInfo, err error) {
	rc := &protocol.GetReserverInfoReq{
		UserType: 0,
		UserID:   userID,
	}

	ret, err := transac.Send(sn, protocol.CmdGetReserverReq, rc)
	if err != nil {
		return nil, err
	}
	v, ok := ret.(*protocol.GetReserverInfoConf)
	if !ok {
		return nil, fmt.Errorf("set charger current error. error[%+v]", ok)
	}
	for _, ri := range v.ReserverInfos {
		ris = append(ris, appproto.ReserverInfo{
			Flag:       ri.Repeat,
			StartTime:  ri.StartTime,
			ChargeTime: ri.ChargeTime,
		})
	}
	return
}

func SetReserverInfo(sn string, userID uint32, ris []appproto.ReserverInfo) (err error) {
	sri := &protocol.SetReserverInfoReq{
		UserType: 0,
		UserID:   userID,
	}
	for _, ri := range ris {
		sri.ReserverInfos = append(sri.ReserverInfos, protocol.ReserverInfo{
			Repeat:     ri.Flag,
			StartTime:  ri.StartTime,
			ChargeTime: ri.ChargeTime,
		})
	}

	ret, err := transac.Send(sn, protocol.CmdSetReserverReq, sri)
	if err != nil {
		return err
	}
	v, ok := ret.(*protocol.SetReserverInfoConf)
	if !ok || v.Status != 0 {
		return fmt.Errorf("set reserver error[%+v]", ok)
	}

	return
}

func GetWhitelistCard(sn string, userID uint32) (cards []string, err error) {
	rc := &protocol.GetWhitelistReq{
		UserType: 0,
		UserID:   userID,
	}

	ret, err := transac.Send(sn, protocol.CmdGetWhitelistReq, rc)
	if err != nil {
		return nil, err
	}
	v, ok := ret.(*protocol.GetWhitelistConf)
	if !ok {
		return nil, fmt.Errorf("set charger current error. error[%+v]", ok)
	}
	for _, card := range v.Cards {
		cards = append(cards, card.String())
	}
	return
}

func SetWhitelistCard(sn string, userID uint32, del bool, card string) (err error) {
	wl := &protocol.SetWhitelistReq{}
	if del {
		wl.Func = 1
	}
	copy(wl.Card[:], []byte(card))

	ret, err := transac.Send(sn, protocol.CmdSetWhitelistReq, wl)
	if err != nil {
		return err
	}
	v, ok := ret.(*protocol.SetWhitelistConf)
	if !ok || v.Status != 0 {
		return fmt.Errorf("set whitelist error[%+v]", v.Status)
	}

	return
}

func SetWorkMode(sn string, userID uint32, workmode uint8) (err error) {
	wm := &protocol.SetWorkModeReq{
		UserID:   userID,
		WorkMode: workmode,
	}

	ret, err := transac.Send(sn, protocol.CmdSetWorkModeReq, wm)
	if err != nil {
		return err
	}
	v, ok := ret.(*protocol.SetWorkModeConf)
	if !ok || v.Status != 0 {
		return fmt.Errorf("set SetWorkMode error[%+v]", ok)
	}

	return
}
