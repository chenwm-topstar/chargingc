package driver

import (
	"context"
)

const (
	// InjectProtocol 注入协议
	InjectProtocol = "protocol"

	// InjectDriverName 注入驱动名
	InjectDriverName = "drivername"

	// InjectPointDriver 转发端driver
	InjectPointDriver = "point_driver"

	// InjectConnectCallback 初始化连接回调函数 tcp客户端使用
	InjectConnectCallback = "connect_callback"

	// InjectDisconnectCallback 断开回调函数 tcp服务器使用
	InjectDisconnectCallback = "disconnect_callback"

	// InjectReceiveCallback tcp接收报文数据回调函数
	InjectReceiveCallback = "receive_callback"

	DebugSessions     = "debug_sessions"
	DebugHostSessions = "debug_host_sessions"
)

// Inject 注入
type Inject interface {
	Inject(name string, p interface{}) Inject
}

// Driver ac接入驱动
type Driver interface {
	// Run 运行驱动
	Run() (err error)

	// Stop 停止
	Stop()

	// CheckOnline 检查是否在线
	CheckOnline(mark string) (ok bool)

	// Disconnector 断开连接
	Disconnector(mark, reason string)

	// Send 发送消息
	Send(msg Msg) (err error)

	// TraficSize 流量统计
	TraficSize(mark string) (recv, send int32, err error)

	Debug(t string)
}

// Msg 消息(普通消息/tcp消息/通道消息)
type Msg interface {
	// GetMark 设备标记
	GetMark() string
	// GetMsg 获取消息内容
	GetMsg() interface{} // 返回 []byte 或者 [][]byte
	// GetSource 获取消息源数据
	GetSource() interface{}
}

// MqttMsg mqtt消息
type MqttMsg interface {
	Msg

	// GetQos qos
	GetQos() byte // qos

	// GetRetained  是否遗留信息
	GetRetained() bool // retained

	// GetTopic topic
	GetTopic() string // topic
}

// Protocol 协议接口
type Protocol interface {
	Translate(ctx context.Context) (tos []Msg, rets []Msg, err error)
}
