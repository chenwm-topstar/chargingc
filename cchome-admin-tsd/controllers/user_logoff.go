package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/chenwm-topstar/chargingc/cchome-admin-tsd/models"
	"github.com/chenwm-topstar/chargingc/utils/crypt"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/lg"
	"github.com/chenwm-topstar/chargingc/utils/redigo"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
	"github.com/garyburd/redigo/redis"
)

type UserLogoffController struct {
	Main
}

func (c *UserLogoffController) Prepare() {
	c.Main.Prepare()

	appidStr := c.GetString(":appid")
	if appidStr == "" {
		c.Error(http.StatusBadRequest, "missing app info")
	}

	appid, _ := strconv.Atoi(appidStr)
	if appid != 0 {
		ai := &models.APPConfig{}
		if err := gormv2.MustFind(c.HeaderToContext(), ai, "id=?", appid); err != nil {
			c.Error(http.StatusBadRequest, "load app error: "+err.Error())
		}
		appidStr = ai.Name
	}

	appInfo, err := models.GetAppByName(c.HeaderToContext(), appidStr)
	if err != nil {
		c.Error(http.StatusBadRequest, "load app info error: "+err.Error())
	}
	c.Data["app"] = appInfo
	c.Data["appid"] = appInfo.ID

}
func (c *UserLogoffController) Logoff() {
	c.TplName = "user/logoff.html"
	if !c.Ctx.Input.IsPost() {
		c.Data["config"].(map[string]interface{})["jsname"] = "backend/user"
		c.Data["config"].(map[string]interface{})["actionname"] = "logoff"
		c.Data["title"] = "user logoff"
		return
	}

	from := Manager{}
	if err := c.ParseForm(&from); err != nil {
		c.Error(http.StatusBadRequest, err.Error())
	}
	if from.Name == "" || from.Password == "" {
		c.Error(http.StatusBadRequest, "missing info")
	}
	verifyCodeFlag := false

	var users []*models.User
	if isEmailValid(from.Name) { // 验证码处理
		key, _ := getVerifyCode(c.Data["appid"].(uuid.ID), from.Name, 3)
		vcode, err := redis.String(redigo.Do("get", key))
		if err != nil && err != redis.ErrNil {
			c.Error(http.StatusInternalServerError, "get verify code error. Please later try again")
		}
		if err != redis.ErrNil {
			verifyCodeFlag = true
			if vcode != from.Password {
				c.Error(http.StatusInternalServerError, "verify code error. Please later try again")
			}

			if err := gormv2.Find(c.HeaderToContext(), &users, "appid=? and email=? and is_act=?", c.Data["appid"], from.Name, true); err != nil {
				c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
			}
			err = nil
		}
	}

	if len(users) == 0 && !verifyCodeFlag {
		user := &models.User{}
		if err := gormv2.Find(c.HeaderToContext(), user, "appid=? and (name=? or (oauth_type=? and oauth_flag=?)) and passwd=? and is_act=?", c.Data["appid"], from.Name, 0, from.Name, from.Password, true); err != nil {
			c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
		}
		if user.IsNew() {
			c.Error(http.StatusBadRequest, "not found. Try again after 5 minutes")
		}
		if crypt.MD5(from.Password) != crypt.MD5(user.Passwd) {
			c.Error(http.StatusBadRequest, "passwd error")
		}
		users = append(users, user)
	}

	var err error
	ctx := gormv2.NewDBContext(c.HeaderToContext(), gormv2.Begin())
	defer func() {
		if cerr := gormv2.CommitWithErr(ctx, err); cerr != nil && err == nil {
			err = cerr
		}
		if err != nil {
			c.Error(http.StatusBadRequest, err.Error())
		}
		c.CustomAbort(http.StatusOK, "delete account success")
		c.JsonData = map[string]interface{}{}
		c.Msg = "delete account success"
	}()

	for _, user := range users {
		var binds []*models.EvseBind
		if err := gormv2.Find(c.HeaderToContext(), &binds, "uid=?", user.ID); err != nil {
			lg.Infof("clean bind evse error: " + err.Error())
			c.Error(http.StatusBadRequest, "clean bind evse fail")
		}

		for _, evseBind := range binds {
			if evseBind.IsMaster != nil && *evseBind.IsMaster {
				if err = gormv2.FromDBContext(ctx).Unscoped().Delete(&models.EvseBind{}, "sn=?", evseBind.SN).Error; err != nil {
					err = fmt.Errorf("unbind all error: " + err.Error())
					return
				}
			} else {
				if err = gormv2.FromDBContext(ctx).Unscoped().Delete(&models.EvseBind{}, "id", evseBind.ID).Error; err != nil {
					err = fmt.Errorf("unbind error: " + err.Error())
					return
				}
			}
		}

		user.IsACT = nil
		if err = gormv2.Save(ctx, user); err != nil {
			err = fmt.Errorf("delete error: " + err.Error())
			return
		}
	}

}

func (c *UserLogoffController) VerifyCode() {
	if !c.Ctx.Input.IsPost() {
		return
	}
	email := c.GetString("email")
	if email == "" {
		c.Error(http.StatusBadRequest, "missing email")
	}
	if !isEmailValid(email) { // 验证码处理
		c.Error(http.StatusBadRequest, "email format error")
	}

	key, verifyCode := getVerifyCode(c.Data["appid"].(uuid.ID), email, 3)
	if ttl, err := redis.Uint64(redigo.Do("ttl", key)); err == nil || ttl > 5 {
		c.Error(http.StatusTooManyRequests, "too many requests. Try again after 5 minutes")
	}

	cnt, err := gormv2.Count(c.HeaderToContext(), &models.User{}, "appid=? and email=? and is_act=?", c.Data["appid"], email, true)
	if err != nil {
		c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
	}
	if cnt <= 0 {
		c.Error(http.StatusBadRequest, "email not found")
	}

	if _, err := redigo.Do("set", key, verifyCode, "ex", 305); err != nil {
		c.Error(http.StatusInternalServerError, "gen verify code error: "+err.Error())
	}

	subject := "Charging APP email Delete account verification"
	content := fmt.Sprintf(deleteTmpl, verifyCode)

	c.Data["app"].(*models.APPConfig).SendMail([]string{email}, []string{}, subject, content)
}

func isEmailValid(email string) bool {
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(regex, email)
	return match
}
