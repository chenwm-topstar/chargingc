package driver

/////////////////////////////////////////////////////////////////////////////////////
// 转发消息类型
/////////////////////////////////////////////////////////////////////////////////////

// DefMsg 转发基本消息
// todo: 可以加个延时属性
type DefMsg struct {
	Mark   string
	Raw    interface{} // 要发送的消息内容(是消息数据源处理后的内容)
	Source interface{} // 消息数据源头
}

// NewMsg 创建一个基本消息
// vs[0] 打包后的消息体
// vs[1] 打包前的消息体
func NewMsg(mark string, vs ...interface{}) Msg {
	l := len(vs)
	switch l {
	case 1:
		return &DefMsg{Mark: mark, Raw: vs[0]}
	case 2:
		return &DefMsg{Mark: mark, Raw: vs[0], Source: vs[1]}
	}
	return nil
}

// GetMark 转发标记
func (m *DefMsg) GetMark() string {
	return m.Mark
}

// GetMsg 获取消息内容
func (m *DefMsg) GetMsg() interface{} {
	return m.Raw
}

// GetSource 消息数据源头
func (m *DefMsg) GetSource() interface{} {
	return m.Source
}

// DefaultMqttMsg mqtt消息
type DefaultMqttMsg struct {
	*DefMsg
	Qos      byte   // qos
	Retained bool   // retained
	Topic    string // topic
}

// NewMqttMsg 创建一个mqtt消息
func NewMqttMsg(msg Msg, qos byte, retained bool, topic string) Msg {
	return &DefaultMqttMsg{
		DefMsg: &DefMsg{
			Mark:   msg.GetMark(),
			Raw:    msg.GetMsg(),
			Source: msg.GetSource(),
		},
		Qos:      qos,
		Retained: retained,
		Topic:    topic,
	}
}

// GetQos qos
func (mm *DefaultMqttMsg) GetQos() byte {
	return mm.Qos
}

// GetRetained  是否遗留信息
func (mm *DefaultMqttMsg) GetRetained() bool {
	return mm.Retained
}

// GetTopic topic
func (mm *DefaultMqttMsg) GetTopic() string {
	return mm.Topic
}
