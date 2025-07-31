package driver

// NDBase 基础网络驱动
// 实现网络驱动时，需要继续这个对象
type NDBase struct {
	driverName  string      // 驱动名称
	pointDriver Driver      // 转发端驱动
	translate   interface{} // 翻译对象
}

// Inject 注入相关信息
func (b *NDBase) Inject(fname string, f interface{}) Inject {
	switch fname {
	case InjectProtocol:
		b.translate = f
	case InjectDriverName:
		b.driverName = f.(string)
	case InjectPointDriver:
		b.pointDriver = f.(Driver)
	}
	return b
}

// GetDriverName 驱动名称
func (b *NDBase) GetDriverName() string {
	return b.driverName
}

// GetPointDriver 获取转发端驱动
func (b *NDBase) GetPointDriver() Driver {
	return b.pointDriver
}

// GetTranslate 获取协议翻译对象
func (b *NDBase) GetTranslate() interface{} {
	return b.translate
}
