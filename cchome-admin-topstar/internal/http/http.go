package http

import "encoding/json"

type IResp interface {
	GetData() interface{}
	GetCode() int
	GetMsg() string
}

type Resp struct {
	Code    int           `json:"code"`
	Msg     string        `json:"msg,omitempty"`
	KindMsg int           `json:"kind_msg"`
	Data    interface{}   `json:"data,omitempty"`
	Total   int           `json:"total,omitempty"`
	Rows    []interface{} `json:"rows"`
	Wait    int           `json:"wait,omitempty"`
}

func (r Resp) GetCode() int {
	return r.Code
}

func (r Resp) GetMsg() string {
	return r.Msg
}

func (r Resp) GetData() interface{} {
	return r.Data
}

func (r Resp) JsonEncode() ([]byte, error) {
	return json.Marshal(r)
}

func JsonDecode(j string) (IResp, error) {
	var r Resp
	err := json.Unmarshal([]byte(j), &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
