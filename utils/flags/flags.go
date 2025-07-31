package flags

import (
	"context"
	"time"

	"cchome-admin-topstar/utils/service/exp"

	"github.com/spf13/pflag"
)

func String(key, defaultValue, usage string) func() string {
	pflag.String(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	v.BindPFlag(key, pflag.Lookup(key))
	return func() string {
		return v.GetString(key)
	}
}

func StringExp(key, defaultValue, usage string) func(ctx context.Context) string {
	p := String(key, defaultValue, usage)
	return func(ctx context.Context) string {
		return exp.GetExperimentString(ctx, key, p())
	}
}
func StringRequired(key, usage string) func() string {
	requiredKey = append(requiredKey, key)
	return String(key, "", usage)
}

func Bool(key string, defaultValue bool, usage string) func() bool {
	pflag.Bool(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	v.BindPFlag(key, pflag.Lookup(key))
	return func() bool {
		return v.GetBool(key)
	}
}

func BoolExp(key string, defaultValue bool, usage string) func(ctx context.Context) bool {
	p := Bool(key, defaultValue, usage)
	return func(ctx context.Context) bool {
		return exp.GetExperimentBool(ctx, key, p())
	}
}

func BoolRequired(key, usage string) func() bool {
	requiredKey = append(requiredKey, key)
	return Bool(key, false, usage)
}

func Int(key string, defaultValue int, usage string) func() int {
	pflag.Int(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	v.BindPFlag(key, pflag.Lookup(key))
	return func() int {
		return v.GetInt(key)
	}
}

func IntExp(key string, defaultValue int, usage string) func(ctx context.Context) int {
	p := Int(key, defaultValue, usage)
	return func(ctx context.Context) int {
		return exp.GetExperimentInt(ctx, key, p())
	}
}

func IntRequired(key, usage string) func() int {
	requiredKey = append(requiredKey, key)
	return Int(key, 0, usage)
}

// 通过环境变量获取有点问题
func Slice(key string, defaultValue []string, usage string) func() []string {
	pflag.StringSlice(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	v.BindPFlag(key, pflag.Lookup(key))
	return func() []string {
		return v.GetStringSlice(key)
	}
}

func SliceRequired(key, usage string) func() []string {
	requiredKey = append(requiredKey, key)
	return Slice(key, nil, usage)
}

func Float64(key string, defaultValue float64, usage string) func() float64 {
	pflag.Float64(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	v.BindPFlag(key, pflag.Lookup(key))
	return func() float64 {
		return v.GetFloat64(key)
	}
}

func Float64Exp(key string, defaultValue float64, usage string) func(ctx context.Context) float64 {
	p := Float64(key, defaultValue, usage)
	return func(ctx context.Context) float64 {
		return exp.GetExperimentFloat64(ctx, key, p())
	}
}

func Float64Required(key, usage string) func() float64 {
	requiredKey = append(requiredKey, key)
	return Float64(key, 0, usage)
}

func Duration(key string, defaultValue time.Duration, usage string) func() time.Duration {
	pflag.Duration(key, defaultValue, usage)
	v.SetDefault(key, defaultValue)
	v.BindPFlag(key, pflag.Lookup(key))
	return func() time.Duration {
		return v.GetDuration(key)
	}
}

func DurationRequired(key, usage string) func() time.Duration {
	requiredKey = append(requiredKey, key)
	return Duration(key, 0, usage)
}
