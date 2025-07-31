package lib

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

type DataTablesRequest struct {
	Order  string                 `json:"order"`
	Sort   string                 `json:"sort"`
	Limit  int                    `json:"limit"`
	Search string                 `json:"search"`
	Offset int                    `json:"offset"`
	OP     *KindMapStringJSON     `json:"op"`
	Filter *KindMapInterfaceJSON  `json:"filter,omitempty"`
	Where  map[string]interface{} `json:"where"`
}

func NewDataTableRequest() *DataTablesRequest {
	f := &DataTablesRequest{}
	f.Where = make(map[string]interface{})
	return f
}

type KindMapStringJSON map[string]string

func (v *KindMapStringJSON) UnmarshalJSON(data []byte) error {
	var jsonMap map[string]string
	if len(data) <= 2 {
		return nil
	}
	s := strings.Replace(strings.Trim(string(data), `"`), `\"`, `"`, -1)
	err := json.Unmarshal([]byte(s), &jsonMap)
	if err != nil {
		return err
	}
	*v = KindMapStringJSON(jsonMap)
	return nil
}

type KindMapInterfaceJSON map[string]interface{}

func (v *KindMapInterfaceJSON) UnmarshalJSON(data []byte) error {
	var jsonMap map[string]interface{}
	if len(data) <= 2 {
		return nil
	}
	var err error
	if string(data)[0] == '"' {
		s := strings.Replace(strings.Trim(string(data), `"`), `\"`, `"`, -1)
		err = json.Unmarshal([]byte(s), &jsonMap)
	} else {
		err = json.Unmarshal(data, &jsonMap)
	}

	if err != nil {
		return err
	}
	*v = KindMapInterfaceJSON(jsonMap)
	return nil
}

func (v *KindMapInterfaceJSON) Scan(input interface{}) error {
	return v.UnmarshalJSON(input.([]byte))
}

func (v KindMapInterfaceJSON) Value() (driver.Value, error) {
	return json.Marshal(v)
}
