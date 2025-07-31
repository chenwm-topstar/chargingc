package transac

import (
	"encoding/binary"
	"time"

	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/models"
	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/transac/itransac"
	gp "github.com/chenwm-topstar/chargingc/cchome-admin-topstar/transac/protocol"
	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/transac/tcp"
	"github.com/chenwm-topstar/chargingc/utils/access/codec"
	"github.com/sirupsen/logrus"
)

var tcpac *tcp.TMAC

func Run(addr string) (err error) {
	tcp.SetEndian(binary.BigEndian)
	codec.SetEndian(binary.BigEndian)
	tcp.SetLenFieldIndex(0, 2)
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

func CheckOnline(evseID string) (ok bool) {
	return tcpac.CheckOnline(evseID)
}

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
