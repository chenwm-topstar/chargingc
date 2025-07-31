package models

import (
	"context"
	"crypto/tls"
	"database/sql/driver"
	"encoding/json"
	"sync"
	"time"

	"github.com/chenwm-topstar/chargingc/utils/gormv2"
	"github.com/chenwm-topstar/chargingc/utils/lg"
	"github.com/chenwm-topstar/chargingc/utils/uuid"
	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"
)

// app 配置信息
type APPConfig struct {
	ID              uuid.ID         `gorm:"column:id;primary_key;" `                       // ID
	Name            string          `gorm:"column:name;type:char(20);uniqueIndex:u_name" ` // 名称
	LatestVersion   string          `gorm:"column:latest_version;type:char(20);" `         // 最新版本
	ForceUpgrade    bool            `gorm:"column:force_upgrade;" `                        // 强制升级
	IOSClientId     string          `gorm:"column:ios_client_id;type:char(64);" `          // ios客户端ID, 登录校验时使用
	AndroidClientId string          `gorm:"column:android_client_id;type:char(64);" `      // android客户端ID, 登录校验时使用
	EmailConfig     KindEmailConfig `gorm:"column:email_config;type:text;" `               // 邮件配置

	senderOnce sync.Once            `gorm:"-"`
	mailCH     chan *gomail.Message `gorm:"-"`

	gormv2.Base
}

func (e APPConfig) DBName() string {
	return "cchome-admin"
}

func (e APPConfig) TableName() string {
	return "app_configs"
}

type KindEmailConfig struct {
	SendHost      string `json:"send_host"`       // 发送服务器服务器. e.g.: smtp.exmail.qq.com
	SendPort      int    `json:"send_port"`       // 发送服务器服务器端口. e.g.: 465
	UserName      string `json:"user_name"`       // 发送授权账号. 一般是邮箱账号
	Passwd        string `json:"passwd"`          // 发送授权账号密码. 一般是邮箱密码
	DefSenderMail string `json:"def_sender_mail"` // 默认发送邮件
}

func (c KindEmailConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *KindEmailConfig) Scan(input interface{}) error {
	switch input.(type) {
	case []byte:
		if err := json.Unmarshal(input.([]byte), c); err != nil {
			return err
		}
	}
	return nil
}

func (e *APPConfig) SendMail(tos, ccs []string, subject, content string) {
	mailMsg := gomail.NewMessage()
	mailMsg.SetHeader("From", e.EmailConfig.DefSenderMail)
	mailMsg.SetHeader("To", tos...)
	mailMsg.SetHeader("Cc", ccs...)
	mailMsg.SetHeader("Subject", subject)
	mailMsg.SetBody("text/html", content)

	e.mailCH <- mailMsg
}

func (e *APPConfig) SenderEmailWorker(context.Context) error {
	e.senderOnce.Do(func() {
		go func() {
			d := gomail.NewDialer(e.EmailConfig.SendHost, e.EmailConfig.SendPort, e.EmailConfig.UserName, e.EmailConfig.Passwd)
			d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
			var s gomail.SendCloser
			var err error
			open := false
			for {
				if err = func() (err error) {
					select {
					case m, ok := <-e.mailCH:
						if !ok {
							return errors.Wrap(err, "mail is nil")
						}
						if !open {
							if s, err = d.Dial(); err != nil {
								return errors.Wrap(err, "dial smtp server error")
							}
							open = true
						}
						if err := gomail.Send(s, m); err != nil {
							return errors.Wrapf(err, "send email[%+v] error", m)
						}
					// Close the connection to the SMTP server if no email was sent in
					// the last 30 seconds.
					case <-time.After(30 * time.Second):
						if open {
							if err := s.Close(); err != nil {
								return errors.Wrapf(err, "close sender error")
							}
							open = false
						}
					}
					return nil
				}(); err != nil {
					lg.Errorf("send mail process error: %+v", err)
				}

			}
		}()
	})
	return nil
}
