package protocol

import (
	"errors"

	"github.com/chenwm-topstar/chargingc/cchome-admin-topstar/transac/itransac"
)

var ErrPayloadNotSupport = errors.New("payload not support")

type IUpPayload interface {
	ToPlatformPayload(ctx *itransac.Ctx) (retapdu []byte, err error)
}

type IDownPayload interface {
	ToDevicePayload(ctx *itransac.Ctx) error
}

type IPayload interface {
	IUpPayload
	IDownPayload
}
