package controllers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdinternal/appproto"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdinternal/oauth2"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdinternal/randstring"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdmodels"
	"github.com/chenwm-topstar/chargingc/utils/crypt"
	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/redigo"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
	"github.com/garyburd/redigo/redis"
	"github.com/medivhzhan/weapp/v3"
	"github.com/medivhzhan/weapp/v3/auth"
)

type AppBeferLoginController struct {
	AppController
}

func (c *AppBeferLoginController) Prepare() {
	c.AppController.Prepare()
}

func (c *AppBeferLoginController) CheckAppUpgrade() {
	req := &appproto.CheckAppUpgradeReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "request param error: "+err.Error())
	}
	v, _ := strconv.ParseInt(c.app.LatestVersion, 10, 64)
	resp := &appproto.CheckAppUpgradeReply{
		Force:           c.app.ForceUpgrade,
		NeedUpgrade:     v > int64(req.VersionCode),
		LastVersion:     c.app.LatestVersion,
		LastVersionCode: uint32(v),
	}

	c.Resp.RawData = resp
}

func (c *AppBeferLoginController) CheckAccountExists() {
	req := &appproto.CheckAccountExistsReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "request param error: "+err.Error())
	}
	if req.Account == "" {
		c.Error(http.StatusBadRequest, "request param is nil")
	}
	key := fmt.Sprintf("%s:%d:account:check:user", req.Account, c.app.ID)

	count := int64(0)
	if v, ok := ccache.Get(key); ok && v != nil {
		count = v.(int64)
	} else {
		tmp, _, _ := sg.Do(key, func() (interface{}, error) {
			cnt, err := gormv2.Count(c.HeaderToContext(), &models.User{}, "account=?", req.Account)
			if err != nil {
				c.Error(http.StatusBadRequest, "check account error: "+err.Error())
			}
			return cnt, nil
		})
		count = tmp.(int64)
		ccache.Set(key, count, 1*time.Minute)
	}

	if count > 0 {
		c.Error(http.StatusBadRequest, "account is exists")
	}
	c.Resp.RawData = &appproto.CheckAccountExistsReply{}
}

func (c *AppBeferLoginController) CheckEmailExists() {
	req := &appproto.CheckEmailExistsReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "request param error: "+err.Error())
	}
	if req.Email == "" {
		c.Error(http.StatusBadRequest, "request param is nil")
	}
	key := fmt.Sprintf("%s:%d:email:check:user", req.Email, c.app.ID)

	count := int64(0)
	if v, ok := ccache.Get(key); ok && v != nil {
		count = v.(int64)
	} else {
		tmp, _, _ := sg.Do(key, func() (interface{}, error) {
			cnt, err := gormv2.Count(c.HeaderToContext(), &models.User{}, "oauth_type=? and oauth_flag=?", 0, req.Email)
			if err != nil {
				c.Error(http.StatusBadRequest, "check account error: "+err.Error())
			}
			return cnt, nil
		})
		count = tmp.(int64)
		ccache.Set(key, count, 1*time.Minute)
	}

	if count > 0 {
		c.Error(http.StatusBadRequest, "email is exists")
	}
	c.Resp.RawData = &appproto.CheckEmailExistsReply{}
}

