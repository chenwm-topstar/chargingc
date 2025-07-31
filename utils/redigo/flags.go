package redigo

import "cchome-admin-topstar/utils/flags"

var (
	rdsAddress        = flags.String("redis_address", "192.168.0.54:6379", "redis server address. default: redis:6379")
	rdsPasswd         = flags.String("redis_passwd", "", "redis server passwd")
	rdsDb             = flags.Int("redis_db", 0, "redis select db. default: 0")
	rdsMaxIdleConns   = flags.Int("redis_max_idle_conns", 10, "redis max idle session. default: 10")
	rdsMaxActiveConns = flags.Int("redis_max_active_conns", 200, "redis max active session. default: 200")
)
