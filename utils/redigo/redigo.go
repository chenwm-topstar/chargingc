package redigo

import (
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	pool *redis.Pool
)

func Init() {
	ManualInit(
		rdsAddress(),
		rdsPasswd(),
		rdsDb(),
		rdsMaxIdleConns(),
		rdsMaxActiveConns(),
	)
}

func ManualInit(addr, password string, db int, maxIdle, maxActive int) {
	pool = &redis.Pool{
		MaxActive:   maxIdle,
		MaxIdle:     maxActive,
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			options := []redis.DialOption{
				redis.DialDatabase(db),
				redis.DialConnectTimeout(5 * time.Second),
			}
			if password != "" {
				options = append(options, redis.DialPassword(password))
			}
			return redis.Dial("tcp", addr, options...)
		},
		Wait: true,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

// GetRedis 从redis连接池中获取一个连接
func GetRedis() redis.Conn {
	return pool.Get()
}

func GetPool() *redis.Pool {
	return pool
}

// Do 执行一个redis命令
func Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("missing required arguments")
	}
	c := GetRedis()
	reply, err = c.Do(commandName, args...)
	c.Close()
	return
}
