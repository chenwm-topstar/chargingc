package log

import (
	"github.com/astaxie/beego/context"
	"github.com/sirupsen/logrus"
)

func FromBeegoContext(ctx *context.Context) *logrus.Entry {
	entry := ctx.Input.GetData("logEntry")
	if entry == nil {
		_entry := logrus.WithFields(logrus.Fields{
			"module":    "admin",
			"requestID": ctx.Input.GetData("requestID"),
		})
		ctx.Input.SetData("logEntry", _entry)
		return _entry
	}
	return entry.(*logrus.Entry)
}

func NewFromBeegoContext(ctx *context.Context, module string) *logrus.Entry {
	oldLogEntry := FromBeegoContext(ctx)
	entry := logrus.WithFields(logrus.Fields{
		"module":    module,
		"requestID": oldLogEntry.Data["requestID"],
	})
	ctx.Input.SetData("logEntry", entry)
	return entry
}
