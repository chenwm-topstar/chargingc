package slices

import (
	"reflect"
)

// Filter removes the items failing predict function in a slice.
// For performance consideration, it's a in-place op to modify the slice.
// The return `newSize` is the filtered size of slice,
// and should be applied like slice[:newSize]
func Filter(slices interface{}, predictFn func(i int) bool) (newSize int) {
	v := reflect.ValueOf(slices)
	if v.Kind() != reflect.Slice {
		return 0
	}
	var idx = 0
	currentSize := v.Len()
	for i := 0; i < currentSize; i++ {
		if predictFn(i) {
			v.Index(idx).Set(v.Index(i))
			idx++
		}
	}
	return idx
}

// FilterStrings remove given strings from a string slice.
// Be ware that the original slice is modifed in place.
func FilterStrings(slice []string, needles ...string) []string {
	var idx = 0
	set := NewStringSet(needles)
	for i := 0; i < len(slice); i++ {
		if !set.Contains(slice[i]) {
			slice[idx] = slice[i]
			idx++
		}
	}
	return slice[:idx]
}
