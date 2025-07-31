package driver

import (
	"context"

	"github.com/sirupsen/logrus"
)

/////////////////////////////////////////////////////////////////////////////////////
// 自定义上下文信息
/////////////////////////////////////////////////////////////////////////////////////

// Ctx ac上下文
type Ctx struct {
	Raw        interface{}            // 报文字节
	Mark       string                 // 唯一标记，evseid/sn
	Data       map[string]interface{} // data
	Log        *logrus.Entry          // 日志对象
	AfterFuncs []func()               // 回调
	DriverName string                 // 驱动名称
}

// Clone 复制一个上下文信息
// Clone 不能在goroution里面clone，容易并发读写map，导致panic
func (ctx *Ctx) Clone() (ret *Ctx) {
	ret = &Ctx{
		Raw:        ctx.Raw,
		Mark:       ctx.Mark,
		Log:        ctx.Log.Dup(),
		Data:       make(map[string]interface{}),
		DriverName: ctx.DriverName,
	}
	for k, v := range ctx.Data {
		ret.Data[k] = v
	}
	return
}

// NewACCtx ...
func NewACCtx(mark, drivername string, raw interface{}) *Ctx {
	return &Ctx{
		Mark:       mark,
		DriverName: drivername,
		Raw:        raw,
		Log:        logrus.WithField("mark", mark),
		Data:       make(map[string]interface{}),
		AfterFuncs: make([]func(), 0),
	}
}

// NewACContext accontext
func NewACContext(ctx context.Context, acctx *Ctx) context.Context {
	return context.WithValue(ctx, "acctx", acctx)
}

// GetACCtxWithContext ...
func GetACCtxWithContext(ctx context.Context) *Ctx {
	if _acctx := ctx.Value("acctx"); _acctx != nil {
		return _acctx.(*Ctx)
	}
	return nil
}