func (c *AppBeferLoginController) Register() {
	req := &appproto.UserRegisterReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "request param error: "+err.Error())
	}
	if req.Name == "" || req.Passwd == "" || req.Email == "" {
		c.Error(http.StatusBadRequest, "request param is nil")
	}
	key, _ := getVerifyCode(c.app.ID, req.Email, 1)
	verifycode, err := redis.String(redigo.Do("get", key))
	if err != nil {
		if err == redis.ErrNil {
			c.Error(http.StatusBadRequest, "verify code expired")
		}
		c.Error(http.StatusInternalServerError, "verify code error: "+err.Error())
	}
	if verifycode != req.VerifyCode {
		c.Error(http.StatusBadRequest, "verify code error")
	}
	user := &models.User{}

	count := int64(0)
	if count, err = gormv2.Count(c.HeaderToContext(), user, "appid=? and (oauth_type=? and oauth_flag=?)", c.app.ID, 0, req.Email); err != nil {
		c.Error(http.StatusBadRequest, "check email error: "+err.Error())
	}
	if count > 0 {
		c.Error(http.StatusBadRequest, "email is exists")
	}
	if count, err = gormv2.Count(c.HeaderToContext(), user, "appid=? and account=?", c.app.ID, req.Name); err != nil {
		c.Error(http.StatusBadRequest, "get user error: "+err.Error())
	}
	if count > 0 {
		c.Error(http.StatusBadRequest, "account is exists")
	}

	user.APPID = c.app.ID
	user.OAuthType = models.UserOAuthTypeEmail
	user.OAuthFlag = req.Email
	user.Account = req.Name
	user.Name = req.Name
	user.Email = req.Email
	user.Passwd = req.Passwd
	if err := gormv2.Save(c.HeaderToContext(), user); err != nil {
		c.Error(http.StatusInternalServerError, "save user info error: "+err.Error())
	}

	c.Resp.RawData = &appproto.UserRegisterReply{}
}

func (c *AppBeferLoginController) getName() string {
	for i := 0; i < 3; i++ {
		name := randstring.RandStringBytesMaskImprSrc(10)
		count, err := gormv2.Count(c.HeaderToContext(), &models.User{}, "name=?", name)
		if err == nil && count <= 0 {
			return name
		}
	}
	return ""
}

