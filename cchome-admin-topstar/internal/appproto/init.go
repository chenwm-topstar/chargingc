package appproto

type StatusCode string

const (
	StatusCodeSuccess           StatusCode = "0"
	StatusCodeInvalidParamError StatusCode = "400"
	StatusCodeSystemBusy        StatusCode = "403"
	StatusCodeSignErr           StatusCode = "405"
	StatusCodeInternelError     StatusCode = "500"
)

func (s StatusCode) String() string {
	switch s {
	case StatusCodeSuccess:
		return "请求成功"
	case StatusCodeSystemBusy:
		return "系统繁忙"
	case StatusCodeSignErr:
		return "签名错误"

	}
	return "系统错误"
}

type Request struct {
	Uid       string      `json:"uid"`       // 用户ID
	Data      string      `json:"data"`      // 各接口具体请求参数组成的json字符串
	Timestamp int64       `json:"timestamp"` // 接口请求时的时间戳信息
	Sign      string      `json:"sign"`      // data+timeStamp+"chargingc"组成的参数签名，暂时采用md5算法
	RawData   interface{} `json:"-"`         // Data的序列化后的对象
}

type Response struct {
	Ret     int         `json:"ret"`  // 必填字段
	Msg     string      `json:"msg"`  // 具体错误信息，无错误返回成功信息
	Data    string      `json:"data"` // 参数内容，所有数据采用UTF-8编码，JSON格式
	RawData interface{} `json:"-"`    // Data的序列化后的对象
}

type VerifyCodeReq struct {
	Email string `json:"email"` // 邮箱
	Op    int    `json:"op"`    // 1 注册请求验证码, 2 忘记密码请求验证码
}
type VerifyCodeReply struct {
}

type CheckAppUpgradeReq struct {
}

type CheckAppUpgradeReply struct {
	NeedUpgrade bool `json:"need_upgrade"`
}

type CheckAccountExistsReq struct {
	Account string `json:"account"`
}
type CheckAccountExistsReply struct {
}
type CheckEmailExistsReq struct {
	Email string `json:"email"`
}
type CheckEmailExistsReply struct {
}

type UserRegisterReq struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Passwd     string `json:"passwd"`
	VerifyCode string `json:"verify_code"`
}
type UserRegisterReply struct {
}

type UserLoginReq struct {
	LoginType uint8 `json:"login_type"` // 登录方式  0 邮箱登录 1 手机号登录 2 appleid登录 3 google登录

	Name   string `json:"name"`
	Passwd string `json:"passwd"`

	AppleLoginInfo  *AppleLoginInfo  `json:"apple_login_info"`  // apply登录
	GoogleLoginInfo *GoogleLoginInfo `json:"google_login_info"` // google登录
}

