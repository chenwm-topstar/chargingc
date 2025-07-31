package uuid

import (
	"fmt"
	"strconv"
	"strings"

	"cchome-admin-topstar/utils/lg"

	uuid "github.com/lnyyj/go.uuid"
	"github.com/sony/sonyflake"
)

type ID uint64

func (id ID) String() string {
	return fmt.Sprintf("%d", id)
}

func (id ID) Uint64() uint64 {
	return uint64(id)
}

var sf *sonyflake.Sonyflake
var ids chan ID

func init() {
	var st sonyflake.Settings
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("sonyflake not created")
	}
	ids = make(chan ID, 100)
	go func() {
		for {
			id, err := sf.NextID()
			if err != nil {
				lg.Warnf("sonyflake nextid error: %s", err.Error())
				continue
			}
			ids <- ID(id)
		}
	}()
}

func GetID() ID {
	return <-ids
}

func ParseID(v string) (ID, error) {
	id, err := strconv.ParseUint(v, 10, 64)
	return ID(id), err
}

func NewUuidV4() (string, error) {
	uuid4, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(uuid4.String(), "-", ""), nil
}

func NewUuidV1() (string, error) {
	uuid1, err := uuid.NewV1()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(uuid1.String(), "-", ""), nil
}
