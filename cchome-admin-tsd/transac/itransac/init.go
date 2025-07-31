package itransac

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Ctx ac上下文
type Ctx struct {
	Raw     interface{}
	Mark    string
	Data    map[string]interface{}
	Log     *logrus.Entry
	WaitRet bool
	WaitKey string
	Kick    bool
}

var sessions sync.Map

// Session ....
type Session struct {
	CH  chan interface{}
	Key string
}

// NewSession 创建
func NewSession(key string) *Session {
	s := &Session{
		CH:  make(chan interface{}, 1),
		Key: key,
	}
	sessions.Store(key, s)
	return s
}

// LoadSession 加载
func LoadSession(key string) *Session {
	if s, ok := sessions.Load(key); ok {
		return s.(*Session)
	}
	return nil
}

// Listen 监听
func (sess *Session) Listen(timeout time.Duration) (ret interface{}, err error) {
	select {
	case <-time.After(timeout):
		err = fmt.Errorf("timeout")
		return
	case ret = <-sess.CH:
		break
	}
	if ret == nil {
		err = fmt.Errorf("session chan had been closed")
		return
	}
	return
}

// Close 关闭通道
func (sess *Session) Close() {
	close(sess.CH)
	sessions.Delete(sess.Key)
}