type GoogleLoginInfo struct {
	IdentityToken string `json:"identity_token"`
	UserId        string `json:"user_id"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
}

type AppleLoginInfo struct {
	UserId            string `json:"user_id"`
	Email             string `json:"email"`
	FullName          string `json:"full_name"`
	AuthorizationCode string `json:"authorization_code"`
	IdentityToken     string `json:"identity_token"`
}

type UserLoginReply struct {
	UID   string `json:"uid"`
	Name  string `json:"name"`
	Token string `json:"token"`
	Email string `json:"email"`
}

type UserLogoffReq struct {
}
type UserLogoffReply struct {
}

type ChangeUserInfoReq struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
type ChangeUserInfoReply struct {
}

type ChangeEvseInfoReq struct {
	SN string `json:"sn"`

	Alias string `json:"alias"`
}
type ChangeEvseInfoReply struct {
}

type ChangePasswdReq struct {
	CurrentPasswd string `json:"current_passwd"`
	NewPasswd     string `json:"new_passwd"`
}
type ChangePasswdReply struct {
}

type ReserverInfo struct {
	Flag       uint8  `json:"flag"`        // Bit7代表单次，bit0到bit6代表周 一到周日
	StartTime  uint32 `json:"start_time"`  // 每天开始时间
	ChargeTime uint16 `json:"charge_time"` // 充电时间
}

type GetReserverInfoReq struct {
	SN string `json:"sn"` // 桩号
}
type GetReserverInfoReply struct {
	ReserverInfos []ReserverInfo `json:"reserver_infos"`
}

type GetWorkModeReq struct {
	SN string `json:"sn"` //	桩号
}
type GetWorkModeReply struct {
	WorkMode uint8 `json:"work_mode"` // 0：取消即插即充（APP模式） 1：使能即插即充
}

type SetWorkModeReq struct {
	SN       string `json:"sn"`        //	桩号
	WorkMode uint8  `json:"work_mode"` // 0：取消即插即充（APP模式） 1：使能即插即充
}
type SetWorkModeReply struct {
}

type SetReserverInfoReq struct {
	SN            string         `json:"sn"` //	桩号
	ReserverInfos []ReserverInfo `json:"reserver_infos"`
}
type SetReserverInfoReply struct {
}

type GetWhitelistCardReq struct {
	SN string `json:"sn"` // 桩号
}
type GetWhitelistCardReply struct {
	Cards []string `json:"cards"`
}

type SetWhitelistCardReq struct {
	SN    string `json:"sn"`     //	桩号
	IsDel bool   `json:"is_del"` // 是否删除
	Card  string `json:"card"`   // card
}
type SetWhitelistCardReply struct {
}

type UserForgotPasswdReq struct {
	Email      string `json:"email"`
	VerifyCode string `json:"verify_code"`
}
type UserForgotPasswdReply struct {
}

type UserBindEvseReq struct {
	SN  string `json:"sn"`  //	桩号
	Mac string `json:"mac"` // mac地址

	Auth           string `json:"auth,omitempty"`            //	授权
	EnableCharging bool   `json:"enable_charging,omitempty"` //	app调用时使用
}

type EvseStaticData struct {
	SN              string `json:"sn"`                //	桩号
	PileModel       string `json:"pile_model"`        //	产品型号
	EquipmentType   string `json:"equipment_type"`    //	设备类型
	FirmwareVersion uint16 `json:"firmware_version"`  // 固件版本号 BIN 1 例如：1.1 ，11.2 的 10 倍
	BTVersion       uint16 `json:"bt_version"`        // 蓝牙软件版本 BIN 1 例如：1.1 ，11.2 的 10 倍
	VehicleType     string `json:"vehicle_type"`      //	使用车辆类型
	RatedPower      int    `json:"rated_power"`       //	额定功率
	RatedMinCurrent int    `json:"rated_min_current"` //	额定最小电流
	RatedMaxCurrent int    `json:"rated_max_current"` //	额定最大电流
	RatedVoltage    int    `json:"rated_voltage"`     //	额定电压
	Mac             string `json:"mac"`               // mac 地址
	Alias           string `json:"alias"`
}

type BindEvseInfo struct {
	EvseStaticData
	Status         int  `json:"status"`          //  0 未知, （说明设备未上线过） 1 离线 2 在线 3 故障
	IsMaster       bool `json:"is_master"`       //
	EnableCharging bool `json:"enable_charging"` //
	IsDefaultEvse  bool `json:"is_default_evse"` //
}
type UserBindEvseReply BindEvseInfo

type UserUnbindEvseReq struct {
	SN string `json:"sn"` //	桩号
}

type UserUnbindEvseReply struct {
}

type EvseInfoReq struct {
	SN      string `json:"sn"`       // 是	设备ID
	OrderID string `json:"order_id"` // 否	订单号， 充电时必须传
}

type EvseDynamicData struct {
	SN                   string `json:"sn"`                     // 是	设备ID
	OrderID              string `json:"order_id"`               // 否	订单号， 充电时必须传
	ChargingVoltage      int    `json:"charging_voltage"`       // 充电电压
	ChargingCurrent      int    `json:"charging_current"`       // 充电电流
	ChargingPower        int    `json:"charging_power"`         // 充电功率
	ChargedElectricity   int    `json:"charged_electricity"`    // 已充电量
	StartChargingTime    int64  `json:"start_charging_time"`    // 开始充电时间
	ChargingTime         int64  `json:"charging_time"`          // 充电时间（分钟）
	Status               string `json:"status"`                 // 桩状态描述
	ConnectingStatus     int    `json:"connecting_status"`      // 枪状态描述
	ConnectingStatusDesc string `json:"connecting_status_desc"` // 枪状态描述
	OrderStatus          int    `json:"order_status"`           // 订单状态
	ReservedStartTime    int64  `json:"reserved_start_time"`    // 预约开始时间
	ReservedStopTime     int64  `json:"reserved_stop_time"`     // 预约结束时间
	StartType            int    `json:"start_type"`             // 0 小程序，1 刷卡
	Phone                string `json:"phone"`                  // 启动手机号
	FaultCode            uint16 `json:"fault_code"`             // 故障码
}
type EvseInfoReply struct {
	EvseDynamicData
	RatedMinCurrent int    `json:"rated_min_current"` //
	RatedMaxCurrent int    `json:"rated_max_current"` //
	HasCharingPrem  bool   `json:"has_charing_prem"`  // 是否拥有启/停充电权限
	SettingCurrent  int    `json:"setting_current"`   // 当前设置电流
	WorkMode        int    `json:"work_mode"`         // 工作模式, 1 即插即充， 其他事授权控制充电
	Alias           string `json:"alias"`             // 设备别名
}

type EvseStartReq struct {
	SN                string `json:"sn"`                  // 是	桩号
	ChargingType      int    `json:"charging_type"`       // 是	充电类型(0充电 1有序 2预约)
	StartChargingTime string `json:"start_charging_time"` // 否	预约充电开始时间. 格式: YYMMddhhmmss
	StopChargingTime  string `json:"stop_charging_time"`  // 否	预约充电结束时间. 格式: YYMMddhhmmss
	ChargingCurrent   int32  `json:"charging_current"`    // 否	有序充电电流. 精度A
	RepeatMode        string `json:"repeat_mode"`         // 预约充电重复模式。 空-单次预约，“day”-每天重复
}
type EvseStartReply struct {
	OrderId string `json:"order_id"` //	订单ID
}

type EvseStopReq struct {
	SN      string `json:"sn"`       // 是	桩号
	OrderId string `json:"order_id"` //	订单ID
}

type EvseStopReply struct {
	Status int32 `json:"status"`
}

type ResetReq struct {
	SN string `json:"sn"` // 是	桩号
}
type ResetReply struct {
}
type OrdersReq struct {
	Page int `json:"page"` // 是	页码, 0 开始开始
	Size int `json:"size"` // 是	条数

	SN        string `json:"sn,omitempty"`
	BeginTime int64  `json:"begin_time,omitempty"` // 开始时间戳
	EndTime   int64  `json:"end_time,omitempty"`   // 结束时间戳
}
type OrdersReply struct {
	Total  int     `json:"total"`
	Orders []Order `json:"orders"`
}

type Order struct {
	ID                string `json:"id"`                  // 订单ID
	Sn                string `json:"sn"`                  // 设备编号
	StartChargingTime int64  `json:"start_charging_time"` // 开始充电时间
	StopChargingTime  int64  `json:"stop_charging_time"`  // 结束充电时间
	Elec              int    `json:"elec"`                // 充电电量, 单位 0.01kw/h
	Reason            string `json:"reason"`              // 结束理由
	StartType         int    `json:"start_type"`          // 0 小程序，1 刷卡
	Phone             string `json:"phone"`               // 启动手机号
}

type OTAUpgradeReq struct {
	SN string `json:"sn"` // 设备编号
}
type OTAUpgradeReply struct {
}

type LatestFirmwareVersionReq struct {
	SN string `json:"sn"` // 设备编号
}
type LatestFirmwareVersionReply struct {
	LatestFirmwareVersion int16  `json:"latest_firmware_version"` // 最新固件版本
	LatestFirmwareDesc    string `json:"latest_firmware_desc"`    // 最新固件版本描述
}

type SetCurrentReq struct {
	SN              string `json:"sn"` //	设备编号
	ChargingCurrent int    `json:"charging_current"`
}

type SetCurrentReply struct {
}
type BTOrder struct {
	RecordID         string  `json:"record_id"`                // 充电流水号
	AuthMode         uint8   `json:"auth_mode"`                // 充电开始时间
	StartTime        uint32  `json:"start_time"`               // 充电时长
	ChargeTime       uint32  `json:"charge_time"`              // 充电时长
	TotalElectricity float64 `json:"total_electricity,string"` // 本次充电电量
	StopReason       uint8   `json:"stop_reason"`              // 充电停止原因
	FaultCode        uint8   `json:"fault_code"`               // 故障码
}

type SyncBTOrderReq struct {
	SN       string    `json:"sn"` // 设备编号
	BTOrders []BTOrder `json:"bt_orders"`
}
type SyncBTOrderReply struct {
}

type QuestionAndAnswer struct {
	Q string `json:"q"`
	A string `json:"a"`
}

type QuestionAndAnswerReq struct {
	Page int `json:"page"` // 0 开始
	Size int `json:"size"` // 每页
}
type QuestionAndAnswerReply struct {
	Total int                 `json:"total"`
	QAA   []QuestionAndAnswer `json:"qaa"`
}

type FeedbackReq struct {
	Content string `json:"content"`
	Email   string `json:"email"`
}
type FeedbackReply struct {
}

type AboutReq struct {
}
type AboutReply struct {
	Content string `json:"content"`
}
type UserEvsesReq struct {
}

type UserEvsesReply struct {
	EvseInfos []BindEvseInfo `json:"evse_infos"`
}

type EvseShareReq struct {
	SN string `json:"sn"` //	设备编号
}

type EvseShareReply struct {
	SN   string `json:"sn"` //	设备编号
	Auth string `json:"auth"`
}

type EvsePermReq struct {
	SN                 string `json:"sn"`                   //	设备编号
	MemberUid          string `json:"member_uid"`           // 	是		从用户uid
	EnableChargingPerm bool   `json:"enable_charging_perm"` // 	是		充电权限开关
}
type EvsePermReply struct {
}

type BindMembersReq struct {
	SN string `json:"sn"` //	设备编号
}
type BindMembersReply struct {
	BindMembers []BindMember `json:"bind_members"`
}
type BindMember struct {
	Id            string `json:"id"`             //	用户ID
	Phone         string `json:"phone"`          //	手机号
	IsMaster      bool   `json:"is_master"`      //	是否是主账号
	EnableCharing bool   `json:"enable_charing"` //	打开启动充电权限
}

type SetEvseDefaultReq struct {
	SN string `json:"sn"` // 设备编号
}
type SetEvseDefaultReply struct {
}
