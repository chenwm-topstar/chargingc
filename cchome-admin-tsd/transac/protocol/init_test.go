package protocol

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func Test_GetConnectState(t *testing.T) {
	fmt.Printf("---------->%d\r\n", getConnectState(150))
}

func getWeek(v uint8) (ret []int) {
	for i := 0; i < 8; i++ {
		if ((v >> i) & 1) == 1 {
			ret = append(ret, i+1)
		}
	}
	return
}

func TestGetWeek(t *testing.T) {
	fmt.Printf("---------->%d\r\n", getWeek(112))
}

func TestUnmarshal(t *testing.T) {
	// 6818de0939333331323131363600000000000000003130303131
	buf, _ := hex.DecodeString("6818b80739333331323131363600000000000000003130303151")
	apdu := &APDU{}
	err := apdu.Unmarshal(buf)
	t.Logf("------>[%+v][%+v]\r\n", err, apdu.Payload)
}