func (c *AppBeferLoginController) Login() {
	req := &appproto.UserLoginReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "parse param error: "+err.Error())
	}
	up := make(map[string]interface{})

	user := &models.User{}
	switch models.UserOAuthType(req.LoginType) {
	case models.UserOAuthTypeEmail:
		if req.Name == "" && req.Passwd == "" {
			c.Error(http.StatusBadRequest, "name or passwd  is null")
		}
		var ids []uuid.ID
		ids = append(ids, c.app.ID)
		if c.app.ID == 1 {
			ids = append(ids, 4)
		}

		if err := gormv2.Find(c.HeaderToContext(), user, "appid in (?) and (name=? or (oauth_type=? and oauth_flag=?)) and passwd=? and is_act=?", ids, req.Name, 0, req.Name, req.Passwd, true); err != nil {
			c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
		}
		if user.IsNew() {
			c.Error(http.StatusBadRequest, "account or passwd error")
		}
		if crypt.MD5(req.Passwd) != crypt.MD5(user.Passwd) {
			c.Error(http.StatusBadRequest, "passwd error")
		}

	case models.UserOAuthTypeAppleid:
		if req.AppleLoginInfo.IdentityToken == "" || req.AppleLoginInfo.UserId == "" {
			c.Error(http.StatusBadRequest, "apple IdentityToken is nil")
		}
		jwtClaims, err := oauth2.AppleVerifyIdentityToken(c.app.IOSClientId, req.AppleLoginInfo.IdentityToken, req.AppleLoginInfo.UserId)
		if err != nil {
			c.Error(http.StatusBadRequest, "apple verify error: "+err.Error())
		}

		if err := gormv2.Find(c.HeaderToContext(), user, "appid=? and oauth_type=? and oauth_flag=? and is_act=?", c.app.ID, models.UserOAuthTypeAppleid, req.AppleLoginInfo.UserId, true); err != nil {
			c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
		}

		if user.IsNew() {
			user.APPID = c.app.ID
			user.OAuthType = models.UserOAuthType(req.LoginType)
			user.OAuthFlag = req.AppleLoginInfo.UserId
			user.Account = c.getName()
			user.Name = req.AppleLoginInfo.FullName
			if user.Name == "" {
				user.Name = user.Account
			}
			user.Email = jwtClaims.Email
			user.Passwd = ""
			if err := gormv2.Save(c.HeaderToContext(), user); err != nil {
				c.Error(http.StatusInternalServerError, "save account error: "+err.Error())
			}
			if err := gormv2.Last(c.HeaderToContext(), user, "appid=? and oauth_type=? and oauth_flag=?", c.app.ID, models.UserOAuthTypeAppleid, req.AppleLoginInfo.UserId); err != nil {
				c.Error(http.StatusInternalServerError, "reload user error: "+err.Error())
			}
		}

	case models.UserOAuthTypeGoogle:
		// if req.GoogleLoginInfo.IdentityToken == "" {
		// 	c.Error(http.StatusBadRequest, "apple IdentityToken is nil")
		// }
		// jwtClaims, err := oauth2.GoogleVerifyIdentityToken(req.GoogleLoginInfo.IdentityToken)
		// if err != nil {
		// 	c.Error(http.StatusBadRequest, "apple verify error: "+err.Error())
		// }
		// c.GetLogger().Infof("--->apple info: [%+v][%+v]", req.GoogleLoginInfo, jwtClaims)
		if req.GoogleLoginInfo.Email == "" || req.GoogleLoginInfo.UserId == "" {
			c.Error(http.StatusBadRequest, "google req param error")
		}

		if err := gormv2.Find(c.HeaderToContext(), user, "appid=? and oauth_type=? and oauth_flag=? and is_act=?", c.app.ID, models.UserOAuthTypeGoogle, req.GoogleLoginInfo.UserId, true); err != nil {
			c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
		}

		if user.IsNew() {
			user.APPID = c.app.ID
			user.OAuthType = models.UserOAuthType(req.LoginType)
			user.OAuthFlag = req.GoogleLoginInfo.UserId
			user.Account = c.getName()
			user.Name = req.GoogleLoginInfo.FullName
			if user.Name == "" {
				user.Name = user.Account
			}
			user.Email = req.GoogleLoginInfo.Email
			user.Passwd = ""

			if err := gormv2.Save(c.HeaderToContext(), user); err != nil {
				c.Error(http.StatusInternalServerError, "save account error: "+err.Error())
			}
			if err := gormv2.Last(c.HeaderToContext(), user, "appid=? and oauth_type=? and oauth_flag=? and is_act=?", c.app.ID, models.UserOAuthTypeGoogle, req.GoogleLoginInfo.UserId, true); err != nil {
				c.Error(http.StatusInternalServerError, "reload user error: "+err.Error())
			}
		}

	case models.UserOAuthTypeWX:
		if req.WXLoginInfo == nil || req.WXLoginInfo.Code == "" {
			c.Error(http.StatusBadRequest, "code is nil")
		}

		rsp, err := weapp.NewClient(c.app.Config.WXAppid, c.app.Config.WXSecret).
			NewAuth().
			Code2Session(&auth.Code2SessionRequest{
				Appid:     c.app.Config.WXAppid,
				Secret:    c.app.Config.WXSecret,
				JsCode:    req.WXLoginInfo.Code,
				GrantType: "authorization_code",
			}) // 登录凭证校验
		if err != nil {
			c.Error(http.StatusBadRequest, "登陆失败:"+err.Error())
		}
		if rsp.ErrCode != 0 {
			c.Error(http.StatusBadRequest, "登陆失败:"+rsp.GetResponseError().Error())
		}

		if err := gormv2.Find(c.HeaderToContext(), user, "appid=? and oauth_type=? and oauth_flag=? and is_act=?", c.app.ID, models.UserOAuthTypeWX, rsp.Openid, true); err != nil {
			c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
		}

		if user.IsNew() {
			user.APPID = c.app.ID
			user.OAuthType = models.UserOAuthType(req.LoginType)
			user.OAuthFlag = rsp.Openid
			user.Account = c.getName()
			user.EX.SessionKey = rsp.SessionKey
			if user.Name == "" {
				user.Name = user.Account
			}
			user.Passwd = ""
			if err := gormv2.Save(c.HeaderToContext(), user); err != nil {
				c.Error(http.StatusInternalServerError, "save account error: "+err.Error())
			}
			if err := gormv2.Last(c.HeaderToContext(), user, "appid=? and oauth_type=? and oauth_flag=?", c.app.ID, models.UserOAuthTypeWX, rsp.Openid); err != nil {
				c.Error(http.StatusInternalServerError, "reload user error: "+err.Error())
			}
		} else {
			user.EX.SessionKey = rsp.SessionKey
			up["ex"], _ = gormv2.JsonValue(user.EX)
		}

	default:
		c.Error(http.StatusInternalServerError, "phone or name is nil")
	}

	if user.Zone != req.Zone {
		up["zone"] = req.Zone
	}

	if len(up) > 0 {
		if err := gormv2.Updates(c.HeaderToContext(), user, up); err != nil {
			c.Error(http.StatusInternalServerError, "update zone error")
		}
	}

	token := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("UserID:%d,Expired:%d", user.ID, time.Now().Add(15*24*time.Hour)))))
	if _, err := redigo.Do("set", token+":token:private", user.ID.String(), "ex", 15*24*3600); err != nil {
		c.Error(http.StatusInternalServerError, "gen token fail")
	}
	c.Resp.RawData = &appproto.UserLoginReply{
		UID:               user.ID.String(),
		Account:           user.Account,
		Name:              user.Name,
		Token:             token,
		Email:             user.Email,
		ManualCtrlConnect: user.ManualCtrlConnect,
		Apikey:            c.app.Config.Apikey,
		FID:               c.app.Config.Fid,
	}
}
func (c *AppBeferLoginController) ForgotPasswd() {
	req := &appproto.UserForgotPasswdReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "parse param error: "+err.Error())
	}
	if req.Email == "" {
		c.Error(http.StatusBadRequest, "email is nil")
	}
	key, _ := getVerifyCode(c.app.ID, req.Email, 2)
	verifycode, err := redis.String(redigo.Do("get", key))
	if err != nil {
		if err == redis.ErrNil {
			c.Error(http.StatusBadRequest, "verify code expired")
		}
		c.Error(http.StatusInternalServerError, "verify code error: "+err.Error())
	}
	if verifycode != req.VerifyCode {
		c.Error(http.StatusBadRequest, "verify code error")
	}

	privateUser := &models.User{}
	if err := gormv2.Find(c.HeaderToContext(), privateUser, "appid=? and email=? ", c.app.ID, req.Email); err != nil {
		c.Error(http.StatusInternalServerError, "get account info error: "+err.Error())
	}

	if privateUser.IsNew() {
		c.Error(http.StatusNotFound, "email not fount")
	}

	content := fmt.Sprintf(accountTmpl, privateUser.Email, privateUser.Account, privateUser.Passwd)
	c.app.SendMail([]string{req.Email}, []string{}, "Obtain the charging APP password", content)

	c.Resp.RawData = &appproto.UserForgotPasswdReply{}
}

