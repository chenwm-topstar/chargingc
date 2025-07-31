package template

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"

	"github.com/astaxie/beego"
)

func init() {

	beego.AddFuncMap("json_encode", func(v interface{}) string {
		b, _ := json.Marshal(v)
		return string(b)
	})

	beego.AddFuncMap("unescaped", func(x string) string {
		return string(template.HTML(x))
	})

	beego.AddFuncMap("addone", func(i int) int {
		return int(i + 1)
	})

	beego.AddFuncMap("String", String)
}

func String(s interface{}, def string, i ...int) string {
	switch s.(type) {
	case int, int8, int32, int64, uint32, uint, uint8, uint16, int16, uint64:
		ss := fmt.Sprintf("%d", s)
		i, err := strconv.ParseInt(ss, 10, 64)
		if err != nil || i == 0 {
			return def
		}
		return fmt.Sprintf("%d", i)
	case string:
		str := s.(string)
		if str == "" {
			return def
		}
		return str
	case float64:
		number := s.(float64)
		if number == 0 {
			return def
		}
		f := 2
		if len(i) > 0 {
			f = i[0]
		}
		return fmt.Sprintf(fmt.Sprintf("%%.%df", f), number)

	}
	return def
}
