package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/astaxie/beego/session"
	"github.com/astaxie/beego/validation"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdinternal/auth"
)

type Auth struct {
	Main
	sess session.Store
}

func (c *Auth) Prepare() {
	// 如果是选项卡，就要重定向
	if !c.Ctx.Input.IsPost() && !c.Ctx.Input.IsAjax() && c.Ctx.Input.Query("addtabs") == "" && c.Ctx.Input.Query("ref") == "addtabs" {
		_url, _ := url.Parse(c.Ctx.Request.RequestURI)

		query := _url.Query()
		query.Del("ref")
		u := _url.Path
		if len(query) != 0 {
			u = u + "?" + query.Encode()
		}
		c.Redirect("/index/index?url="+url.QueryEscape(u), http.StatusFound)
	}
	c.Main.Prepare()

	var err error
	if c.sess, err = globalSessions.SessionStart(c.Ctx.ResponseWriter, c.Ctx.Request); err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	uname := c.sess.Get("uname")
	if uname == nil {
		c.NeedLogon()
		return
	} else {
		m, err := auth.NewManager(uname.(string))
		if err != nil {
			c.Error(http.StatusInternalServerError, err.Error())
		}
		c.Data["manager"] = m
	}

	c.Data["nav"] = ""
	c.Data["subNav"] = ""
	// c.Data["role"] = manager
	c.Layout = "iframe_layout.html"
	c.LayoutSections = make(map[string]string)
	c.Data["requestID"] = fmt.Sprintf("%d", time.Now().Unix())
	logEntry := c.GetLogger()
	logEntry.Data["requestID"] = c.Data["requestID"]
}

func (c *Auth) NeedLogon() {
	var err error
	c.sess, err = globalSessions.SessionStart(c.Ctx.ResponseWriter, c.Ctx.Request)
	if err != nil {
		c.Error(http.StatusInternalServerError, err.Error())
	}
	uname := c.sess.Get("uname")
	if uname == nil {
		if c.Ctx.Input.AcceptsJSON() {
			c.Error(http.StatusForbidden, "请登陆后访问")
		} else {
			c.Redirect("/logon?url="+c.Ctx.Request.URL.EscapedPath(), 302)
		}
		return
	}
	manager, err := auth.NewManager(uname.(string))
	if err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	c.Data["manager"] = manager
}

func (c *Auth) Finish() {
	// c.sess.SessionRelease(c.Ctx.ResponseWriter)
	c.Main.Finish()
}

func (c *Auth) getManager() *auth.Manager {
	m, ok := c.Data["manager"]
	if !ok {
		return nil
	}
	return m.(*auth.Manager)
}

func (c *Auth) Valid(req interface{}) {
	valid := validation.Validation{}
	if b, err := valid.Valid(req); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	} else if !b {
		for k, v := range valid.ErrorsMap {
			s := make([]string, 0)
			for _, _err := range v {
				s = append(s, _err.Message)
			}
			c.Error(http.StatusBadRequest, fmt.Sprintf("%s字段 %s", k, strings.Join(s, "，")))
		}
		c.Error(http.StatusBadRequest, fmt.Sprintf("%v", valid.ErrorsMap))
	}
}
