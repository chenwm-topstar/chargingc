package routers

// @APIVersion 1.0.0
// @Title Core API
// @Description 充电业务接口
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html

//{{{
// @Definition datasource.UUID string
//}}}

import (
	"github.com/astaxie/beego"
	"github.com/chenwm-topstar/chargingc/cchome-admin-tsdcontrollers"
)

func init() {
	//fmt.Println(beego.BConfig)
	// if os.Getenv("RUN_MODE") == "dev" {
	// log.SetLevel(log.DebugLevel)
	// beego.BConfig.RunMode = "dev"
	beego.BConfig.EnableErrorsRender = true
	// } else {
	beego.BConfig.RunMode = "prod"
	// }

	beego.BConfig.CopyRequestBody = true
	beego.BConfig.RouterCaseSensitive = true
	beego.ErrorController(&controllers.Error{})
	beego.AutoRouter(&controllers.AjaxController{})

	index := &controllers.IndexController{}
	beego.Router("/", index, "get:Index")
	beego.Router("/index", index, "get:Index")
	beego.AutoRouter(index)

	logon := &controllers.LogonController{}
	beego.Router("/logon", logon, "*:Index")

	manager := &controllers.ManagerController{}
	beego.AutoRouter(manager)

	deviceList := &controllers.DeviceListController{}
	beego.Router("/device/list", deviceList, "*:List")

	deviceLatestVersion := &controllers.DeviceLatestVersionController{}
	beego.Router("/device/latest", deviceLatestVersion, "*:Edit")

	device := &controllers.DeviceController{}
	beego.Router("/device/info", device, "*:Info")
	beego.Router("/device/connector", device, "*:Connector")
	beego.Router("/device/evse/start_charger", device, "*:StartCharger")
	beego.Router("/device/evse/stop_charger", device, "*:StopCharger")
	beego.Router("/device/upgrade", device, "*:Upgrade")
	beego.Router("/device/getappointconfig", device, "*:GetAppointConfig")
	beego.Router("/device/setconfig", device, "*:SetConfig")
	beego.Router("/device/reset", device, "*:Reset")
	beego.Router("/device/reboot", device, "*:Reboot")

	dlvl := &controllers.DeviceLatestVersionListController{}
	beego.Router("/dlv/list", dlvl, "*:List")
	beego.Router("/dlv/add", dlvl, "*:Add")

	dlv := &controllers.DeviceLatestVersionController{}
	beego.Router("/dlv/edit/ids/:id", dlv, "*:Edit")

	userList := &controllers.UserListController{}
	beego.Router("/user/list", userList, "*:List")

	userLogoff := &controllers.UserLogoffController{}
	beego.Router("/user/logoff/:appid", userLogoff, "*:Logoff")
	beego.Router("/user/verify_code/:appid", userLogoff, "*:VerifyCode")

	recordList := &controllers.OrderListController{}
	beego.Router("/order/list", recordList, "*:List")

	about := &controllers.AboutController{}
	beego.Router("/about", about, "*:Edit")

	feedback := &controllers.FeedbackController{}
	beego.Router("/feedback/list", feedback, "*:List")
	beego.Router("/feedback/edit/ids/:id", feedback, "*:Edit")

	qaaList := &controllers.QAAController{}
	beego.Router("/qaa/list", qaaList, "*:List")
	beego.Router("/qaa/add", qaaList, "*:Add")
	beego.Router("/qaa/edit/ids/:id", qaaList, "*:Edit")
	beego.Router("/qaa/del/ids/:id", qaaList, "*:Del")

	appBeferLogin := &controllers.AppBeferLoginController{}
	beego.Router("/privatec/v1.0/user/register", appBeferLogin, "*:Register")
	beego.Router("/privatec/v1.0/user/login", appBeferLogin, "*:Login")
	beego.Router("/privatec/v1.0/user/passwd/forgot", appBeferLogin, "*:ForgotPasswd")
	beego.Router("/privatec/v1.0/verify_code", appBeferLogin, "*:VerifyCode")
	beego.Router("/privatec/v1.0/user/check/account", appBeferLogin, "*:CheckAccountExists")
	beego.Router("/privatec/v1.0/user/check/email", appBeferLogin, "*:CheckEmailExists")
	beego.Router("/privatec/v1.0/app/check/upgrade", appBeferLogin, "*:CheckAppUpgrade")

	appAfterLogin := &controllers.AppAfterLoginController{}
	beego.Router("/privatec/v1.0/user/apikey", appAfterLogin, "*:GetApikeyInfo")
	beego.Router("/privatec/v1.0/user/logout", appAfterLogin, "*:Logout")
	beego.Router("/privatec/v1.0/user/logoff", appAfterLogin, "*:Logoff")
	beego.Router("/privatec/v1.0/user/change_passwd", appAfterLogin, "*:ChangePasswd")
	beego.Router("/privatec/v1.0/user/change_info", appAfterLogin, "*:ChangeUserInfo")
	beego.Router("/privatec/v1.0/user/evses", appAfterLogin, "*:EvseList")
	beego.Router("/privatec/v1.0/user/bind", appAfterLogin, "*:BindEvse")
	beego.Router("/privatec/v1.0/user/unbind", appAfterLogin, "*:UnbindEvse")

	beego.Router("/privatec/v1.0/user/get_profile", appAfterLogin, "*:GetProfile")
	beego.Router("/privatec/v1.0/user/set_profile", appAfterLogin, "*:SetProfile")
	beego.Router("/privatec/v1.0/user/statistics", appAfterLogin, "*:Statistics")

	beego.Router("/privatec/v1.0/orders", appAfterLogin, "*:Orders")
	beego.Router("/privatec/v1.0/evse/sync_bt_order", appAfterLogin, "*:SyncBTOrder")
	beego.Router("/privatec/v1.0/evse/info", appAfterLogin, "*:GetEvseInfo")
	beego.Router("/privatec/v1.0/evse/changeinfo", appAfterLogin, "*:ChangeEvseInfo")
	beego.Router("/privatec/v1.0/evse/start", appAfterLogin, "*:StartCharger")
	beego.Router("/privatec/v1.0/evse/stop", appAfterLogin, "*:StopCharger")
	beego.Router("/privatec/v1.0/evse/reset", appAfterLogin, "*:Reset")
	beego.Router("/privatec/v1.0/evse/setreserver", appAfterLogin, "*:SetReserverInfo")
	beego.Router("/privatec/v1.0/evse/getreserver", appAfterLogin, "*:GetReserverInfo")
	beego.Router("/privatec/v1.0/evse/get_whitelist_card", appAfterLogin, "*:GetWhitelistCard")
	beego.Router("/privatec/v1.0/evse/set_whitelist_card", appAfterLogin, "*:SetWhitelistCard")
	beego.Router("/privatec/v1.0/evse/set_current", appAfterLogin, "*:SetEvseCurrent")
	beego.Router("/privatec/v1.0/evse/get_work_mode", appAfterLogin, "*:GetWorkMode")
	beego.Router("/privatec/v1.0/evse/set_work_mode", appAfterLogin, "*:SetWorkMode")
	beego.Router("/privatec/v1.0/evse/reboot", appAfterLogin, "*:EvseReboot")
	beego.Router("/privatec/v1.0/evse/reset", appAfterLogin, "*:EvseReset")
	beego.Router("/privatec/v1.0/evse/authcode", appAfterLogin, "*:EvseAuthCode")
	beego.Router("/privatec/v1.0/evse/bind_members", appAfterLogin, "*:BindMembers")

	beego.Router("/privatec/v1.0/about", appAfterLogin, "*:About")
	beego.Router("/privatec/v1.0/question_and_answer", appAfterLogin, "*:QuestionAndAnswer")
	beego.Router("/privatec/v1.0/feedback", appAfterLogin, "*:Feedback")
	beego.Router("/privatec/v1.0/latest_firmware_version", appAfterLogin, "*:LatestFirmwareVersion")
	beego.Router("/privatec/v1.0/evse/ota_upgrade", appAfterLogin, "*:OTAUpgrade")

}
