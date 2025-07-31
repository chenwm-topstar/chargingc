package protocol

import (
	"errors"

	"github.com/chenwm-topstar/chargingc/cchome-admin/transac/itransac"
)

// ErrPayloadNotSupport 不支持上/下行转换
var ErrPayloadNotSupport = errors.New("payload not support")

// IUpPayload  上行数据转换
type IUpPayload interface {
	// ToPlatformPayload 转换成平台的Payload
	ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error)
}

// IDownPayload 下行数据转换
type IDownPayload interface {
	// ToBicyclePayload 转换成单车的Payload
	ToDevicePayload(ctx *itransac.Ctx) error
}

// IPayload 转换成上行数据
type IPayload interface {
	IUpPayload
	IDownPayload
}
