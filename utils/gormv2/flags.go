package gormv2

import (
	"time"

	"cchome-admin-topstar/utils/flags"
)

var (
	mysqlHost = flags.String("mysql_host", "192.168.0.54", "mysql server address. default: mysql")
	mysqlPort = flags.Int("mysql_port", 3306, "mysql server port. default: 3306")
	//mysqlUser            = flags.StringRequired("mysql_user", "mysql user.")
	mysqlUser = flags.String("mysql_user", "root", "mysql user.")
	//mysqlPasswd          = flags.StringRequired("mysql_passwd", "mysql passwd.")
	mysqlPasswd          = flags.String("mysql_passwd", "cchome@admin", "mysql passwd.")
	mysqlMaxIdleConns    = flags.Int("mysql_max_idle_conns", 10, "mysql max idle conns")
	mysqlMaxOpenConns    = flags.Int("mysql_max_open_conns", 100, "mysql max open conns")
	mysqlConnMaxLifetime = flags.Duration("mysql_conn_max_lifetime", 300*time.Second, "mysql max open conns")
)
