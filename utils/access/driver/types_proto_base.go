package driver

// ProtoBase 基础协议翻译
// 协议翻译对象，需要继承这个对象
type ProtoBase struct {
	drivers map[string]Driver
}

// Inject 注入
func (pb *ProtoBase) Inject(name string, p interface{}) Inject {
	if pb.drivers == nil {
		pb.drivers = make(map[string]Driver)
	}
	if f, ok := p.(Driver); ok {
		pb.drivers[name] = f
	}
	return pb
}

// GetDrivers 获取多端网络驱动
// 用于协议中无法映射命令的转发
func (pb *ProtoBase) GetDrivers() map[string]Driver {
	return pb.drivers
}
