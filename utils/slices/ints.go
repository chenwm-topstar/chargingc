package slices

import (
	"math"
	"sort"
	"strings"
)

func ContainInt(slice []int, s int) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func ContainInt32(slice []int32, s int32) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func ContainInt64(slice []int64, s int64) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// TODO(yuheng): This function is ambigious on the function naming, since
// it deletes only ONE element from the slice, instead of all matching elements.
func DeleteInt32(slice []int32, s int32) []int32 {
	for i, item := range slice {
		if item == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func DistinctInt32(slice []int32) (ret []int32) {
	return DedupInt32s(slice)
}

func JoinInt32(slice []int32, sep string) string {
	return strings.Join(Int32ToStrSlice(slice), sep)
}

func SortInt32(slice []int32) []int32 {
	sort.Slice(slice, func(i, j int) bool { return slice[i] < slice[j] })
	return slice
}

// BlurSplit 模糊查分一个int数值
func BlurSplit(v int64, n int) (ret []int64) {
	if n <= 1 {
		return []int64{v}
	}
	remain := v
	val := int64(math.Ceil(float64(v) / float64(n))) // 先采用四舍五入的方式， 后面可以新增一个去尾法
	if val == 0 && v > 1 {
		val = 1
	}
	for i := 0; i < n; i++ {
		if i == n-1 {
			ret = append(ret, int64(remain))
			continue
		}
		ret = append(ret, val)
		remain -= val
	}
	return
}
