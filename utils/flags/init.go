package flags

import (
	"math/rand"
	"time"

	"cchome-admin-topstar/utils/lg"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// requiredKey 必须要配置的参数
	requiredKey []string

	// debug 内置debug模式
	debug *bool

	v = viper.New()
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func initFlags() {
	v.BindPFlags(pflag.CommandLine)
	v.AutomaticEnv()
	v.SetEnvPrefix("env")

	debug = pflag.Bool("debug", false, "Set true to enable debug mode")
}

// Parse has to called after main() before any application code.
func Parse() {
	initFlags()
	pflag.Parse()
	if *debug {
		lg.EnableDebug()
	}

	for _, k := range requiredKey {
		if isZero(v.Get(k)) {
			lg.Fatalf("Missing config:[%s]", k)
		}
	}
}

func isZero(i interface{}) bool {
	switch i.(type) {
	case bool:
		// It's trivial to check a bool, since it makes the flag no sense(always true).
		return !i.(bool)
	case string:
		return i.(string) == ""
	case time.Duration:
		return i.(time.Duration) == 0
	case float64:
		return i.(float64) == 0
	case int:
		return i.(int) == 0
	case []string:
		return len(i.([]string)) == 0
	case []interface{}:
		return len(i.([]interface{})) == 0
	default:
		return true
	}
}
