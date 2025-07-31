package driver

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"

	"cchome-admin-topstar/utils/access/codec"
	"cchome-admin-topstar/utils/access/codec/byteconv"
)

// ///////////////////////////////////////////////////////////////////////////////////
// 常用错误信息
// ///////////////////////////////////////////////////////////////////////////////////
var (
	// ErrBlock 设备/会话已经屏蔽
	ErrBlock = errors.New("error block")

	// ErrIgnore 忽略报文处理
	ErrIgnore = errors.New("error ignore")

	// ErrKick 断开socket
	ErrKick = errors.New("error kick")

	// ErrAlreadySession 重复的连接
	ErrAlreadySession = errors.New("already session")
)

/////////////////////////////////////////////////////////////////////////////////////
// ac常用基本类型
/////////////////////////////////////////////////////////////////////////////////////

type Bytes []byte

func (b Bytes) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte3 [3]byte

func (b Byte3) String() string {
	return byteconv.Byte2String(b[:])
}

type VByte3 [3]byte

func (b VByte3) String() string {
	return fmt.Sprintf("%d.%d.%d", b[0], b[1], b[2])
}

type Byte4 [4]byte

func (b Byte4) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte5 [5]byte

func (b Byte5) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte6 [6]byte

func (b Byte6) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte8 [8]byte

func (b Byte8) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte10 [10]byte

func (b Byte10) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte11 [11]byte

func (b Byte11) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte12 [12]byte

func (b Byte12) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte13 [13]byte

func (b Byte13) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte16 [16]byte

func (b Byte16) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte17 [17]byte

func (b Byte17) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte18 [18]byte

func (b Byte18) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte20 [20]byte

func (b Byte20) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte32 [32]byte

func (b Byte32) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte48 [48]byte

func (b Byte48) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte56 [56]byte

func (b Byte56) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte64 [64]byte

func (b Byte64) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte96 [96]byte

func (b Byte96) String() string {
	return byteconv.Byte2String(b[:])
}

type Byte192 [192]byte

func (b Byte192) String() string {
	return byteconv.Byte2String(b[:])
}

type Bytetime [8]byte

func NewBytetime(t time.Time, isNotBCD ...bool) (ret Bytetime) {
	if len(isNotBCD) > 0 && isNotBCD[0] {
		codec.GetEndian().PutUint16(ret[:2], uint16(t.Year()))
		ret[2] = uint8(t.Month())
		ret[3] = uint8(t.Day())
		ret[4] = uint8(t.Hour())
		ret[5] = uint8(t.Minute())
		ret[6] = uint8(t.Second())
	} else {
		tmp, _ := strconv.ParseUint(fmt.Sprintf("%04d%02d%02d%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()), 0, 64)
		_byt := byteconv.DecToBcdBig(tmp)
		// fmt.Printf("----->[%x][%x]\r\n", _byt, _byt[len(_byt)-7:])
		copy(ret[:7], _byt[len(_byt)-7:])
	}
	ret[7] = 0xff

	return
}

func (b Bytetime) String() string {
	return fmt.Sprintf("%x", b[:7])
}

func (b Bytetime) Unix(def ...time.Time) int64 {
	timeStr := fmt.Sprintf("%x", b[:7])
	if timeStr == "00000000000000" && len(def) == 1 { // 异常情况下，上传的结束时间可能为0
		return def[0].Unix()
	}
	t, _ := time.ParseInLocation("20060102150405", timeStr, time.Local)
	return t.Unix()
}

type BCDByte1 uint8

func (b BCDByte1) String() string {
	return fmt.Sprintf("%c", b)
}

type BCDByte2 [2]byte

func (b BCDByte2) String() string {
	return fmt.Sprintf("%x", b[:])
}

type BCDByte7 [7]byte

func (b BCDByte7) String() string {
	return fmt.Sprintf("%x", b[:])
}

type BCDByte8 [8]byte

func (b BCDByte8) String() string {
	// idx := 7
	// for idx = 7; idx >= 0; idx-- {
	// 	if b[idx] != 0 {
	// 		break
	// 	}
	// }
	// return string(byteconv.BCDToByte(b[:idx+1]))
	return fmt.Sprintf("%x", b[:])
}

type BCDByte10 [10]byte

func (b BCDByte10) String() string {
	return fmt.Sprintf("%x", b[:])
}

type BCDByte16 [16]byte

func (b BCDByte16) String() string {
	// return string(byteconv.BCDToByte(b[:]))
	return fmt.Sprintf("%x", b[:])
}

type CP56Time [7]byte

func (b CP56Time) String() string {
	mill := binary.LittleEndian.Uint16(b[0:2])
	return fmt.Sprintf("20%02d-%02d-%02d %02d:%02d:%02d", b[6]&0x7f, b[5]&0x0f, b[4]&0x1f, b[3]&0x1f, b[2]&0x3f, int32(mill/1000))
}

func (b CP56Time) Time() time.Time {
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", b.String(), time.Local)
	return t
}

// NewCP56Time 根据传入时间，创建一个cp56时间
func NewCP56Time(t time.Time) (ret CP56Time) {
	ret[6] = uint8(t.Year() % 100)
	ret[5] = uint8(t.Month())
	ret[4] = uint8(t.Day())
	ret[3] = uint8(t.Hour())
	ret[2] = uint8(t.Minute())
	binary.LittleEndian.PutUint16(ret[:2], uint16(t.Second()*1000))
	return
}
