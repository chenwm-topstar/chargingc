package exp

import (
	"context"
	"fmt"
	"strconv"

	"cchome-admin-topstar/utils/service/servicecontext"
)

func GetExperimentString(ctx context.Context, key, defaultValue string) string {
	sc := servicecontext.FromContext(ctx)
	val, ok := sc.Variable(key)
	if !ok {
		return defaultValue
	}
	return val
}

func GetExperimentInt(ctx context.Context, key string, defaultValue int) int {
	sc := servicecontext.FromContext(ctx)
	val, ok := sc.Variable(key)
	if !ok {
		return defaultValue
	}
	valInt, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return valInt
}

func GetExperimentFloat64(ctx context.Context, key string, defaultValue float64) float64 {
	sc := servicecontext.FromContext(ctx)
	val, ok := sc.Variable(key)
	if !ok {
		return defaultValue
	}
	var valFloat float64
	if _, err := fmt.Sscanf(val, "%f", &valFloat); err != nil {
		return defaultValue
	}

	return valFloat
}

func GetExperimentBool(ctx context.Context, key string, defaultValue bool) bool {
	sc := servicecontext.FromContext(ctx)
	val, ok := sc.Variable(key)
	if !ok {
		return defaultValue
	}
	var valBool bool
	if _, err := fmt.Sscanf(val, "%v", &valBool); err != nil {
		return defaultValue
	}

	return valBool
}
