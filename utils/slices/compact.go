package slices

import (
	"reflect"
)

// Compact moves all non-nil elements in a slice to leftmost, returning
// the number of non-nil elements.
// If the element in slice is of type int, float, string, zero value will be regarded
// as nil.
// NOTE: This function performs in place modification. It has side-effect on slice.
// A common usage:
// ```
//   myslice := []string{"foo", nil, "bar"}
//   n := routine.Compact(myslice)
//   myslice = myslice[:n]
// ```
func Compact(slice interface{}) int {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return 0
	}
	start := 0
	for i := 0; i < v.Len(); i++ {
		obj := v.Index(i)
		if obj.IsZero() {
			continue
		}
		v.Index(start).Set(obj)
		start++
	}
	return start
}
