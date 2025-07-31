package tcp

import (
	"fmt"
	"io"
	"log"
	"net"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/transac/itransac"
	gp "github.com/chenwm-topstar/chargingc/cchome-admin-topstar/transac/protocol"
	"github.com/funny/link"
	"github.com/funny/slab"
	"github.com/sirupsen/logrus"
)

const connBuckets = 32

type ACCfg struct {
	BufferSize   int
	SendChanSize int
	IdleTimeout  time.Duration
}

type TMAC struct {
	protocol
	servers      *link.Server
	evseSessions sync.Map

	disconnectCallback func(string, string)
}

func NewAC(maxPacketSize int, disconnectCallback func(string, string)) *TMAC {
	tmac := &TMAC{}
	tmac.pool = slab.NewSyncPool(64, 64*1024, 4)
	tmac.maxPacketSize = maxPacketSize
	tmac.disconnectCallback = disconnectCallback

	return tmac
}

func (tmac *TMAC) ServeClients(lsn net.Listener, cfg ACCfg) {
	tmac.servers = link.NewServer(
		lsn,
		link.ProtocolFunc(func(rw io.ReadWriter) (link.Codec, error) {
			return tmac.newCodec(rw.(net.Conn), cfg.BufferSize), nil
		}),
		cfg.SendChanSize,
		link.HandlerFunc(func(session *link.Session) {
			tmac.handleSession(session, cfg.IdleTimeout)
		}),
	)

	tmac.servers.Serve()
}

func (tmac *TMAC) Run(address string) {
	lsn, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Sprintf("listener at %s failed - %s", address, err))
	}
	logrus.Infof("listener %s start...", address)
	go tmac.ServeClients(lsn, ACCfg{BufferSize: 4096, SendChanSize: 1024, IdleTimeout: 10 * time.Minute})

	return
}

func (tmac *TMAC) Stop() {
	tmac.servers.Stop()
}

func (tmac *TMAC) CheckOnline(evseID string) (ok bool) {

	_, ok = tmac.evseSessions.Load(evseID)
	return
}

func (tmac *TMAC) addSessionMapping(session *link.Session, mark string) (err error) {

	if sid, ok := tmac.evseSessions.Load(mark); ok && sid.(uint64) != session.ID() {

		logrus.Warnf("evse:[%s] 被挤下线", mark)

		tmac.delSessionMapping(sid.(uint64), mark, "被挤下线", false)
		if oldsession := tmac.servers.GetSession(sid.(uint64)); oldsession != nil {
			oldsession.Close()
		}
	}
	session.Codec().(*codec).mark = mark
	tmac.evseSessions.Store(mark, session.ID())
	return
}

func (tmac *TMAC) delSessionMapping(sid uint64, mark, reason string, callback bool) {

	_sid, ok := tmac.evseSessions.Load(mark)
	if ok && sid == _sid {
		tmac.evseSessions.Delete(mark)
		if tmac.disconnectCallback != nil && callback {
			tmac.disconnectCallback(mark, reason)
		}
	}
}

func (tmac *TMAC) getSessionByMapping(evseID string) (*link.Session, error) {

	if sid, ok := tmac.evseSessions.Load(evseID); ok {

		if session := tmac.servers.GetSession(sid.(uint64)); session != nil {
			if session.IsClosed() {
				return nil, fmt.Errorf("evseid %v session is close", evseID)
			}
			return session, nil
		}
		return nil, fmt.Errorf("evseid %v session is nil", evseID)
	}
	return nil, fmt.Errorf("evseid %v session not found", evseID)
}

func (tmac *TMAC) Disconnector(evseid, reason string) {
	if session, _ := tmac.getSessionByMapping(evseid); session != nil {
		tmac.delSessionMapping(session.ID(), evseid, reason, true)
		session.Close()
	}
}

func (tmac *TMAC) Send(evseid string, buf []byte) error {
	session, err := tmac.getSessionByMapping(evseid)
	if err != nil {
		return err
	}
	return session.Send(buf)
}

func (tmac *TMAC) handleSession(session *link.Session, idleTimeout time.Duration) {
	var err error
	conn := session.Codec().(*codec).conn

	logrus.Infof("remote addr:[%+v]", conn.RemoteAddr())
	defer func() {
		mark := session.Codec().(*codec).mark
		logrus.Infof("disconnect addr:[%+v]  error:[%+v] evse:[%s]", conn.RemoteAddr(), err, mark)
		reason := "kick"
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			if reason == "EOF" {
				reason = "offline"
			} else if strings.Contains(err.Error(), "timeout") {
				reason = "timeout"
			} else if strings.Contains(err.Error(), "connection reset by peer") {
				reason = "offline"
			}
		}
		tmac.delSessionMapping(session.ID(), mark, reason, true)
		if !session.IsClosed() {
			session.Close()
		}
		if err := recover(); err != nil {
			log.Printf("ac panic: %v\n%s", err, debug.Stack())
		}
	}()

	for {
		if idleTimeout > 0 {
			err = conn.SetReadDeadline(time.Now().Add(idleTimeout))
			if err != nil {
				return
			}
		}

		var buf interface{}
		if buf, err = session.Receive(); err != nil {
			logrus.Error("receive error:" + err.Error())
			return
		}
		go func(session *link.Session, buf *[]byte) {
			ctx := &itransac.Ctx{
				Mark: session.Codec().(*codec).mark,
				Raw:  *buf,
				Data: make(map[string]interface{}),
			}
			ctx.Data["ac"] = tmac

			apdu := &gp.APDU{}
			ret, err := apdu.ToAPDU(ctx)
			if tmp, ok := ctx.Data["keepalive"]; ok {
				idleTimeout = tmp.(time.Duration)
			}
			if err != nil {
				ctx.Log.Errorf("ToAPDU error, err:%s", err.Error())
				if ctx.Kick {
					session.Close()
				}
				return
			} else if ret == nil || len(ret) < 0 {
				return
			}

			if session.Codec().(*codec).mark == "" && ctx.Mark != "" {
				if err := tmac.addSessionMapping(session, ctx.Mark); err != nil {
					ctx.Log.Errorf(err.Error())
					session.Close()
				}
			} else {
				if ok := tmac.CheckOnline(ctx.Mark); !ok {
					ctx.Log.Errorf("设备[%s]未添加到映射列表, 踢掉重新连接", ctx.Mark)
					session.Close()
				}
			}

			if err := session.Send(ret); err != nil {
				ctx.Log.Errorf("ret send error:%s", err.Error())
				return
			}
		}(session, buf.(*[]byte))
	}
}
