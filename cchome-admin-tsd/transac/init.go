package transac

import (
	"encoding/binary"
	"time"

	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/models"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/transac/itransac"
	gp "github.com/chenwm-topstar/chargingc/cchome-admin-tsd/transac/protocol"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/transac/tcp"
	"github.com/chenwm-topstar/chargingc/utils/abiz/access/codec"
	"github.com/sirupsen/logrus"
)

var tcpac *tcp.TMAC

func Run(addr string) (err error) {
	tcp.SetEndian(binary.BigEndian)   // 设置大小段解析
	codec.SetEndian(binary.BigEndian) // 设置大小段解析
	tcp.SetLenFieldIndex(0, 2)        // 设置包长位置
	tcpac = tcp.NewAC(4096, func(mark, reason string) {
		if mark != "" {
			if err = models.EvseOffine(mark); err != nil {
				logrus.Error("evse offine update state error: " + err.Error())
			}
		}
	})
	tcpac.Run(addr)
	return
}

// CheckOnline 检测是否在线
func CheckOnline(evseID string) (ok bool) {
	return tcpac.CheckOnline(evseID)
}

// // GetOnlineDevices 在线设备evseid
// func GetOnlineDevices(offset, limit int, sn string) (evseids []string, total int) {
// 	evseids, total = tcpac.GetOnlineDevices(offset, limit, sn)
// 	return
// }

// // SearchSN 根据sn去搜索完整的evseid，可能有多个
// func SearchSN(sn string) (evseid []string) {
// 	return tcpac.SearchSN(sn)
// }

func Send(sn string, cmd gp.Cmd, v interface{}) (ret interface{}, err error) {
	ctx := &itransac.Ctx{
		Raw:  v,
		Mark: sn,
		Data: make(map[string]interface{}),
	}

	apdu := &gp.APDU{
		Head:    0,
		Length:  0,
		Seq:     0,
		Cmd:     cmd,
		SN:      [16]byte{},
		Payload: nil,
		Check:   0,
	}
	if buf, err := apdu.FromAPDU(ctx); err != nil {
		return nil, err
	} else if err = tcpac.Send(sn, buf); err != nil {
		return nil, err
	}

	if ctx.WaitRet {
		sess := itransac.NewSession(ctx.WaitKey)
		ret, err = sess.Listen(15 * time.Second)
	}
	return
}
