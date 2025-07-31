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

	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/transac/itransac"
	gp "github.com/chenwm-topstar/chargingc/cchome-admin-tsd/transac/protocol"
	"github.com/funny/link"
	"github.com/funny/slab"
	"github.com/sirupsen/logrus"
)

const connBuckets = 32

// ACCfg 配置
type ACCfg struct {
	BufferSize   int
	SendChanSize int
	IdleTimeout  time.Duration
}

// TMAC tcp实现ac通讯层,协议层交给具体的ac做
type TMAC struct {
	protocol
	servers      *link.Server // 服务于桩
	evseSessions sync.Map
	// sessionLock        sync.RWMutex
	// evseSessions       map[string]uint64
	disconnectCallback func(string, string) // 断开回调函数
}

// NewAC 创建一个ac.
func NewAC(maxPacketSize int, disconnectCallback func(string, string)) *TMAC {
	tmac := &TMAC{}
	tmac.pool = slab.NewSyncPool(64, 64*1024, 4)
	tmac.maxPacketSize = maxPacketSize
	tmac.disconnectCallback = disconnectCallback
	// tmac.evseSessions = make(map[string]uint64)
	return tmac
}

// ServeClients 服务于后台桩
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

// Run 启动ac
func (tmac *TMAC) Run(address string) {
	lsn, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Sprintf("listener at %s failed - %s", address, err))
	}
	logrus.Infof("listener %s start...", address)
	go tmac.ServeClients(lsn, ACCfg{BufferSize: 4096, SendChanSize: 1024, IdleTimeout: 10 * time.Minute})
	// go func() {
	// 	select {
	// 	case <-time.After(10 * time.Minute):
	// 		logrus.Infof("total session num:%d", len(tmac.evseSessions))
	// 	}
	// }()
	return
}

// Stop
func (tmac *TMAC) Stop() {
	tmac.servers.Stop()
}

// CheckOnline 检测是否在线
func (tmac *TMAC) CheckOnline(evseID string) (ok bool) {
	// tmac.sessionLock.RLock()
	// _, ok = tmac.evseSessions[evseID]
	// tmac.sessionLock.RUnlock()
	_, ok = tmac.evseSessions.Load(evseID)
	return
}

// func (tmac *TMAC) SearchSN(sn string) (evseids []string) {
// 	// tmac.evseSessions.Range(func(k, v interface{}) bool {
// 	// 	kvs := strings.Split(k.(string), ":")
// 	// 	if len(kvs) != 3 {
// 	// 		logrus.Errorf("保存的设备回话映射错误. k:[%v], v:[%v]", k, v)
// 	// 		return false
// 	// 	} else if kvs[2] == sn {
// 	// 		evseids = append(evseids, k.(string))
// 	// 	}
// 	// 	return true
// 	// })
// 	return
// }

// func (tmac *TMAC) GetOnlineDevices(offset, limit int, sn string) (evseids []string, total int) {
// 	// tmac.evseSessions.Range(func(k, v interface{}) bool {
// 	// 	if sn != "" {
// 	// 		kvs := strings.Split(k.(string), ":")
// 	// 		if len(kvs) != 3 {
// 	// 			logrus.Errorf("保存的设备回话映射错误. k:[%v], v:[%v]", k, v)
// 	// 			return false
// 	// 		} else if kvs[2] == sn || kvs[1] == sn {
// 	// 			evseids = append(evseids, k.(string))
// 	// 		}
// 	// 	} else {
// 	// 		evseids = append(evseids, k.(string))
// 	// 	}
// 	// 	return true
// 	// })
// 	// l := len(evseids)
// 	// if l == 0 {
// 	// 	return
// 	// }
// 	// sort.Strings(evseids)
// 	// end := offset + limit
// 	// if end > l {
// 	// 	end = l
// 	// }
// 	return evseids[offset:end], l
// }

func (tmac *TMAC) addSessionMapping(session *link.Session, mark string) (err error) {

	if sid, ok := tmac.evseSessions.Load(mark); ok && sid.(uint64) != session.ID() {
		// return fmt.Errorf("设备[%s]已存在，重复连接", evseID)
		logrus.Warnf("evse:[%s] 被挤下线", mark)

		// 如果sn信息已经存在则踢掉老连接, 保留新连接
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
	// tmac.sessionLock.Lock()
	// defer tmac.sessionLock.Unlock()

	// if _sid, ok := tmac.evseSessions[evseID]; ok && sid == _sid {
	// 	delete(tmac.evseSessions, evseID)
	// 	if tmac.disconnectCallback != nil && callback {
	// 		tmac.disconnectCallback(evseID, reason)
	// 	}
	// }

	_sid, ok := tmac.evseSessions.Load(mark)
	if ok && sid == _sid {
		tmac.evseSessions.Delete(mark)
		if tmac.disconnectCallback != nil && callback {
			tmac.disconnectCallback(mark, reason)
		}
	}
}

func (tmac *TMAC) getSessionByMapping(evseID string) (*link.Session, error) {
	// tmac.sessionLock.RLock()
	// defer tmac.sessionLock.RUnlock()

	if sid, ok := tmac.evseSessions.Load(evseID); ok {
		// if sid, ok := tmac.evseSessions[evseID]; ok {
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

// Disconnector 离线
func (tmac *TMAC) Disconnector(evseid, reason string) {
	if session, _ := tmac.getSessionByMapping(evseid); session != nil {
		tmac.delSessionMapping(session.ID(), evseid, reason, true)
		session.Close()
	}
}

// Send 发送数据
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