func (c *AppBeferLoginController) VerifyCode() {
	req := &appproto.VerifyCodeReq{}
	if err := json.Unmarshal([]byte(c.Req.Data), req); err != nil {
		c.Error(http.StatusBadRequest, "unmarshal error: "+err.Error())
	}
	key, verifyCode := getVerifyCode(c.app.ID, req.Email, req.Op)
	if ttl, err := redis.Uint64(redigo.Do("ttl", key)); err == nil || ttl > 5 {
		c.Error(http.StatusTooManyRequests, "too many requests")
	}
	if _, err := redigo.Do("set", key, verifyCode, "ex", 305); err != nil {
		c.Error(http.StatusInternalServerError, "gen verify code error: "+err.Error())
	}

	subject, content := "", ""
	switch req.Op {
	case 1: // 注册
		subject = "Charging APP email registration verification"
		content = fmt.Sprintf(registerTmpl, verifyCode)
	case 2: // 忘记密码
		subject = "Charging APP forget password email verification"
		content = fmt.Sprintf(forgotTmpl, verifyCode)
	case 3: // 发送删除账户信息
		subject = "Charging APP email Delete account verification"
		content = fmt.Sprintf(deleteTmpl, verifyCode)
	default:
		c.Error(http.StatusBadRequest, "not support verifycode request")
	}
	c.app.SendMail([]string{req.Email}, []string{}, subject, content)

	c.Resp.RawData = &appproto.VerifyCodeReply{}
}

