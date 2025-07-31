package redigo

import (
	"context"
	"fmt"
	"time"

	"cchome-admin-topstar/utils/sync/errgroup.v2"

	"github.com/garyburd/redigo/redis"
)

func Lock(key string, timeout int, block ...bool) (func() error, error) {
	val, err := TryLock(key, timeout, block...)
	if err != nil {
		return nil, err
	}
	return func() error {
		return Unlock(key, val)
	}, nil
}

// TryLock
// key 键名
// timeout 获取锁多长时间超时，单位秒
// block 是否堵塞等待，默认为false：获取不了锁就返回错误, true：则会堵塞等待，直到获取到锁或超时
// 参考 https://huoding.com/2015/09/14/463
func TryLock(key string, timeout int, block ...bool) (val int64, err error) {
	conn := GetRedis()
	defer conn.Close()

	var ret string
	if len(block) == 1 && block[0] {
		g := errgroup.WithTimeout(context.TODO(), time.Duration(timeout)*time.Second)
		g.Go(func(ctx context.Context) error {
			t := time.NewTicker(500 * time.Millisecond)
			firstExec := make(chan struct{}, 1)
			defer func() {
				close(firstExec)
				t.Stop()
			}()
			firstExec <- struct{}{}

			for {
				select {
				case <-ctx.Done():
					return fmt.Errorf("timeout")
				case <-firstExec:
				case <-t.C:
				}
				val = time.Now().UnixNano()
				ret, err = redis.String(conn.Do("set", key, val, "ex", timeout, "nx"))
				if err != nil && err != redis.ErrNil {
					return err
				}
				if ret == "OK" {
					return nil
				}
			}
		})
		err = g.Wait()
		return
	}

	val = time.Now().UnixNano()
	if ret, err = redis.String(conn.Do("set", key, val, "ex", timeout, "nx")); err != nil {
		if err == redis.ErrNil {
			err = fmt.Errorf("get lock return is nil")
		}
		return
	}
	if ret != "OK" {
		return val, fmt.Errorf("get lock fail. %s", ret)
	}

	return
}

// Unlock 解锁
func Unlock(key string, val int64) error {
	rd := GetRedis()
	defer rd.Close()

	v, err := redis.Int64(rd.Do("get", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil
		}
		return err
	}

	if v == val {
		rd.Do("del", key)
	} else {
		return fmt.Errorf("unlock fail, version not match")
	}
	return nil
}
