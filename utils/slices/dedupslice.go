package slices

import "reflect"

// DedupSlice remove duplicated elements in slice.
// User needs to provide a "key" function to determine key of each elements,
// and receives a length "n" to sorten the slice.
// The usage is similar to Compact() function:
// ```
//   myslice := []Object{a, b, a}
//   n := routine.Dedup(myslice, func(i int) { return myslice[i].Key })
//   myslice = myslice[:n]
// NOTE: This function performs in place modification. It has side-effect on slice.
// ```
func Dedup(slice interface{}, keyFn func(i int) string) int {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return 0
	}

	dup := NewStringSet()
	start := 0
	for i := 0; i < v.Len(); i++ {
		key := keyFn(i)
		if dup.Contains(key) {
			continue
		}
		dup.Add(key)
		if i == start {
			start++
			continue
		}
		v.Index(start).Set(v.Index(i))
		start++
	}
	return start
}