var accountTmpl = `
<html>
 <head></head>
 <body>
  <div> 
   <h1 style="font-size:32px;line-height:36px;font-weight:500;padding-bottom:10px;color:#333;text-align:center">Obtain the charging APP account password</h1> 
   <div style="font-size:17px;line-height:25px;color:#333;font-weight:normal">
    <span class="im"> <p></p> <p>The charging APP you applied for forgot the password verification has been passed, please pay attention to receive the account password。</span>
    <div style="font-size:23px;line-height:25px;color:#333;font-weight:normal"> 
     <p>account: <b>%s or %s</b></p>
     <p>password: <b>%s</b></p>
    </div>
    <span class="im"> <p>Pay attention to account security。</p>
     <div class="adL"> 
      <p></p> 
     </div></span>
   </div>
   <div class="adL"> 
   </div>
  </div>
 </body>
</html>
`

var registerTmpl = `
<html>
 <head></head>
 <body>
  <div> 
   <h1 style="font-size:32px;line-height:36px;font-weight:500;padding-bottom:10px;color:#333;text-align:center">Verify your charging APP registration email address</h1> 
   <div style="font-size:17px;line-height:25px;color:#333;font-weight:normal">
    <span class="im"> <p></p> <p>You have selected this email address as your registration email address for the <span style="white-space:nowrap">charging APP</span>。
	To verify that this email address belongs to you，<wbr />please enter the following verification code on the email verification page:</p> </span>
    <div style="font-size:23px;line-height:25px;color:#333;font-weight:normal"> 
     <p><b>%s</b></p>
    </div>
    <span class="im"> <p>The verification code will expire 5 minutes after this email is sent。</p>
     <div class="adL"> 
      <p></p> 
     </div></span>
   </div>
   <div class="adL"> 
   </div>
  </div>
 </body>
</html>
`

var forgotTmpl = `
<html>
 <head></head>
 <body>
  <div> 
   <h1 style="font-size:32px;line-height:36px;font-weight:500;padding-bottom:10px;color:#333;text-align:center">Verify your charging APP forgot your password email</h1> 
   <div style="font-size:17px;line-height:25px;color:#333;font-weight:normal">
    <span class="im"> <p></p> <p>You have selected this email address as the forgotten password verification address for the <span style="white-space:nowrap">charging APP</span>。
	To verify that this email address belongs to you，<wbr />please enter the following verification code on the email verification page:</p> </span>
    <div style="font-size:23px;line-height:25px;color:#333;font-weight:normal"> 
     <p><b>%s</b></p>
    </div>
    <span class="im"> <p>The verification code will expire 5 minutes after this email is sent。</p>
     <div class="adL"> 
      <p></p> 
     </div></span>
   </div>
   <div class="adL"> 
   </div>
  </div>
 </body>
</html>
`

var deleteTmpl = `
<html>
 <head></head>
 <body>
  <div> 
   <h1 style="font-size:32px;line-height:36px;font-weight:500;padding-bottom:10px;color:#333;text-align:center">Verify your charging APP delete account email</h1> 
   <div style="font-size:17px;line-height:25px;color:#333;font-weight:normal">
    <span class="im"> <p></p> <p>You have selected this email address as the delete account verification address for the <span style="white-space:nowrap">charging APP</span>。
	To verify that this email address belongs to you，<wbr />please enter the following verification code on the email verification page:</p> </span>
    <div style="font-size:23px;line-height:25px;color:#333;font-weight:normal"> 
     <p><b>%s</b></p>
    </div>
    <span class="im"> <p>The verification code will expire 5 minutes after this email is sent。</p>
     <div class="adL"> 
      <p></p> 
     </div></span>
   </div>
   <div class="adL"> 
   </div>
  </div>
 </body>
</html>
`
