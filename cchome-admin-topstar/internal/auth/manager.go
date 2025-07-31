package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/astaxie/beego/context"

	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/models"
	"github.com/chenwm-topstar/chargingc/utils/redigo"
	"github.com/garyburd/redigo/redis"
)

var managers sync.Map

func init() {

}

type Manager struct {
	model                 models.Manager
	siderVersionKeySuffix string
	l                     sync.Mutex
	releaseOnce           sync.Once

	mc cache.Cache
}

func NewManager(name string) (*Manager, error) {
	if ret, ok := managers.Load(name); ok {
		return ret.(*Manager), nil
	}

	model, err := models.GetManagerByName(name)
	if err != nil {
		return nil, err
	}

	_m := &Manager{
		model:                 *model,
		siderVersionKeySuffix: "version:siderbar:" + beego.BConfig.AppName,
		mc:                    cache.NewMemoryCache(),
	}
	managers.Store(name, _m)
	return _m, nil
}

func (m Manager) GetModel() models.Manager {
	return m.model
}

func (m *Manager) Release() {
	m.releaseOnce.Do(func() {
		rd := redigo.GetRedis()
		defer rd.Close()
		rd.Do("del", fmt.Sprintf("%s:*:user:%s", m.GetModel().Name, m.siderVersionKeySuffix))

		managers.Delete(m.model.Name)
	})
}

func (m *Manager) GetSiderbar(ctx *context.Context, refererUrl string) (ret map[string]interface{}, err error) {
	ret = map[string]interface{}{}
	var siderVersion string
	siderVersionNow := fmt.Sprintf("%d", int(time.Now().Unix()))
	rd := redigo.GetRedis()
	defer rd.Close()
	if siderVersion, err = redis.String(rd.Do("get", m.siderVersionKeySuffix)); err != nil {
		if err != redis.ErrNil {
			return
		}
		siderVersion = siderVersionNow
		m.l.Lock()
		defer m.l.Unlock()

		rd.Do("set", m.siderVersionKeySuffix, siderVersionNow)
	}

	key := fmt.Sprintf("%s:%s:user:"+m.siderVersionKeySuffix, m.GetModel().Name, siderVersion)
	if tmp := m.mc.Get(key); tmp == nil {
		if ret["menulist"], ret["navlist"], err = GetSidebar(ctx, m.GetModel(), "/device/online/list", refererUrl); err != nil {
			return
		}
		m.mc.Put(key, map[string]interface{}{
			"menulist": ret["menulist"],
			"navlist":  ret["navlist"],
		}, 24*time.Hour)
	} else {
		val := tmp.(map[string]interface{})
		ret["menulist"] = val["menulist"]
		ret["navlist"] = val["navlist"]
	}
	return
}
