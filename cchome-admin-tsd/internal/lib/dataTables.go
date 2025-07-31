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

// func (f *DataTablesRequest) WithSearchFilter(wheres ...map[string]interface{}) (ret []*SearchFilter) {
// 	newSearchFilter := func(name, op string, val string) *SearchFilter {
// 		return &SearchFilter{
// 			FilterName: name,
// 			FilterOp:   op,
// 			FilterVal:  val,
// 		}
// 	}
// 	if f != nil && f.Filter != nil {
// 		for k, v := range *f.Filter {
// 			op, _ := (map[string]string(*f.OP))[k]
// 			ret = append(ret, newSearchFilter(k, op, v.(string)))
// 		}
// 	}
// 	for _, w := range wheres {
// 		if w != nil {
// 			for k, v := range w {
// 				ret = append(ret, newSearchFilter(k, "=", fmt.Sprintf("%v", v)))
// 			}
// 		}
// 	}
// 	return
// }

type KindMapStringJSON map[string]string

//
//func (v KindMapStringJSON) MarshalJSON() ([]byte, error) {
//	return json.Marshal(v)
//}

func (v *KindMapStringJSON) UnmarshalJSON(data []byte) error {
	//jsonMap := make(map[string]interface{})
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

//KindMapInterfaceJSON json格式的map对象
type KindMapInterfaceJSON map[string]interface{}

//
//func (v KindMapInterfaceJSON) MarshalJSON() ([]byte, error) {
//	return json.Marshal(v)
//}

func (v *KindMapInterfaceJSON) UnmarshalJSON(data []byte) error {
	//jsonMap := make(map[string]interface{})
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
	//fmt.Println("vvvvvvvvvvvvvvvvv", input, string(input.([]byte)))
	//return json.Unmarshal(input.([]byte), v)
	return v.UnmarshalJSON(input.([]byte))
}

func (v KindMapInterfaceJSON) Value() (driver.Value, error) {
	return json.Marshal(v)
}
